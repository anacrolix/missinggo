package conntrack

import (
	"fmt"
	"io"
	"sync"
	"text/tabwriter"
	"time"
)

type Instance struct {
	MaxEntries int
	Timeout    func(Entry) time.Duration

	mu      sync.Mutex
	cond    sync.Cond
	entries map[Entry]handles
}

type handles = map[*EntryHandle]struct{}

func NewInstance() *Instance {
	i := &Instance{
		MaxEntries: 1000,
		Timeout: func(e Entry) time.Duration {
			// udp is the main offender, and the default is allegedly 30s.
			return 30 * time.Second
		},
		entries: make(map[Entry]handles),
	}
	i.cond.L = &i.mu
	return i
}

func (i *Instance) remove(eh *EntryHandle) {
	i.mu.Lock()
	defer i.mu.Unlock()
	hs := i.entries[eh.e]
	delete(hs, eh)
	if len(hs) == 0 {
		delete(i.entries, eh.e)
		i.cond.Signal()
	}
}

func (i *Instance) Wait(e Entry, reason string) (eh *EntryHandle) {
	eh = &EntryHandle{
		reason: reason,
		e:      e,
		i:      i,
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	for {
		hs, ok := i.entries[eh.e]
		if ok {
			hs[eh] = struct{}{}
			return
		}
		if len(i.entries) < i.MaxEntries {
			i.entries[eh.e] = handles{
				eh: struct{}{},
			}
			return
		}
		i.cond.Wait()
	}
}

func (i *Instance) PrintTable(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	i.mu.Lock()
	fmt.Fprintf(w, "num entries: %d\n", len(i.entries))
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
