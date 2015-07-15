package perf

import (
	"sync"
)

type TimedLocker struct {
	L    sync.Locker
	Desc string
}

func (me *TimedLocker) Lock() {
	tr := NewTimer()
	me.L.Lock()
	tr.Stop(me.Desc)
}

func (me *TimedLocker) Unlock() {
	me.L.Unlock()
}
