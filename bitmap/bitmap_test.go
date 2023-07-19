package bitmap

import (
	"math"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/anacrolix/missinggo/iter"
	"github.com/anacrolix/missinggo/slices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyBitmap(t *testing.T) {
	var bm Bitmap
	assert.False(t, bm.Contains(0))
	bm.Remove(0)
	it := iter.NewIterator(&bm)
	assert.Panics(t, func() { it.Value() })
	assert.False(t, it.Next())
}

func bitmapSlice(bm *Bitmap) (ret []int) {
	sl := iter.IterableAsSlice(bm)
	slices.MakeInto(&ret, sl)
	return
}

func TestSimpleBitmap(t *testing.T) {
	bm := new(Bitmap)
	assert.EqualValues(t, []int(nil), bitmapSlice(bm))
	bm.Add(0)
	assert.True(t, bm.Contains(0))
	assert.False(t, bm.Contains(1))
	assert.EqualValues(t, 1, bm.Len())
	bm.Add(3)
	assert.True(t, bm.Contains(0))
	assert.True(t, bm.Contains(3))
	assert.EqualValues(t, []int{0, 3}, bitmapSlice(bm))
	assert.EqualValues(t, 2, bm.Len())
	bm.Remove(0)
	assert.EqualValues(t, []int{3}, bitmapSlice(bm))
	assert.EqualValues(t, 1, bm.Len())
}

func TestSub(t *testing.T) {
	var left, right Bitmap
	left.Add(2, 5, 4)
	right.Add(3, 2, 6)
	assert.Equal(t, []BitIndex{4, 5}, Sub(left, right).ToSortedSlice())
	assert.Equal(t, []BitIndex{3, 6}, Sub(right, left).ToSortedSlice())
}

func TestSubUninited(t *testing.T) {
	var left, right Bitmap
	assert.Empty(t, Sub(left, right).ToSortedSlice())
}

func TestAddRange(t *testing.T) {
	var bm Bitmap
	bm.AddRange(21, 26)
	bm.AddRange(9, 14)
	bm.AddRange(11, 16)
	bm.Remove(12)
	assert.EqualValues(t, []BitIndex{9, 10, 11, 13, 14, 15, 21, 22, 23, 24, 25}, bm.ToSortedSlice())
	assert.EqualValues(t, 11, bm.Len())
	bm.Clear()
	bm.AddRange(3, 7)
	bm.AddRange(0, 3)
	bm.AddRange(2, 4)
	bm.Remove(3)
	assert.EqualValues(t, []BitIndex{0, 1, 2, 4, 5, 6}, bm.ToSortedSlice())
	assert.EqualValues(t, 6, bm.Len())
}

func TestRemoveRange(t *testing.T) {
	var bm Bitmap
	bm.AddRange(3, 12)
	assert.EqualValues(t, 9, bm.Len())
	bm.RemoveRange(14, ToEnd)
	assert.EqualValues(t, 9, bm.Len())
	bm.RemoveRange(2, 5)
	assert.EqualValues(t, 7, bm.Len())
	bm.RemoveRange(10, ToEnd)
	assert.EqualValues(t, 5, bm.Len())
}

func TestLimits(t *testing.T) {
	var bm Bitmap

	// We can't reliably test out of bounds for systems where int is only 32-bit. Rather than guess
	// for every possible GOARCH, I'll just skip the test here. The BitIndex/int wrapper around
	// roaring's types are bad anyway. See https://github.com/anacrolix/missinggo/issues/16.

	//assert.Panics(t, func() { bm.Add(math.MaxInt64) })

	bm.Add(MaxInt)
	assert.EqualValues(t, 1, bm.Len())
	assert.EqualValues(t, []BitIndex{MaxInt}, bm.ToSortedSlice())
}

func TestRoaringRangeEnd(t *testing.T) {
	r := roaring.New()
	r.Add(roaring.MaxUint32)
	require.EqualValues(t, 1, r.GetCardinality())
	r.RemoveRange(0, roaring.MaxUint32)
	assert.EqualValues(t, 1, r.GetCardinality())
	r.RemoveRange(0, math.MaxUint64)
	assert.EqualValues(t, 0, r.GetCardinality())
}
