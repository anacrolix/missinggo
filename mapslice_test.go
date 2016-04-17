package missinggo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapSlice(t *testing.T) {
	type kv struct {
		Key   string
		Value int
	}
	sl := MapAsSlice(map[string]int{"two": 2, "one": 1})
	assert.Len(t, sl, 2)
	assert.EqualValues(t, []struct {
		Key   string
		Value int
	}{{"one", 1}, {"two", 2}}, Sort(sl, func(left, right kv) bool {
		return left.Key < right.Key
	}))
}
