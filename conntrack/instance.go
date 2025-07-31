package conntrack

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/anacrolix/stm"
	"github.com/anacrolix/stm/stmutil"

	"github.com/anacrolix/missinggo/v2"
	"github.com/anacrolix/missinggo/v2/iter"
)

type reason = string

type Instance struct {
	maxEntries   *stm.Var[int]
	noMaxEntries *stm.Var[bool]
	Timeout      func(Entry) time.Duration

	// Occupied slots
	entries *stm.Var[any]

	// priority to entryHandleSet, ordered by priority ascending
	waitersByPriority *stm.Var[any] //Mappish
	waitersByReason   *stm.Var[any] //Mappish
	waitersByEntry    *stm.Var[any] //Mappish
	waiters           *stm.Var[any] // Settish
}

type (
	priority int
)

func NewInstance() *Instance {
	i := &Instance{
		// A quarter of the commonly quoted absolute max on a Linux system.
		maxEntries:   stm.NewVar[int](1 << 14),
		noMaxEntries: stm.NewVar[bool](false),
		Timeout: func(e Entry) time.Duration {
			// udp is the main offender, and the default is allegedly 30s.
			return 30 * time.Second
		},
		entries: stm.NewVar[any](stmutil.NewMap[any, any]()),
		waitersByPriority: stm.NewVar[any](stmutil.NewSortedMap[any, any](func(l, r any) bool {
			return l.(priority) > r.(priority)
		})),

		waitersByReason: stm.NewVar[any](stmutil.NewMap[any, any]()),
		waitersByEntry:  stm.NewVar[any](stmutil.NewMap[any, any]()),
		waiters:         stm.NewVar[any](stmutil.NewSet[any]()),
	}
	return i
}

func (i *Instance) SetNoMaxEntries() {
	stm.AtomicSet(i.noMaxEntries, true)
}

func (i *Instance) SetMaxEntries(max int) {
	stm.Atomically(stm.VoidOperation(func(tx *stm.Tx) {
		i.noMaxEntries.Set(tx, false)
		i.maxEntries.Set(tx, max)
	}))
}

func (i *Instance) remove(eh *EntryHandle) {
	stm.Atomically(func(tx *stm.Tx) interface{} {
		es, _ := deleteFromMapToSet(i.entries.Get(tx).(any), eh.e, eh)
		i.entries.Set(tx, es)
		return nil
	})
}

func deleteFromMapToSet(m any, mapKey, setElem interface{}) (any, bool) {
	mappish := m.(stmutil.Mappish[any, any])
	_s, ok := mappish.Get(mapKey)
	if !ok {
		return m, true
	}
	s := _s.(stmutil.Settish[any])
	s = s.Delete(setElem)
	if s.Len() == 0 {
		return mappish.Delete(mapKey), true
	}
	return mappish.Set(mapKey, s), false
}

func (i *Instance) deleteWaiter(eh *EntryHandle, tx *stm.Tx) {
	i.waiters.Set(tx, i.waiters.Get(tx).(stmutil.Settish[any]).Delete(eh))
	i.waitersByPriority.Set(tx, stmutil.GetLeft(deleteFromMapToSet(i.waitersByPriority.Get(tx).(any), eh.priority, eh)))
	i.waitersByReason.Set(tx, stmutil.GetLeft(deleteFromMapToSet(i.waitersByReason.Get(tx).(any), eh.reason, eh)))
	i.waitersByEntry.Set(tx, stmutil.GetLeft(deleteFromMapToSet(i.waitersByEntry.Get(tx).(any), eh.e, eh)))
}

func (i *Instance) addWaiter(eh *EntryHandle) {
	stm.Atomically(stm.VoidOperation(func(tx *stm.Tx) {
		i.waitersByPriority.Set(tx, addToMapToSet(i.waitersByPriority.Get(tx).(any), eh.priority, eh))
		i.waitersByReason.Set(tx, addToMapToSet(i.waitersByReason.Get(tx).(any), eh.reason, eh))
		i.waitersByEntry.Set(tx, addToMapToSet(i.waitersByEntry.Get(tx).(any), eh.e, eh))
		i.waiters.Set(tx, i.waiters.Get(tx).(stmutil.Settish[any]).Add(eh))
	}))
}

func addToMapToSet(m any, mapKey, setElem interface{}) any {
	mappish := m.(stmutil.Mappish[any, any])
	s, ok := mappish.Get(mapKey)
	if ok {
		s = s.(stmutil.Settish[any]).Add(setElem)
	} else {
		s = stmutil.NewSet[any]().Add(setElem)
	}
	return mappish.Set(mapKey, s)
}

func (i *Instance) WaitDefault(ctx context.Context, e Entry) *EntryHandle {
	return i.Wait(ctx, e, "", 0)
}

