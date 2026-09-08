package orderedmap

import (
	"testing"

	"github.com/anacrolix/missinggo/iter"

	"github.com/go-quicktest/qt"
)

func slice(om OrderedMap) (ret []interface{}) {
	om.Iter(func(i interface{}) bool {
		ret = append(ret, om.Get(i))
		return true
	})
	return
}

func TestSimple(t *testing.T) {
	om := New(func(l, r interface{}) bool {
		return l.(int) < r.(int)
	})
	om.Set(3, 1)
	om.Set(2, 2)
	om.Set(1, 3)
	qt.Check(t, qt.DeepEquals(slice(om), []interface{}{3, 2, 1}))
	om.Set(3, 2)
	om.Unset(2)
	qt.Check(t, qt.DeepEquals(slice(om), []interface{}{3, 2}))
	om.Set(-1, 4)
	qt.Check(t, qt.DeepEquals(slice(om), []interface{}{4, 3, 2}))
}

func TestIterEmpty(t *testing.T) {
	om := New(nil)
	it := iter.NewIterator(om)
	qt.Check(t, qt.PanicMatches(func() { it.Value() }, ".*"))
	qt.Check(t, qt.IsFalse(it.Next()))
	it.Stop()
}
