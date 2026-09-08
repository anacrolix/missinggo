package bitmap

import (
	"math"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/anacrolix/missinggo/iter"
	"github.com/anacrolix/missinggo/slices"

	"github.com/go-quicktest/qt"
)

func TestEmptyBitmap(t *testing.T) {
	var bm Bitmap
	qt.Check(t, qt.IsFalse(bm.Contains(0)))
	bm.Remove(0)
	it := iter.NewIterator(&bm)
	qt.Check(t, qt.PanicMatches(func() { it.Value() }, ".*"))
	qt.Check(t, qt.IsFalse(it.Next()))
}

func bitmapSlice(bm *Bitmap) (ret []int) {
	sl := iter.IterableAsSlice(bm)
	slices.MakeInto(&ret, sl)
	return
}

func TestSimpleBitmap(t *testing.T) {
	bm := new(Bitmap)
	qt.Check(t, qt.DeepEquals(bitmapSlice(bm), []int(nil)))
	bm.Add(0)
	qt.Check(t, qt.IsTrue(bm.Contains(0)))
	qt.Check(t, qt.IsFalse(bm.Contains(1)))
	qt.Check(t, qt.Equals(bm.Len(), 1))
	bm.Add(3)
	qt.Check(t, qt.IsTrue(bm.Contains(0)))
	qt.Check(t, qt.IsTrue(bm.Contains(3)))
	qt.Check(t, qt.DeepEquals(bitmapSlice(bm), []int{0, 3}))
	qt.Check(t, qt.Equals(bm.Len(), 2))
	bm.Remove(0)
	qt.Check(t, qt.DeepEquals(bitmapSlice(bm), []int{3}))
	qt.Check(t, qt.Equals(bm.Len(), 1))
}

func TestSub(t *testing.T) {
	var left, right Bitmap
	left.Add(2, 5, 4)
	right.Add(3, 2, 6)
	qt.Check(t, qt.DeepEquals(Sub(left, right).ToSortedSlice(), []BitIndex{4, 5}))
	qt.Check(t, qt.DeepEquals(Sub(right, left).ToSortedSlice(), []BitIndex{3, 6}))
}

func TestSubUninited(t *testing.T) {
	var left, right Bitmap
	qt.Check(t, qt.HasLen(Sub(left, right).ToSortedSlice(), 0))
}

func TestAddRange(t *testing.T) {
	var bm Bitmap
	bm.AddRange(21, 26)
	bm.AddRange(9, 14)
	bm.AddRange(11, 16)
	bm.Remove(12)
	qt.Check(t, qt.DeepEquals(bm.ToSortedSlice(), []BitIndex{9, 10, 11, 13, 14, 15, 21, 22, 23, 24, 25}))
	qt.Check(t, qt.Equals(bm.Len(), 11))
	bm.Clear()
	bm.AddRange(3, 7)
	bm.AddRange(0, 3)
	bm.AddRange(2, 4)
	bm.Remove(3)
	qt.Check(t, qt.DeepEquals(bm.ToSortedSlice(), []BitIndex{0, 1, 2, 4, 5, 6}))
	qt.Check(t, qt.Equals(bm.Len(), 6))
}

func TestRemoveRange(t *testing.T) {
	var bm Bitmap
	bm.AddRange(3, 12)
	qt.Check(t, qt.Equals(bm.Len(), 9))
	bm.RemoveRange(14, ToEnd)
	qt.Check(t, qt.Equals(bm.Len(), 9))
	bm.RemoveRange(2, 5)
	qt.Check(t, qt.Equals(bm.Len(), 7))
	bm.RemoveRange(10, ToEnd)
	qt.Check(t, qt.Equals(bm.Len(), 5))
}

func TestLimits(t *testing.T) {
	var bm Bitmap

	// We can't reliably test out of bounds for systems where int is only 32-bit. Rather than guess
	// for every possible GOARCH, I'll just skip the test here. The BitIndex/int wrapper around
	// roaring's types are bad anyway. See https://github.com/anacrolix/missinggo/issues/16.

	//qt.Check(t, qt.PanicMatches(func() { bm.Add(math.MaxInt64) }, ".*"))

	bm.Add(MaxInt)
	qt.Check(t, qt.Equals(bm.Len(), 1))
	qt.Check(t, qt.DeepEquals(bm.ToSortedSlice(), []BitIndex{MaxInt}))
}

func TestRoaringRangeEnd(t *testing.T) {
	r := roaring.New()
	r.Add(roaring.MaxUint32)
	qt.Assert(t, qt.Equals(r.GetCardinality(), 1))
	r.RemoveRange(0, roaring.MaxUint32)
	qt.Check(t, qt.Equals(r.GetCardinality(), 1))
	r.RemoveRange(0, math.MaxUint64)
	qt.Check(t, qt.Equals(r.GetCardinality(), 0))
}
