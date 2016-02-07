package bitmap

import (
	"github.com/RoaringBitmap/roaring"

	"github.com/anacrolix/missinggo"
)

// Bitmaps store the existence of values in [0,math.MaxUint32] more
// efficiently than []bool. The empty value starts with no bits set.
type Bitmap struct {
	rb *roaring.RoaringBitmap
}

func (me *Bitmap) Len() int {
	if me.rb == nil {
		return 0
	}
	return int(me.rb.GetCardinality())
}

func (me Bitmap) ToSortedSlice() (ret []int) {
	if me.rb == nil {
		return
	}
	missinggo.CastSlice(&ret, me.rb.ToArray())
	return
}

func (me *Bitmap) lazyRB() *roaring.RoaringBitmap {
	if me.rb == nil {
		me.rb = roaring.NewRoaringBitmap()
	}
	return me.rb
}

func (me *Bitmap) Iter(f func(interface{}) bool) {
	me.IterTyped(func(i int) bool {
		return f(i)
	})
}

func (me Bitmap) IterTyped(f func(int) bool) bool {
	if me.rb == nil {
		return true
	}
	it := me.rb.Iterator()
	for it.HasNext() {
		if !f(int(it.Next())) {
			return false
		}
	}
	return true
}

func (me *Bitmap) Add(is ...int) {
	rb := me.lazyRB()
	for _, i := range is {
		rb.AddInt(i)
	}
}

func (me *Bitmap) AddRange(begin, end int) {
	if begin >= end {
		return
	}
	me.lazyRB().AddRange(uint32(begin), uint32(end))
}

func (me *Bitmap) Remove(i int) {
	if me.rb == nil {
		return
	}
	me.rb.Remove(uint32(i))
}

func (me *Bitmap) Union(other *Bitmap) {
	me.lazyRB().Or(other.lazyRB())
}

func (me *Bitmap) Contains(i int) bool {
	if me.rb == nil {
		return false
	}
	return me.rb.Contains(uint32(i))
}

type Iter struct {
	ii roaring.IntIterable
}

func (me *Iter) Next() bool {
	if me == nil {
		return false
	}
	return me.ii.HasNext()
}

func (me *Iter) Value() interface{} {
	return me.ValueInt()
}

func (me *Iter) ValueInt() int {
	return int(me.ii.Next())
}

func (me *Iter) Stop() {}

func Sub(left, right *Bitmap) *Bitmap {
	return &Bitmap{
		rb: roaring.AndNot(left.lazyRB(), right.lazyRB()),
	}
}

func (me *Bitmap) Sub(other *Bitmap) {
	if other.rb == nil {
		return
	}
	if me.rb == nil {
		return
	}
	me.rb.AndNot(other.rb)
}

func (me *Bitmap) Clear() {
	if me.rb == nil {
		return
	}
	me.rb.Clear()
}

func (me Bitmap) Copy() (ret Bitmap) {
	ret = me
	if ret.rb != nil {
		ret.rb = ret.rb.Clone()
	}
	return
}

func (me *Bitmap) FlipRange(begin, end int) {
	me.lazyRB().FlipInt(begin, end)
}
