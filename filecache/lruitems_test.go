package filecache

import (
	"math/rand"
	"testing"
	"time"

	"github.com/bradfitz/iter"
	"github.com/stretchr/testify/assert"
)

func BenchmarkInsert(b *testing.B) {
	for range iter.N(b.N) {
		li := newLRUItems()
		for range iter.N(10000) {
			r := rand.Int63()
			t := time.Unix(r/1e9, r%1e9)
			li.Insert(ItemInfo{
				Accessed: t,
			})
		}
	}
}

func TestLruDuplicateAccessTimes(t *testing.T) {
	li := newLRUItems()
	now := time.Now()
	li.Insert(ItemInfo{Accessed: now, Path: "a"})
	li.Insert(ItemInfo{Accessed: now, Path: "b"})
	assert.EqualValues(t, 2, li.om.Len())
}
