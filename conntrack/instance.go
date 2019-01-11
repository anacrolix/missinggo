package conntrack

import (
	"fmt"
	"io"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/anacrolix/missinggo/orderedmap"
)

type reason = string

type Instance struct {
	maxEntries   int
	noMaxEntries bool
	Timeout      func(Entry) time.Duration

	mu      sync.Mutex
	entries map[Entry]handles
	// priority to entryHandleSet, ordered by priority ascending
	waitersByPriority orderedmap.OrderedMap
	waitersByReason   map[reason]entryHandleSet
	waitersByEntry    map[Entry][]*EntryHandle
	numWaiters        int
	numWaitersChanged sync.Cond
}

type (
	entryHandleSet         = map[*EntryHandle]struct{}
	waitersByPriorityValue = entryHandleSet
	priority               int
	handles                = map[*EntryHandle]struct{}
)

func NewInstance() *Instance {
	i := &Instance{
		// A quarter of the commonly quoted absolute max on a Linux system.
		maxEntries: 1 << 14,
		Timeout: func(e Entry) time.Duration {
			// udp is the main offender, and the default is allegedly 30s.
			return 30 * time.Second
		},
		entries: make(map[Entry]handles),
		waitersByPriority: orderedmap.New(func(_l, _r interface{}) bool {
			return _l.(priority) > _r.(priority)
		}),
		waitersByReason: make(map[reason]entryHandleSet),
		waitersByEntry:  make(map[Entry][]*EntryHandle),
	}
	i.numWaitersChanged.L = &i.mu
	return i
}

func (i *Instance) SetNoMaxEntries() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.noMaxEntries = true
	i.wakeAll()
}

func (i *Instance) SetMaxEntries(max int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.noMaxEntries = false
	prev := i.maxEntries
	i.maxEntries = max
	for j := prev; j < max; j++ {
		i.wakeOne()
	}
}

func (i *Instance) remove(eh *EntryHandle) {
	i.mu.Lock()
	defer i.mu.Unlock()
	hs := i.entries[eh.e]
	delete(hs, eh)
	if len(hs) == 0 {
		delete(i.entries, eh.e)
		i.wakeOne()
	}
}

// Wakes all waiters.
func (i *Instance) wakeAll() {
	for i.numWaiters != 0 {
		i.wakeOne()
	}
}

// Wakes the highest priority waiter.
func (i *Instance) wakeOne() {
	i.waitersByPriority.Iter(func(key interface{}) bool {
		value := i.waitersByPriority.Get(key).(entryHandleSet)
		for eh := range value {
			i.wakeEntry(eh.e)
			break
		}
		return false
	})
}

func (i *Instance) deleteWaiter(eh *EntryHandle) {
	p := i.waitersByPriority.Get(eh.priority).(entryHandleSet)
	delete(p, eh)
	if len(p) == 0 {
		i.waitersByPriority.Unset(eh.priority)
	}
	r := i.waitersByReason[eh.reason]
	delete(r, eh)
	if len(r) == 0 {
		delete(i.waitersByReason, eh.reason)
	}
	i.numWaiters--
	i.numWaitersChanged.Broadcast()
}

func (i *Instance) addWaiter(eh *EntryHandle) {
	p, ok := i.waitersByPriority.GetOk(eh.priority)
	if ok {
		p.(entryHandleSet)[eh] = struct{}{}
	} else {
		i.waitersByPriority.Set(eh.priority, entryHandleSet{eh: struct{}{}})
	}
	if r := i.waitersByReason[eh.reason]; r == nil {
		i.waitersByReason[eh.reason] = entryHandleSet{eh: struct{}{}}
	} else {
		r[eh] = struct{}{}
	}
	i.waitersByEntry[eh.e] = append(i.waitersByEntry[eh.e], eh)
	i.numWaiters++
	i.numWaitersChanged.Broadcast()
}

// Wakes all waiters on an entry. Note that the entry is also added
// immediately, the waiters are all let through.
func (i *Instance) wakeEntry(e Entry) {
	if _, ok := i.entries[e]; ok {
		panic(e)
	}
	i.entries[e] = make(handles)
	for _, eh := range i.waitersByEntry[e] {
		i.entries[e][eh] = struct{}{}
		i.deleteWaiter(eh)
		eh.added.Unlock()
	}
	delete(i.waitersByEntry, e)
}

func (i *Instance) Wait(e Entry, reason string, p priority) (eh *EntryHandle) {
	eh = &EntryHandle{
		reason:   reason,
		e:        e,
		i:        i,
		priority: p,
	}
	i.mu.Lock()
	hs, ok := i.entries[eh.e]
	if ok {
		hs[eh] = struct{}{}
		i.mu.Unlock()
		expvars.Add("waits for existing entry", 1)
		return
	}
	if i.noMaxEntries || len(i.entries) < i.maxEntries {
		i.entries[eh.e] = handles{
			eh: struct{}{},
		}
		i.mu.Unlock()
		expvars.Add("waits with space in table", 1)
		return
	}
	eh.added.Lock()
	i.addWaiter(eh)
	i.mu.Unlock()
	expvars.Add("waits that blocked", 1)
	eh.added.Lock()
	return
}

func (i *Instance) PrintStatus(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	i.mu.Lock()
	fmt.Fprintf(w, "num entries: %d\n", len(i.entries))
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%d waiters:\n", i.numWaiters)
	fmt.Fprintf(tw, "num\treason\n")
	for r, ws := range i.waitersByReason {
		fmt.Fprintf(tw, "%d\t%q\n", len(ws), r)
	}
	tw.Flush()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "handles:")
	fmt.Fprintf(tw, "protocol\tlocal\tremote\treason\texpires\n")
	for e, hs := range i.entries {
		for h := range hs {
			fmt.Fprintf(tw,
				"%q\t%q\t%q\t%q\t%s\n",
				e.Protocol, e.LocalAddr, e.RemoteAddr, h.reason,
				func() interface{} {
					if h.expires.IsZero() {
						return "not done"
					} else {
						return time.Until(h.expires)
					}
				}(),
			)
		}
	}
	i.mu.Unlock()
	tw.Flush()
}
