package missinggo

import (
	"testing"
	"time"

	"github.com/go-quicktest/qt"
)

// Calls suite with the used time.Now function used by MonotonicNow replaced
// with stdNow for the duration of the call.
func withCustomStdNow(stdNow func() time.Time, suite func()) {
	oldStdNow := stdNowFunc
	oldSkew := monotonicSkew
	defer func() {
		stdNowFunc = oldStdNow
		monotonicSkew = oldSkew
	}()
	stdNowFunc = stdNow
	suite()
}

// Returns a time.Now-like function that walks seq returning time.Unix(0,
// seq[i]) in successive calls.
func stdNowSeqFunc(seq []int64) func() time.Time {
	var i int
	return func() time.Time {
		defer func() { i++ }()
		return time.Unix(0, seq[i])
	}
}

func TestMonotonicTime(t *testing.T) {
	started := MonotonicNow()
	withCustomStdNow(stdNowSeqFunc([]int64{2, 1, 3, 3, 2, 3}), func() {
		i0 := MonotonicNow() // 0
		i1 := MonotonicNow() // 1
		qt.Check(t, qt.Equals(i0.Sub(i1), 0))
		qt.Check(t, qt.Equals(MonotonicSince(i0), 2)) // 2
		qt.Check(t, qt.Equals(MonotonicSince(i1), 2)) // 3
		i4 := MonotonicNow()
		qt.Check(t, qt.Equals(i4.Sub(i0), 2))
		qt.Check(t, qt.Equals(i4.Sub(i1), 2))
		i5 := MonotonicNow()
		qt.Check(t, qt.Equals(i5.Sub(i0), 3))
		qt.Check(t, qt.Equals(i5.Sub(i1), 3))
		qt.Check(t, qt.Equals(i5.Sub(i4), 1))
	})
	// Ensure that skew and time function are restored correctly and within
	// reasonable bounds.
	qt.Check(t, qt.IsTrue(MonotonicSince(started) >= 0 && MonotonicSince(started) < time.Second))
}
