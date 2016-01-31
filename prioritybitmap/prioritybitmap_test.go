package prioritybitmap

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/anacrolix/missinggo/itertools"
)

func TestEmpty(t *testing.T) {
	var pb PriorityBitmap
	it := pb.Iter()
	assert.Panics(t, func() { it.Value() })
	assert.False(t, it.Next())
}

func TestIntBounds(t *testing.T) {
	var pb PriorityBitmap
	pb.Set(math.MaxInt32, math.MinInt32)
	pb.Set(math.MinInt32, math.MaxInt32)
	assert.EqualValues(t, []interface{}{math.MaxInt32, math.MinInt32}, itertools.IteratorAsSlice(pb.Iter()))
}

func TestDistinct(t *testing.T) {
	var pb PriorityBitmap
	pb.Set(0, 0)
	pb.Set(1, 1)
	assert.EqualValues(t, []interface{}{0, 1}, itertools.IteratorAsSlice(pb.Iter()))
	pb.Set(0, -1)
	assert.EqualValues(t, []interface{}{0, 1}, itertools.IteratorAsSlice(pb.Iter()))
	pb.Set(1, -2)
	assert.EqualValues(t, []interface{}{1, 0}, itertools.IteratorAsSlice(pb.Iter()))
}

func TestNextAfterIterFinished(t *testing.T) {
	var pb PriorityBitmap
	pb.Set(0, 0)
	it := pb.Iter()
	assert.True(t, it.Next())
	assert.False(t, it.Next())
	assert.False(t, it.Next())
}
