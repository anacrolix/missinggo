package prioritybitmap

import (
	"github.com/anacrolix/missinggo"
	"github.com/anacrolix/missinggo/orderedmap"
)

type Iter struct {
	it      *orderedmap.Iter
	ch      chan int
	cur     int
	ok      bool
	stopped missinggo.Event
}

// Sends the bits in a priority slot. Avoid reference to *Iter in case GC
// collection is used.
func sendSet(out chan int, stop <-chan struct{}, set map[int]struct{}) {
	defer close(out)
	for i := range set {
		select {
		case out <- i:
		case <-stop:
			return
		}
	}
}

func (me *Iter) Stop() {
	me.stopped.Set()
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
		go sendSet(me.ch, me.stopped.C(), me.it.Value().(map[int]struct{}))
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
