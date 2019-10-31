package conntrack

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/lukechampine/stm"

	"github.com/anacrolix/missinggo/v2/iter"
)

type reason = string

type Instance struct {
	maxEntries   *stm.Var
	noMaxEntries *stm.Var
	Timeout      func(Entry) time.Duration

	// Occupied slots
	entries *stm.Var

	// priority to entryHandleSet, ordered by priority ascending
	waitersByPriority *stm.Var //Mappish
	waitersByReason   *stm.Var //Mappish
	waitersByEntry    *stm.Var //Mappish
	waiters           *stm.Var // Set
}

type (
	priority int
)

func NewInstance() *Instance {
	i := &Instance{
		// A quarter of the commonly quoted absolute max on a Linux system.
		maxEntries:   stm.NewVar(1 << 14),
		noMaxEntries: stm.NewVar(false),
		Timeout: func(e Entry) time.Duration {
			// udp is the main offender, and the default is allegedly 30s.
			return 30 * time.Second
		},
		entries: stm.NewVar(NewMap()),
		waitersByPriority: stm.NewVar(NewSortedMap(func(l, r interface{}) bool {
			return l.(priority) > r.(priority)
		})),

		waitersByReason: stm.NewVar(NewMap()),
		waitersByEntry:  stm.NewVar(NewMap()),
		waiters:         stm.NewVar(NewSet()),
	}
	return i
}

func (i *Instance) SetNoMaxEntries() {
	stm.AtomicSet(i.noMaxEntries, true)
}

func (i *Instance) SetMaxEntries(max int) {
	stm.Atomically(func(tx *stm.Tx) {
		tx.Set(i.noMaxEntries, false)
		tx.Set(i.maxEntries, max)
	})
}

func (i *Instance) remove(eh *EntryHandle) {
	stm.Atomically(func(tx *stm.Tx) {
		es, _ := deleteFromMapToSet(tx.Get(i.entries).(Mappish), eh.e, eh)
		tx.Set(i.entries, es)
	})
}

func deleteFromMapToSet(m Mappish, mapKey, setElem interface{}) (Mappish, bool) {
	_s, ok := m.Get(mapKey)
	if !ok {
		return m, true
	}
	s := _s.(Set)
	s = s.Delete(setElem)
	if s.Len() == 0 {
		return m.Delete(mapKey), true
	}
	return m.Set(mapKey, s), false
}

func (i *Instance) deleteWaiter(eh *EntryHandle, tx *stm.Tx) {
	tx.Set(i.waiters, tx.Get(i.waiters).(Set).Delete(eh))
	tx.Set(i.waitersByPriority, getLeft(deleteFromMapToSet(tx.Get(i.waitersByPriority).(Mappish), eh.priority, eh)))
	tx.Set(i.waitersByReason, getLeft(deleteFromMapToSet(tx.Get(i.waitersByReason).(Mappish), eh.reason, eh)))
	tx.Set(i.waitersByEntry, getLeft(deleteFromMapToSet(tx.Get(i.waitersByEntry).(Mappish), eh.e, eh)))
}

func (i *Instance) addWaiter(eh *EntryHandle) {
	stm.Atomically(func(tx *stm.Tx) {
		tx.Set(i.waitersByPriority, addToMapToSet(tx.Get(i.waitersByPriority).(Mappish), eh.priority, eh))
		tx.Set(i.waitersByReason, addToMapToSet(tx.Get(i.waitersByReason).(Mappish), eh.reason, eh))
		tx.Set(i.waitersByEntry, addToMapToSet(tx.Get(i.waitersByEntry).(Mappish), eh.e, eh))
		tx.Set(i.waiters, tx.Get(i.waiters).(Set).Add(eh))
	})
}

func addToMapToSet(m Mappish, mapKey, setElem interface{}) Mappish {
	s, ok := m.Get(mapKey)
	if ok {
		s = s.(Set).Add(setElem)
	} else {
		s = NewSet().Add(setElem)
	}
	return m.Set(mapKey, s)
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
	ctxDone := stm.NewVar(false)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		stm.AtomicSet(ctxDone, true)
	}()
	var success bool
	stm.Atomically(func(tx *stm.Tx) {
		es := tx.Get(i.entries).(Mappish)
		if s, ok := es.Get(e); ok {
			tx.Set(i.entries, es.Set(e, s.(Set).Add(eh)))
			success = true
			return
		}
		haveRoom := tx.Get(i.noMaxEntries).(bool) || es.Len() < tx.Get(i.maxEntries).(int)
		var topPrio priority
		if !iter.First(func(prio interface{}) {
			topPrio = prio.(priority)
		}, tx.Get(i.waitersByPriority).(iter.Iterable).Iter) {
			panic("y u no waiting?!")
		}
		if haveRoom && p == topPrio {
			tx.Set(i.entries, addToMapToSet(es, e, eh))
			success = true
			return
		}
		if tx.Get(ctxDone).(bool) {
			success = false
			return
		}
		tx.Retry()
	})
	stm.Atomically(func(tx *stm.Tx) {
		i.deleteWaiter(eh, tx)
	})
	if !success {
		eh = nil
	}
	return
}

func (i *Instance) PrintStatus(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "num entries: %d\n", stm.AtomicGet(i.entries).(Lenner).Len())
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%d waiters:\n", stm.AtomicGet(i.waiters).(Lenner).Len())
	fmt.Fprintf(tw, "num\treason\n")
	stm.AtomicGet(i.waitersByReason).(Mappish).Range(func(r, ws interface{}) bool {
		fmt.Fprintf(tw, "%d\t%q\n", ws.(Set).Len(), r.(reason))
		return true
	})
	tw.Flush()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "handles:")
	fmt.Fprintf(tw, "protocol\tlocal\tremote\treason\texpires\tcreated\n")
	stm.AtomicGet(i.entries).(Mappish).Range(func(_e, hs interface{}) bool {
		e := _e.(Entry)
		hs.(Set).Range(func(_h interface{}) bool {
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
		return true
	})
	tw.Flush()
}
