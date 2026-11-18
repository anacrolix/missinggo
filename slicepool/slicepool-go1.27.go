//go:build go1.27

package slicepool

import "golang.org/x/exp/constraints"

// Gets a slice with at least minCap capacity, allocating one if the pool has none suitable.
func (me *Pool[T]) GetMinCap[Cap constraints.Integer](minCap Cap) *[]T {
	return getMinCap(me, minCap)
}
