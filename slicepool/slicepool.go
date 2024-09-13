package slicepool

import (
	"sync"

	"golang.org/x/exp/constraints"
)

type Pool[T any] struct {
	p sync.Pool
}

func (me *Pool[T]) Put(b *[]T) {
	// Why bother?
	if cap(*b) == 0 {
		return
	}
	*b = (*b)[:0]
	me.p.Put(b)
}

// Get one of whatever is available. If there's no indication of required cap, then presumably
// pushing an item later will help raise the caps present in the pool.
func (me *Pool[T]) Get() (ret *[]T) {
	gotAny := me.p.Get()
	if gotAny == nil {
		return new([]T)
	}
	ret = gotAny.(*[]T)
	*ret = (*ret)[:0]
	return ret
}

func GetMinCap[PoolT any, Cap constraints.Integer](pool *Pool[PoolT], minCap Cap) *[]PoolT {
	for {
		gotAny := pool.p.Get()
		if gotAny == nil {
			slice := make([]PoolT, 0, minCap)
			return &slice
		}
		b := gotAny.(*[]PoolT)
		if cap(*b) >= int(minCap) {
			return b
		}
	}
}
