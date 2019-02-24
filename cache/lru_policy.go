package cache

import (
	"sync"

	"github.com/anacrolix/missinggo"
	"github.com/anacrolix/missinggo/orderedmap"
)

type LruPolicy struct {
	mu     sync.RWMutex
	sorted orderedmap.OrderedMap
	keys   map[Key]Usage
}

type lruItem struct {
	Key
	Usage
}

var _ Policy = (*LruPolicy)(nil)

func (me *LruPolicy) Candidate() (k Key, ok bool) {
	me.mu.RLock()
	defer me.mu.RUnlock()
	if me.sorted == nil {
		return
	}
	me.sorted.Iter(func(i interface{}) bool {
		k = i.(lruItem).Key
		ok = true
		return false
	})
	return
}

func (me *LruPolicy) Forget(k Key) {
	me.mu.Lock()
	defer me.mu.Unlock()
	u, ok := me.keys[k]
	if !ok {
		return
	}
	me.sorted.Unset(lruItem{k, u})
	delete(me.keys, k)
}

func (me *LruPolicy) NumItems() int {
	return len(me.keys)
}

func (me *LruPolicy) Update(k Key, u Usage) {
	me.mu.Lock()
	defer me.mu.Unlock()
	if me.sorted == nil {
		me.sorted = orderedmap.New(func(l, r interface{}) bool {
			_l := l.(lruItem)
			_r := r.(lruItem)
			var ml missinggo.MultiLess
			ml.NextBool(_l.Usage.Less(_r.Usage), _r.Usage.Less(_l.Usage))
			ml.StrictNext(_l.Key == _r.Key, _l.Key < _r.Key)
			return ml.Less()
		})
	}
	if u, ok := me.keys[k]; ok {
		me.sorted.Unset(lruItem{k, u})
	}
	me.sorted.Set(lruItem{k, u}, struct{}{})
	if me.keys == nil {
		me.keys = make(map[Key]Usage)
	}
	me.keys[k] = u
}
