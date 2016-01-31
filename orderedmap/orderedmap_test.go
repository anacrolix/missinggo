package orderedmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNil(t *testing.T) {
	var om *OrderedMap
	_, ok := om.GetOk(nil)
	assert.False(t, ok)
	assert.Nil(t, om.Get(nil))
	it := om.Iter()
	assert.Panics(t, func() { it.Value() })
	assert.False(t, it.Next())
}

func slice(om *OrderedMap) (ret []interface{}) {
	for it := om.Iter(); it.Next(); {
		ret = append(ret, it.Value())
	}
	return
}

func TestSimple(t *testing.T) {
	om := New(func(l, r interface{}) bool {
		return l.(int) < r.(int)
	})
	om.Set(3, 1)
	om.Set(2, 2)
	om.Set(1, 3)
	assert.EqualValues(t, []interface{}{3, 2, 1}, slice(om))
	om.Set(3, 2)
	om.Unset(2)
	assert.EqualValues(t, []interface{}{3, 2}, slice(om))
	om.Set(-1, 4)
	assert.EqualValues(t, []interface{}{4, 3, 2}, slice(om))
}
