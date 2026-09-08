package prioritybitmap

import (
	"math"
	"testing"

	"github.com/anacrolix/missinggo/iter"

	"github.com/go-quicktest/qt"
)

func TestEmpty(t *testing.T) {
	var pb PriorityBitmap
	it := iter.NewIterator(&pb)
	qt.Check(t, qt.PanicMatches(func() { it.Value() }, ".*"))
	qt.Check(t, qt.IsFalse(it.Next()))
}

func TestIntBounds(t *testing.T) {
	var pb PriorityBitmap
	qt.Check(t, qt.IsTrue(pb.Set(math.MaxInt32, math.MinInt32)))
	qt.Check(t, qt.IsTrue(pb.Set(math.MinInt32, math.MaxInt32)))
	qt.Check(t, qt.DeepEquals(iter.IterableAsSlice(&pb), []interface{}{math.MaxInt32, math.MinInt32}))
}

func TestDistinct(t *testing.T) {
	var pb PriorityBitmap
	qt.Check(t, qt.IsTrue(pb.Set(0, 0)))
	pb.Set(1, 1)
	qt.Check(t, qt.DeepEquals(iter.IterableAsSlice(&pb), []interface{}{0, 1}))
	pb.Set(0, -1)
	qt.Check(t, qt.DeepEquals(iter.IterableAsSlice(&pb), []interface{}{0, 1}))
	pb.Set(1, -2)
	qt.Check(t, qt.DeepEquals(iter.IterableAsSlice(&pb), []interface{}{1, 0}))
}

func TestNextAfterIterFinished(t *testing.T) {
	var pb PriorityBitmap
	pb.Set(0, 0)
	it := iter.NewIterator(&pb)
	qt.Check(t, qt.IsTrue(it.Next()))
	qt.Check(t, qt.IsFalse(it.Next()))
	qt.Check(t, qt.IsFalse(it.Next()))
}

func TestMutationResults(t *testing.T) {
	var pb PriorityBitmap
	qt.Check(t, qt.IsFalse(pb.Remove(1)))
	qt.Check(t, qt.IsTrue(pb.Set(1, -1)))
	qt.Check(t, qt.IsTrue(pb.Set(1, 2)))
	qt.Check(t, qt.IsTrue(pb.Set(2, 2)))
	qt.Check(t, qt.IsTrue(pb.Set(2, -1)))
	qt.Check(t, qt.IsFalse(pb.Set(1, 2)))
	qt.Check(t, qt.DeepEquals(iter.IterableAsSlice(&pb), []interface{}{2, 1}))
	qt.Check(t, qt.IsTrue(pb.Set(1, -1)))
	qt.Check(t, qt.IsFalse(pb.Remove(0)))
	qt.Check(t, qt.IsTrue(pb.Remove(1)))
	qt.Check(t, qt.IsFalse(pb.Remove(0)))
	qt.Check(t, qt.IsFalse(pb.Remove(1)))
	qt.Check(t, qt.IsTrue(pb.Remove(2)))
	qt.Check(t, qt.IsFalse(pb.Remove(2)))
	qt.Check(t, qt.IsFalse(pb.Remove(0)))
	qt.Check(t, qt.IsTrue(pb.IsEmpty()))
	qt.Check(t, qt.HasLen(iter.IterableAsSlice(&pb), 0))
}

func TestDoubleRemove(t *testing.T) {
	var pb PriorityBitmap
	qt.Check(t, qt.IsTrue(pb.Set(0, 0)))
	qt.Check(t, qt.IsTrue(pb.Remove(0)))
	qt.Check(t, qt.IsFalse(pb.Remove(0)))
}
