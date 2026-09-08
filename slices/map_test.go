package slices

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestFromMap(t *testing.T) {
	sl := FromMap(map[string]int{"two": 2, "one": 1})
	qt.Check(t, qt.HasLen(sl, 2))
	Sort(sl, func(left, right MapItem) bool {
		return left.Key.(string) < right.Key.(string)
	})
	qt.Check(t, qt.DeepEquals(sl, []MapItem{{"one", 1}, {"two", 2}}))
}
