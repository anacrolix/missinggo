package expect

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestEqualDifferentIntTypes(t *testing.T) {
	var a int = 1
	var b int64 = 1
	// Equal in value once converted, but not the same type. qt.Equals is
	// typed, so the coercion testify's EqualValues did is explicit here, and
	// the "not equal" half compares them as any to keep the types in play.
	qt.Check(t, qt.Equals(int64(a), b))
	qt.Check(t, qt.Not(qt.DeepEquals[any](a, b)))
	qt.Check(t, qt.Not(qt.PanicMatches(func() { Equal(a, b) }, ".*")))
	qt.Check(t, qt.PanicMatches(func() { StrictlyEqual(a, b) }, ".*"))
}
