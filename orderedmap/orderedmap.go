package orderedmap

import (
	"github.com/ryszard/goskiplist/skiplist"
)

type OrderedMap struct {
	sl *skiplist.SkipList
}

func New(lesser func(l, r interface{}) bool) *OrderedMap {
	return &OrderedMap{skiplist.NewCustomMap(lesser)}
}

func (me *OrderedMap) Set(key interface{}, value interface{}) {
	me.sl.Set(key, value)
}

func (me *OrderedMap) Get(key interface{}) interface{} {
	if me == nil {
		return nil
	}
	ret, _ := me.sl.Get(key)
	return ret
}

func (me *OrderedMap) GetOk(key interface{}) (interface{}, bool) {
	if me == nil {
		return nil, false
	}
	return me.sl.Get(key)
}

type Iter struct {
	it skiplist.Iterator
}

func (me *Iter) Next() bool {
	if me == nil {
		return false
	}
	return me.it.Next()
}

func (me *Iter) Value() interface{} {
	return me.it.Value()
}

func (me *OrderedMap) Iter() *Iter {
	if me == nil {
		return nil
	}
	return &Iter{me.sl.Iterator()}
}

func (me *OrderedMap) Unset(key interface{}) {
	if me == nil {
		return
	}
	me.sl.Delete(key)
}
