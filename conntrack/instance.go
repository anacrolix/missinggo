package conntrack

import (
	"fmt"
	"io"
	"sync"
	"text/tabwriter"
	"time"
)

type reason = string

type Instance struct {
	maxEntries int
	Timeout    func(Entry) time.Duration

	mu              sync.Mutex
	entries         map[Entry]handles
	waitersByReason map[reason]handles
	waitersByEntry  map[Entry][]*EntryHandle
}

type handles = map[*EntryHandle]struct{}

func NewInstance() *Instance {
	i := &Instance{
		maxEntries: 200,
		Timeout: func(e Entry) time.Duration {
			// udp is the main offender, and the default is allegedly 30s.
			return 30 * time.Second
		},
		entries:         make(map[Entry]handles),
		waitersByReason: make(map[reason]handles),
		waitersByEntry:  make(map[Entry][]*EntryHandle),
	}
	return i
}

func (i *Instance) SetMaxEntries(max int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	prev := i.maxEntries
	i.maxEntries = max
	for j := prev; j < max; j++ {
		i.wakeAny()
	}
}

func (i *Instance) remove(eh *EntryHandle) {
	i.mu.Lock()
	defer i.mu.Unlock()
	hs := i.entries[eh.e]
	delete(hs, eh)
	if len(hs) == 0 {
		delete(i.entries, eh.e)
		i.wakeWaiter(eh.reason)
	}
}

func (i *Instance) chooseWakeReason(avoid reason) reason {
	for k := range i.waitersByReason {
		if k == avoid {
			continue
		}
		return k
	}
	return avoid
}

func (i *Instance) wakeWaiter(avoid reason) {
	r := i.chooseWakeReason(avoid)
	i.wakeReason(r)
}

func (i *Instance) wakeAny() {
	for r := range i.waitersByReason {
		i.wakeReason(r)
		break
	}
}

func (i *Instance) wakeReason(r reason) {
	for k := range i.waitersByReason[r] {
		i.wakeEntry(k.e)
		break
	}
}

func (i *Instance) wakeEntry(e Entry) {
	i.entries[e] = make(handles)
	for _, eh := range i.waitersByEntry[e] {
		i.entries[e][eh] = struct{}{}
		delete(i.waitersByReason[eh.reason], eh)
		eh.added.Unlock()
	}
	delete(i.waitersByEntry, e)
}

func (i *Instance) Wait(e Entry, reason string) (eh *EntryHandle) {
	eh = &EntryHandle{
		reason: reason,
		e:      e,
		i:      i,
	}
	i.mu.Lock()
	hs, ok := i.entries[eh.e]
	if ok {
		hs[eh] = struct{}{}
		i.mu.Unlock()
		return
	}
	if len(i.entries) < i.maxEntries {
		i.entries[eh.e] = handles{
			eh: struct{}{},
		}
		i.mu.Unlock()
		return
	}
	eh.added.Lock()
	if i.waitersByReason[reason] == nil {
		i.waitersByReason[reason] = make(handles)
	}
	i.waitersByReason[reason][eh] = struct{}{}
	i.waitersByEntry[e] = append(i.waitersByEntry[e], eh)
	i.mu.Unlock()
	eh.added.Lock()
	return
}

func (i *Instance) PrintStatus(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	i.mu.Lock()
	fmt.Fprintf(w, "num entries: %d\n", len(i.entries))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "waiters:")
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
