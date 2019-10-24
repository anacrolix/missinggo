package conntrack

import (
	"context"
	"fmt"
	"io"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/benbjohnson/immutable"
)

//go:generate peds -file peds.go -pkg conntrack -sets entryHandleSet<*EntryHandle>

//go:generate jenny mapset.tmpl entryToEntryHandleSet.go entryToEntryHandleSet conntrack Entry *entryHandleSet *EntryHandle

//go:generate jenny mapset.tmpl reasonToEntryHandleSet.go reasonToEntryHandleSet conntrack reason *entryHandleSet *EntryHandle

type reason = string

type Instance struct {
	maxEntries   int
	noMaxEntries bool
	Timeout      func(Entry) time.Duration

	mu sync.Mutex
	// Occupied slots
	entries entryToEntryHandleSet

	// priority to entryHandleSet, ordered by priority ascending
	waitersByPriority *immutable.SortedMap
	waitersByReason   reasonToEntryHandleSet
	waitersByEntry    entryToEntryHandleSet
	waiters           *entryHandleSet
	numWaitersChanged sync.Cond
}

type (
	waitersByPriorityValue = entryHandleSet
	priority               int
)

type priorityComparer struct{}

func (priorityComparer) Compare(l, r interface{}) int {
	_l := l.(priority)
	_r := r.(priority)
	if _l > _r {
		return -1
	} else if _r > _l {
		return 1
	} else {
		return 0
	}
}

func NewInstance() *Instance {
	i := &Instance{
		// A quarter of the commonly quoted absolute max on a Linux system.
		maxEntries: 1 << 14,
		Timeout: func(e Entry) time.Duration {
			// udp is the main offender, and the default is allegedly 30s.
			return 30 * time.Second
		},
		entries:           NewentryToEntryHandleSet(),
		waitersByPriority: immutable.NewSortedMap(priorityComparer{}),

		waitersByReason: NewreasonToEntryHandleSet(),
		waitersByEntry:  NewentryToEntryHandleSet(),
		waiters:         NewentryHandleSet(),
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
	if i.entries.DeleteFromSet(eh.e, eh) {
		i.wakeOne()
	}
}

// Wakes all waiters.
func (i *Instance) wakeAll() {
	for i.waiters.Len() != 0 {
		i.wakeOne()
	}
}

// Wakes the highest priority waiter.
func (i *Instance) wakeOne() {
	iter := i.waitersByPriority.Iterator()
	if iter.Done() {
		return
	}
	_, _value := iter.Next()
	value := _value.(*entryHandleSet)
	value.Range(func(eh *EntryHandle) bool {
		i.wakeEntry(eh.e)
		return false
	})
}

func (i *Instance) deleteWaiter(eh *EntryHandle) {
	i.waiters = i.waiters.Delete(eh)
	_p, _ := i.waitersByPriority.Get(eh.priority)
	p := _p.(*entryHandleSet)
	p = p.Delete(eh)
	if p.Len() == 0 {
		i.waitersByPriority = i.waitersByPriority.Delete(eh.priority)
	} else {
		i.waitersByPriority = i.waitersByPriority.Set(eh.priority, p)
	}
	i.waitersByReason.DeleteFromSet(eh.reason, eh)
	i.waitersByEntry.DeleteFromSet(eh.e, eh)
	i.numWaitersChanged.Broadcast()
}

func (i *Instance) addWaiter(eh *EntryHandle) {
	p, ok := i.waitersByPriority.Get(eh.priority)
	if ok {
		p = p.(*entryHandleSet).Add(eh)
	} else {
		p = NewentryHandleSet(eh)
	}
	i.waitersByPriority = i.waitersByPriority.Set(eh.priority, p)
	i.waitersByReason.AddToSet(eh.reason, eh)
	i.waitersByEntry.AddToSet(eh.e, eh)
	i.waiters = i.waiters.Add(eh)
	i.numWaitersChanged.Broadcast()
}

// Wakes all waiters on an entry. Note that the entry is also woken
// immediately, the waiters are all let through.
func (i *Instance) wakeEntry(e Entry) {
	if i.entries.Contains(e) {
		panic(e)
	}
	i.waitersByEntry.Get(e).Range(func(eh *EntryHandle) bool {
		i.entries.AddToSet(e, eh)
		i.deleteWaiter(eh)
		eh.wake.Unlock()
		return true
	})
	if i.waitersByEntry.Contains(e) {
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
	if i.entries.Contains(eh.e) {
		i.entries.AddToSet(eh.e, eh)
		i.mu.Unlock()
		expvars.Add("waits for existing entry", 1)
		return
	}
	if i.noMaxEntries || i.entries.Len() < i.maxEntries {
		i.entries.AddToSet(eh.e, eh)
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
	if !i.entries.Contains(eh.e) || !i.entries.Get(eh.e).Contains(eh) {
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
	i.waitersByReason.Range(func(r reason, ws *entryHandleSet) bool {
		fmt.Fprintf(tw, "%d\t%q\n", ws.Len(), r)
		return true
	})
	tw.Flush()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "handles:")
	fmt.Fprintf(tw, "protocol\tlocal\tremote\treason\texpires\tcreated\n")
	i.entries.Range(func(e Entry, hs *entryHandleSet) bool {
		hs.Range(func(h *EntryHandle) bool {
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
