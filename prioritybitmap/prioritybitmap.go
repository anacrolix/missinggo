// Package prioritybitmap implements a set of integers ordered by attached
// priorities.
package prioritybitmap

import (
	"runtime"

	"github.com/anacrolix/missinggo/orderedmap"
)

type PriorityBitmap struct {
	inited     bool
	om         *orderedmap.OrderedMap
	priorities map[int]int
}

func (me *PriorityBitmap) delete(key int) {
	p, ok := me.priorities[key]
	if !ok {
		return
	}
	keys := me.om.Get(p).(map[int]struct{})
	delete(keys, key)
}

func (me *PriorityBitmap) getMake(priority int) map[int]struct{} {
	ret, ok := me.om.GetOk(priority)
	if !ok {
		ret = make(map[int]struct{})
		me.om.Set(priority, ret)
	}
	return ret.(map[int]struct{})
}

func (me *PriorityBitmap) lazyInit() {
	if me.inited {
		return
	}
	me.om = orderedmap.New(func(l, r interface{}) bool {
		return l.(int) < r.(int)
	})
	me.priorities = make(map[int]int)
	me.inited = true
}

func (me *PriorityBitmap) Set(key int, priority int) {
	me.lazyInit()
	me.delete(key)
	me.getMake(priority)[key] = struct{}{}
	me.priorities[key] = priority
}

type Iter struct {
	it  *orderedmap.Iter
	ch  chan int
	cur int
	ok  bool
	gc  chan struct{}
}

func (me *Iter) sendSet() {
	set := me.it.Value().(map[int]struct{})
	for i := range set {
		me.ch <- i
	}
	close(me.ch)
}

func (me *Iter) Next() bool {
	if me == nil {
		return false
	}
	for {
		me.cur, me.ok = <-me.ch
		if me.ok {
			return true
		}
		if !me.it.Next() {
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

func (me *PriorityBitmap) Iter() *Iter {
	if me.om == nil {
		return nil
	}
	ret := &Iter{
		it: me.om.Iter(),
		ch: make(chan int),
		gc: make(chan struct{}),
	}
	close(ret.ch)
	runtime.SetFinalizer(ret, func(it *Iter) {
		// Important not to hold references to ret in this function.
		close(it.gc)
	})
	return ret
}
