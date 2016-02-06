package orderedmap

import (
	"github.com/google/btree"

	"github.com/anacrolix/missinggo"
	"github.com/anacrolix/missinggo/itertools"
)

type GoogleBTree struct {
	bt     *btree.BTree
	lesser func(l, r interface{}) bool
}

type googleBTreeItem struct {
	less  func(l, r interface{}) bool
	key   interface{}
	value interface{}
}

func (me googleBTreeItem) Less(right btree.Item) bool {
	return me.less(me.key, right.(googleBTreeItem).key)
}

func NewGoogleBTree(lesser func(l, r interface{}) bool) *GoogleBTree {
	return &GoogleBTree{
		bt:     btree.New(32),
		lesser: lesser,
	}
}

func (me *GoogleBTree) Set(key interface{}, value interface{}) {
	me.bt.ReplaceOrInsert(googleBTreeItem{me.lesser, key, value})
}

func (me *GoogleBTree) Get(key interface{}) interface{} {
	ret, _ := me.GetOk(key)
	return ret
}

func (me *GoogleBTree) GetOk(key interface{}) (interface{}, bool) {
	item := me.bt.Get(googleBTreeItem{me.lesser, key, nil})
	if item == nil {
		return nil, false
	}
	return item.(googleBTreeItem).value, true
}

type googleBTreeIter struct {
	valueCh  chan interface{}
	stopped  missinggo.Event
	valueOk  bool
	curValue interface{}
}

func (me *googleBTreeIter) Next() bool {
	me.curValue, me.valueOk = <-me.valueCh
	return me.valueOk
}

func (me *googleBTreeIter) Value() interface{} {
	if !me.valueOk {
		panic(me)
	}
	return me.curValue
}

func (me *googleBTreeIter) Stop() {
	me.stopped.Set()
}

func (me *GoogleBTree) Iter() itertools.Iterator {
	ret := &googleBTreeIter{
		valueCh: make(chan interface{}),
	}
	if me.bt.Len() == 0 {
		close(ret.valueCh)
	} else {
		go func() {
			me.bt.Ascend(func(item btree.Item) bool {
				select {
				case ret.valueCh <- item.(googleBTreeItem).value:
					return true
				case <-ret.stopped.C():
					return false
				}
			})
			close(ret.valueCh)
			ret.stopped.Set()
		}()
	}
	return ret
}

func (me *GoogleBTree) Unset(key interface{}) {
	me.bt.Delete(googleBTreeItem{me.lesser, key, nil})
}

func (me *GoogleBTree) Len() int {
	return me.bt.Len()
}
