package filecache

import (
	"testing"
	"time"

	"github.com/go-quicktest/qt"
)

func testChooseForgottenKey(t *testing.T, p Policy) {
	qt.Check(t, qt.Equals(p.NumItems(), 0))
	qt.Check(t, qt.PanicMatches(func() { p.Choose() }, ".*"))
	p.Used(key("a"), time.Now())
	qt.Check(t, qt.Equals(p.NumItems(), 1))
	p.Used(key("a"), time.Now().Add(1))
	qt.Check(t, qt.Equals(p.NumItems(), 1))
	p.Forget(key("a"))
	qt.Check(t, qt.Equals(p.NumItems(), 0))
	qt.Check(t, qt.PanicMatches(func() { p.Choose() }, ".*"))
}

func testPolicy(t *testing.T, p Policy) {
	testChooseForgottenKey(t, p)
}
