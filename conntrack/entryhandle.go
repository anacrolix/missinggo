package conntrack

import "time"

type EntryHandle struct {
	reason  string
	e       Entry
	i       *Instance
	expires time.Time
}

// func (eh *EntryHandle) Forget() {
// 	eh.remove()
// }

func (eh *EntryHandle) Done() {
	timeout := eh.timeout()
	eh.expires = time.Now().Add(timeout)
	time.AfterFunc(eh.timeout(), eh.remove)
}

func (eh *EntryHandle) remove() {
	eh.i.remove(eh)
}

func (eh *EntryHandle) timeout() time.Duration {
	return eh.i.Timeout(eh.e)
}
