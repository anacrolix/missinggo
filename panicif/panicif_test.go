package panicif

import (
	"syscall"
	"testing"

	"github.com/go-quicktest/qt"
)

func TestUintptrNotNil(t *testing.T) {
	var err error = syscall.Errno(0)
	qt.Assert(t, qt.PanicMatches(func() { NotNil(err) }, "errno 0"))
	NotNil[any](nil)
	NotNil((*int)(nil))
	var i int
	qt.Assert(t, qt.PanicMatches(func() { NotNil(&i) }, "0x.*"))
	err = nil
	NotNil(err)
	var m map[int]int
	NotNil(err)
	m = make(map[int]int)
	qt.Assert(t, qt.PanicMatches(func() { NotNil(m) }, `map\[\]`))
}
