// Package bitmap provides a []bool/bitmap implementation with standardized
// iteration. Bitmaps are the equivalent of []bool, with improved compression
// for runs of similar values, and faster operations on ranges and the like.
package bitmap

import (
	"math"

	"github.com/RoaringBitmap/roaring"
	"github.com/anacrolix/missinggo/iter"
)

const (
	RangeEnd = 0x100000000
)

type Interface interface {
	Len() int
}

type Bitmap struct {
	rb roaring.Bitmap
}

var ToEnd int = -1
func (me Bitmap) Len() int {
	return int(me.rb.GetCardinality())
}

func (me Bitmap) ToSortedSlice() (ret []int) {
	for _, ui32 := range me.rb.ToArray() {
		ret = append(ret, int(int32(ui32)))
	}
	return
}

func (me Bitmap) Iter(cb iter.Callback) {
	me.IterTyped(func(i int) bool {
		return cb(i)
	})
}

// Returns true if all values were traversed without early termination.
func (me Bitmap) IterTyped(f func(int) bool) bool {
	it := me.rb.Iterator()
	for it.HasNext() {
		if !f(int(it.Next())) {
			return false
		}
	}
	return true
}

func checkInt(i int) {
	if i < math.MinInt32 || i > math.MaxInt32 {
		panic("out of bounds")
	}
}

func (me *Bitmap) Add(is ...int) {
	for _, i := range is {
		checkInt(i)
		me.rb.AddInt(i)
	}
}

func (me *Bitmap) AddRange(begin, end int) {
	me.rb.AddRange(uint64(begin), uint64(end))
}

func (me *Bitmap) Remove(i int) bool {
	return me.rb.CheckedRemove(uint32(i))
}

func (me *Bitmap) Union(other Bitmap) {
	me.rb.Or(other.rb)
}

func (me *Bitmap) Contains(i int) bool {
	return me.rb.Contains(uint32(i))
}

func (me *Bitmap) Clear() {
	me.rb.Clear()
}

func (me Bitmap) Copy() Bitmap {
	return Bitmap{me.rb.Clone()}
}

func (me *Bitmap) FlipRange(begin, end int) {
	me.rb.FlipInt(begin, end)
}

func (me *Bitmap) Get(bit int) bool {
	return me.rb.ContainsInt(bit)
}

func (me *Bitmap) Set(bit int, value bool) {
	if value {
		me.rb.AddInt(bit)
	} else {
		me.rb.Remove(uint32(bit))
	}
}

func (me *Bitmap) RemoveRange(begin, end int) *Bitmap {
	me.rb.RemoveRange(uint64(begin), uint64(end))
	return me
}

func (me Bitmap) IsEmpty() bool {
	return me.rb.IsEmpty()
}
