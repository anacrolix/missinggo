package slices

import (
	"testing"

	"github.com/go-quicktest/qt"
)

type herp int

func TestCastSliceInterface(t *testing.T) {
	var dest []herp
	MakeInto(&dest, []interface{}{herp(1), herp(2)})
	qt.Check(t, qt.HasLen(dest, 2))
	qt.Check(t, qt.Equals(dest[0], 1))
	qt.Check(t, qt.Equals(dest[1], 2))
}

func TestCastSliceInts(t *testing.T) {
	var dest []int
	MakeInto(&dest, []uint32{1, 2})
	qt.Check(t, qt.HasLen(dest, 2))
	qt.Check(t, qt.Equals(dest[0], 1))
	qt.Check(t, qt.Equals(dest[1], 2))
}
