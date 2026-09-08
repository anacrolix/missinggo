package missinggo

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestJitterDuration(t *testing.T) {
	qt.Check(t, qt.Equals(JitterDuration(0, 0), 0))
	qt.Check(t, qt.PanicMatches(func() { JitterDuration(1, -1) }, ".*"))
}
