package filecache

import "github.com/anacrolix/missinggo/orderedmap"

func newLRUItems() *lruItems {
	return &lruItems{orderedmap.New(func(_a, _b interface{}) bool {
		a := _a.(ItemInfo)
		b := _b.(ItemInfo)
		if a.Accessed.Equal(b.Accessed) {
			return a.Path < b.Path
		}
		return a.Accessed.Before(b.Accessed)
	})}
}

type lruItems struct {
	om orderedmap.OrderedMap
}

func (me *lruItems) Iter(f func(ItemInfo) bool) {
	me.om.Iter(func(i interface{}) bool {
		return f(i.(ItemInfo))
	})
}

func (me *lruItems) LRU() (ret ItemInfo) {
	me.Iter(func(ii ItemInfo) bool {
		ret = ii
		return false
	})
	return
}

func (me *lruItems) Insert(ii ItemInfo) {
	me.om.Set(ii, ii)
}

func (me *lruItems) Remove(ii ItemInfo) bool {
	_, ret := me.om.GetOk(ii)
	me.om.Unset(ii)
	return ret
}
