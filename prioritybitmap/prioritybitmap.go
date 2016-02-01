// Package prioritybitmap implements a set of integers ordered by attached
// priorities.
package prioritybitmap

import (
	"sync"

	"github.com/anacrolix/missinggo/itertools"
	"github.com/anacrolix/missinggo/orderedmap"
)

type PriorityBitmap struct {
	inited     bool
	om         *orderedmap.OrderedMap
	priorities map[int]int
	mu         sync.RWMutex
}

func (me *PriorityBitmap) Clear() {
	me.inited = false
}

func (me *PriorityBitmap) delete(key int) {
	p, ok := me.priorities[key]
	if !ok {
		return
	}
	me.mu.Lock()
	keys := me.om.Get(p).(map[int]struct{})
	delete(keys, key)
	if len(keys) == 0 {
		me.om.Unset(p)
	}
	me.mu.Unlock()
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

func (me *PriorityBitmap) Remove(key int) {
	if !me.inited {
		return
	}
	me.delete(key)
	delete(me.priorities, key)
}

func (me *PriorityBitmap) Iter() itertools.Iterator {
	return me.IterTyped()
}

func (me *PriorityBitmap) IterTyped() *Iter {
	if !me.inited {
		return nil
	}
	ret := &Iter{
		it: me.om.Iter(),
		ch: make(chan int),
		mu: &me.mu,
	}
	close(ret.ch)
	me.mu.RLock()
	return ret
}
