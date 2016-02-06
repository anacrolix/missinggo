package prioritybitmap

import (
	"github.com/anacrolix/missinggo"
	"github.com/anacrolix/missinggo/itertools"
)

type Iter struct {
	it itertools.Iterator
	ch chan int
	// The current iterator value.
	cur int
	// Whether there is a valid iterator value available.
	ok      bool
	stopped missinggo.Event
	// RLocked once when the Iter is created, with a nested RLock while
	// iterating each priority slot.
	mu missinggo.RWLocker
}

// Sends the bits in a priority slot. Avoid reference to *Iter in case GC
// collection is used.
func (me *Iter) sendSet() {
	defer close(me.ch)
	me.mu.RLock()
	defer me.mu.RUnlock()
	switch bits := me.it.Value().(type) {
	case int:
		select {
		case me.ch <- bits:
		case <-me.stopped.C():
			return
		}
	case map[int]struct{}:
		for i := range bits {
			select {
			case me.ch <- i:
			case <-me.stopped.C():
				return
			}
		}
	}
}

func (me *Iter) Stop() {
	if me.stopped.Set() {
		me.mu.RUnlock()
	}
	me.ok = false
}

func (me *Iter) Next() bool {
	if me == nil {
		return false
	}
	for {
		if me.stopped.IsSet() {
			return false
		}
		me.cur, me.ok = <-me.ch
		if me.ok {
			return true
		}
		if !me.it.Next() {
			me.Stop()
			return false
		}
		me.ch = make(chan int)
		go me.sendSet()
	}
}

func (me *Iter) Value() interface{} {
	return me.ValueInt()
}

func (me *Iter) ValueInt() int {
	if !me.ok {
		panic("no value")
	}
	return me.cur
}
