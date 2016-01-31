package orderedmap

import (
	"github.com/ryszard/goskiplist/skiplist"
)

type OrderedMap struct {
	sl *skiplist.SkipList
}

func New() *OrderedMap {
	return &OrderedMap{skiplist.New()}
}

func (me *OrderedMap) Set(key interface{}, value interface{}) {
	me.sl.Set(key, value)
}

func (me *OrderedMap) Get(key interface{}) interface{} {
	ret, _ := me.sl.Get(key)
	return ret
}

func (me *OrderedMap) GetOk(key interface{}) (interface{}, bool) {
	return me.sl.Get(key)
}
