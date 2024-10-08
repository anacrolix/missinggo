package panicif

import (
	qt "github.com/frankban/quicktest"
	"syscall"
	"testing"
)

func TestUintptrNotNil(t *testing.T) {
	var err error = syscall.Errno(0)
	c := qt.New(t)
	c.Assert(func() { NotNil(err) }, qt.PanicMatches, "errno 0")
	NotNil(nil)
	NotNil((*int)(nil))
	var i int
	c.Assert(func() { NotNil(&i) }, qt.PanicMatches, "0x.*")
	err = nil
	NotNil(err)
	var m map[int]int
	NotNil(err)
	m = make(map[int]int)
	c.Assert(func() { NotNil(m) }, qt.PanicMatches, `map\[\]`)
}