// Nil returns are due to context completion.
func (i *Instance) Wait(ctx context.Context, e Entry, reason string, p priority) (eh *EntryHandle) {
	eh = &EntryHandle{
		reason:   reason,
		e:        e,
		i:        i,
		priority: p,
		created:  time.Now(),
	}
	i.addWaiter(eh)
	ctxDone, cancel := stmutil.ContextDoneVar(ctx)
	defer cancel()
	success := stm.Atomically(func(tx *stm.Tx) interface{} {
		es := i.entries.Get(tx).(stmutil.Mappish[any, any])
		if s, ok := es.Get(e); ok {
			i.entries.Set(tx, es.Set(e, s.(stmutil.Settish[any]).Add(eh)))
			return true
		}
		haveRoom := i.noMaxEntries.Get(tx) || es.Len() < i.maxEntries.Get(tx)
		topPrio, ok := iter.First(i.waitersByPriority.Get(tx).(iter.Iterable).Iter)
		if !ok {
			panic("y u no waiting")
		}
		if haveRoom && p == topPrio {
			i.entries.Set(tx, addToMapToSet(es, e, eh))
			return true
		}
		if ctxDone.Get(tx) {
			return false
		}
		tx.Retry()
		panic("unreachable")
	}).(bool)
	stm.Atomically(stm.VoidOperation(func(tx *stm.Tx) {
		i.deleteWaiter(eh, tx)
	}))
	if !success {
		eh = nil
	}
	return
}

func (i *Instance) Allow(tx *stm.Tx, e Entry, reason string, p priority) *EntryHandle {
	eh := &EntryHandle{
		reason:   reason,
		e:        e,
		i:        i,
		priority: p,
		created:  time.Now(),
	}
	es := i.entries.Get(tx).(stmutil.Mappish[any, any])
	if s, ok := es.Get(e); ok {
		i.entries.Set(tx, es.Set(e, s.(stmutil.Settish[any]).Add(eh)))
		return eh
	}
	haveRoom := i.noMaxEntries.Get(tx) || es.Len() < i.maxEntries.Get(tx)
	topPrio, ok := iter.First(i.waitersByPriority.Get(tx).(iter.Iterable).Iter)
	if haveRoom && (!ok || p == topPrio) {
		i.entries.Set(tx, addToMapToSet(es, e, eh))
		return eh
	}
	return nil
}

func parseHostPort(hostport string) (ret struct {
	hostportErr error
	host        string
	hostIp      net.IP
	port        string
	portInt64   int64
	portIntErr  error
}) {
	ret.host, ret.port, ret.hostportErr = net.SplitHostPort(hostport)
	ret.hostIp = net.ParseIP(ret.host)
	ret.portInt64, ret.portIntErr = strconv.ParseInt(ret.port, 0, 64)
	return
}

func (i *Instance) PrintStatus(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "num entries: %d\n", stm.AtomicGet(i.entries).(stmutil.Lenner).Len())
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%d waiters:\n", stm.AtomicGet(i.waiters).(stmutil.Lenner).Len())
	fmt.Fprintf(tw, "num\treason\n")
	stm.AtomicGet(i.waitersByReason).(stmutil.Mappish[any, any]).Range(func(r, ws interface{}) bool {
		fmt.Fprintf(tw, "%d\t%q\n", ws.(stmutil.Settish[any]).Len(), r.(reason))
		return true
	})
	tw.Flush()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "handles:")
	fmt.Fprintf(tw, "protocol\tlocal\tremote\treason\texpires\tcreated\n")
	entries := stm.AtomicGet(i.entries).(stmutil.Mappish[any, any])
	type entriesItem struct {
		Entry
		stmutil.Settish[any]
	}
	entriesItems := make([]entriesItem, 0, entries.Len())
	entries.Range(func(e, hs interface{}) bool {
		entriesItems = append(entriesItems, entriesItem{e.(Entry), hs.(stmutil.Settish[any])})
		return true
	})
	sort.Slice(entriesItems, func(i, j int) bool {
		l := entriesItems[i].Entry
		r := entriesItems[j].Entry
		var ml missinggo.MultiLess
		f := func(l, r string) {
			pl := parseHostPort(l)
			pr := parseHostPort(r)
			ml.NextBool(pl.hostportErr != nil, pr.hostportErr != nil)
			ml.NextBool(pl.hostIp.To4() == nil, pr.hostIp.To4() == nil)
			ml.Compare(bytes.Compare(pl.hostIp, pr.hostIp))
			ml.NextBool(pl.portIntErr != nil, pr.portIntErr != nil)
			ml.StrictNext(pl.portInt64 == pr.portInt64, pl.portInt64 < pr.portInt64)
			ml.StrictNext(pl.port == pr.port, pl.port < pr.port)
		}
		f(l.RemoteAddr, r.RemoteAddr)
		ml.StrictNext(l.Protocol == r.Protocol, l.Protocol < r.Protocol)
		f(l.LocalAddr, r.LocalAddr)
		return ml.Less()
	})
	for _, ei := range entriesItems {
		e := ei.Entry
		ei.Settish.Range(func(_h interface{}) bool {
			h := _h.(*EntryHandle)
			fmt.Fprintf(tw,
				"%q\t%q\t%q\t%q\t%s\t%v ago\n",
				e.Protocol, e.LocalAddr, e.RemoteAddr, h.reason,
				func() interface{} {
					if h.expires.IsZero() {
						return "not done"
					} else {
						return time.Until(h.expires)
					}
				}(),
				time.Since(h.created),
			)
			return true
		})
	}
	tw.Flush()
}
