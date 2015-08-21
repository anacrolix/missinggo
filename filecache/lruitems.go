package filecache

import "container/list"

type Iterator interface {
	Next() Iterator
	Value() interface{}
}

type listElementIterator struct {
	le *list.Element
}

func (me listElementIterator) Next() Iterator {
	e := me.le.Next()
	if e == nil {
		return nil
	}
	return listElementIterator{e}
}

func (me listElementIterator) Value() interface{} {
	return me.le.Value
}

func newLRUItems() *lruItems {
	return &lruItems{list.New(), make(map[ItemInfo]*list.Element)}
}

type lruItems struct {
	list  *list.List
	elems map[ItemInfo]*list.Element
}

func (me *lruItems) Front() Iterator {
	e := me.list.Front()
	if e == nil {
		return nil
	}
	return listElementIterator{e}
}

func (me *lruItems) LRU() ItemInfo {
	return me.list.Front().Value.(ItemInfo)
}

func (me *lruItems) insertList(ii ItemInfo) *list.Element {
	for e := me.list.Back(); e != nil; e = e.Prev() {
		if ii.Accessed.After(e.Value.(ItemInfo).Accessed) {
			return me.list.InsertAfter(ii, e)
		}
	}
	return me.list.PushFront(ii)
}

func (me *lruItems) Insert(ii ItemInfo) {
	le := me.insertList(ii)
	me.elems[ii] = le
}

func (me *lruItems) Remove(ii ItemInfo) bool {
	e, ok := me.elems[ii]
	if !ok {
		return false
	}
	delete(me.elems, ii)
	if me.list.Remove(e).(ItemInfo) != ii {
		panic("f7u12")
	}
	return true
}
