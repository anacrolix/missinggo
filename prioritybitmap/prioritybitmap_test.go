package prioritybitmap

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/anacrolix/missinggo/iter"
)

func TestEmpty(t *testing.T) {
	var pb PriorityBitmap
	it := iter.NewIterator(&pb)
	assert.Panics(t, func() { it.Value() })
	assert.False(t, it.Next())
}

func TestIntBounds(t *testing.T) {
	var pb PriorityBitmap
	assert.True(t, pb.Set(math.MaxInt32, math.MinInt32))
	assert.True(t, pb.Set(math.MinInt32, math.MaxInt32))
	assert.EqualValues(t, []interface{}{math.MaxInt32, math.MinInt32}, iter.IterableAsSlice(&pb))
}

func TestDistinct(t *testing.T) {
	var pb PriorityBitmap
	assert.True(t, pb.Set(0, 0))
	pb.Set(1, 1)
	assert.EqualValues(t, []interface{}{0, 1}, iter.IterableAsSlice(&pb))
	pb.Set(0, -1)
	assert.EqualValues(t, []interface{}{0, 1}, iter.IterableAsSlice(&pb))
	pb.Set(1, -2)
	assert.EqualValues(t, []interface{}{1, 0}, iter.IterableAsSlice(&pb))
}

func TestNextAfterIterFinished(t *testing.T) {
	var pb PriorityBitmap
	pb.Set(0, 0)
	it := iter.NewIterator(&pb)
	assert.True(t, it.Next())
	assert.False(t, it.Next())
	assert.False(t, it.Next())
}

func TestMutationResults(t *testing.T) {
	var pb PriorityBitmap
	assert.False(t, pb.Remove(1))
	assert.True(t, pb.Set(1, -1))
	assert.True(t, pb.Set(1, 2))
	assert.True(t, pb.Set(2, 2))
	assert.True(t, pb.Set(2, -1))
	assert.False(t, pb.Set(1, 2))
	assert.EqualValues(t, []interface{}{2, 1}, iter.IterableAsSlice(&pb))
	assert.True(t, pb.Set(1, -1))
	assert.False(t, pb.Remove(0))
	assert.True(t, pb.Remove(1))
	assert.False(t, pb.Remove(0))
	assert.False(t, pb.Remove(1))
	assert.True(t, pb.Remove(2))
	assert.False(t, pb.Remove(2))
	assert.False(t, pb.Remove(0))
	assert.True(t, pb.IsEmpty())
	assert.Len(t, iter.IterableAsSlice(&pb), 0)
}

func TestDoubleRemove(t *testing.T) {
	var pb PriorityBitmap
	assert.True(t, pb.Set(0, 0))
	assert.True(t, pb.Remove(0))
	assert.False(t, pb.Remove(0))
}
