package bitmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyBitmap(t *testing.T) {
	var bm Bitmap
	assert.False(t, bm.Contains(0))
	bm.Remove(0)
	it := bm.Iter()
	assert.Panics(t, func() { it.Value() })
	assert.False(t, it.Next())
}

func bitmapSlice(bm *Bitmap) (ret []int) {
	for it := bm.IterTyped(); it.Next(); {
		ret = append(ret, it.ValueInt())
	}
	return
}

func TestSimpleBitmap(t *testing.T) {
	bm := new(Bitmap)
	assert.EqualValues(t, []int(nil), bitmapSlice(bm))
	bm.Add(0)
	assert.True(t, bm.Contains(0))
	assert.False(t, bm.Contains(1))
	bm.Add(3)
	assert.True(t, bm.Contains(0))
	assert.True(t, bm.Contains(3))
	assert.EqualValues(t, []int{0, 3}, bitmapSlice(bm))
	bm.Remove(0)
	assert.EqualValues(t, []int{3}, bitmapSlice(bm))
}
