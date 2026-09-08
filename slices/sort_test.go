package slices

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestSort(t *testing.T) {
	a := []int{3, 2, 1}
	Sort(a, func(left, right int) bool {
		return left < right
	})
	qt.Check(t, qt.DeepEquals(a, []int{1, 2, 3}))
}
