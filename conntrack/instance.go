package conntrack

import (
	"context"
	"fmt"
	"io"
	"sync"
	"text/tabwriter"
	"time"
)

type reason = string

type Instance struct {
	maxEntries   int
	noMaxEntries bool
	Timeout      func(Entry) time.Duration

	mu sync.Mutex
	// Occupied slots
	entries Mappish

	// priority to entryHandleSet, ordered by priority ascending
	waitersByPriority Mappish
	waitersByReason   Mappish
	waitersByEntry    Mappish
	waiters           Set
	numWaitersChanged sync.Cond
}

type (
	priority int
)

func NewInstance() *Instance {
	i := &Instance{
		// A quarter of the commonly quoted absolute max on a Linux system.
		maxEntries: 1 << 14,
		Timeout: func(e Entry) time.Duration {
			// udp is the main offender, and the default is allegedly 30s.
			return 30 * time.Second
		},
		entries: NewMap(),
		waitersByPriority: NewSortedMap(func(l, r interface{}) bool {
			return l.(priority) > r.(priority)
		}),

		waitersByReason: NewMap(),
		waitersByEntry:  NewMap(),
		waiters:         NewSet(),
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
	var ok bool
	if i.entries, ok = deleteFromMapToSet(i.entries, eh.e, eh); ok {
		i.wakeOne()
	}
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

// Wakes all waiters.
func (i *Instance) wakeAll() {
	for i.waiters.Len() != 0 {
		i.wakeOne()
	}
}

// Wakes the highest priority waiter.
func (i *Instance) wakeOne() {
	i.waitersByPriority.Range(func(_, value interface{}) bool {
		value.(Set).Range(func(elem interface{}) bool {
			i.wakeEntry(elem.(*EntryHandle).e)
			return false
		})
		return false
	})
}

func (i *Instance) deleteWaiter(eh *EntryHandle) {
	i.waiters = i.waiters.Delete(eh)
	i.waitersByPriority, _ = deleteFromMapToSet(i.waitersByPriority, eh.priority, eh)
	i.waitersByReason, _ = deleteFromMapToSet(i.waitersByReason, eh.reason, eh)
	i.waitersByEntry, _ = deleteFromMapToSet(i.waitersByEntry, eh.e, eh)
	i.numWaitersChanged.Broadcast()
}

func (i *Instance) addWaiter(eh *EntryHandle) {
	i.waitersByPriority = addToMapToSet(i.waitersByPriority, eh.priority, eh)
	i.waitersByReason = addToMapToSet(i.waitersByReason, eh.reason, eh)
	i.waitersByEntry = addToMapToSet(i.waitersByEntry, eh.e, eh)
	i.waiters = i.waiters.Add(eh)
	i.numWaitersChanged.Broadcast()
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

// Wakes all waiters on an entry. Note that the entry is also woken
// immediately, the waiters are all let through.
func (i *Instance) wakeEntry(e Entry) {
	if _, ok := i.entries.Get(e); ok {
		panic(e)
	}
	ws, _ := i.waitersByEntry.Get(e)
	ws.(Set).Range(func(_eh interface{}) bool {
		eh := _eh.(*EntryHandle)
		i.entries = addToMapToSet(i.entries, e, eh)
		i.deleteWaiter(eh)
		eh.wake.Unlock()
		return true
	})
	if _, ok := i.waitersByEntry.Get(e); ok {
		panic(e)
	}
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
	i.mu.Lock()
	if _, ok := i.entries.Get(eh.e); ok {
		i.entries = addToMapToSet(i.entries, eh.e, eh)
		i.mu.Unlock()
		expvars.Add("waits for existing entry", 1)
		return
	}
	if i.noMaxEntries || i.entries.Len() < i.maxEntries {
		i.entries = addToMapToSet(i.entries, eh.e, eh)
		i.mu.Unlock()
		expvars.Add("waits with space in table", 1)
		return
	}
	// Lock the mutex, so that a following Lock will block until it's unlocked by a wake event.
	eh.wake.Lock()
	i.addWaiter(eh)
	i.mu.Unlock()
	expvars.Add("waits that blocked", 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		i.mu.Lock()
		if i.waiters.Contains(eh) {
			i.deleteWaiter(eh)
			eh.wake.Unlock()
		}
		i.mu.Unlock()
	}()
	// Blocks until woken by an Unlock.
	eh.wake.Lock()
	i.mu.Lock()
	if s, ok := i.entries.Get(eh.e); !ok || s.(Set).Contains(eh) {
		eh = nil
	}
	i.mu.Unlock()
	return
}

func (i *Instance) PrintStatus(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	i.mu.Lock()
	fmt.Fprintf(w, "num entries: %d\n", i.entries.Len())
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%d waiters:\n", i.waiters.Len())
	fmt.Fprintf(tw, "num\treason\n")
	i.waitersByReason.Range(func(r, ws interface{}) bool {
		fmt.Fprintf(tw, "%d\t%q\n", ws.(Set).Len(), r.(reason))
		return true
	})
	tw.Flush()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "handles:")
	fmt.Fprintf(tw, "protocol\tlocal\tremote\treason\texpires\tcreated\n")
	i.entries.Range(func(_e, hs interface{}) bool {
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
	i.mu.Unlock()
	tw.Flush()
}
