package missinggo

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestSplitHostMaybePortNoPort(t *testing.T) {
	hmp := SplitHostMaybePort("some.domain")
	qt.Check(t, qt.Equals(hmp.Host, "some.domain"))
	qt.Check(t, qt.IsTrue(hmp.NoPort))
	qt.Check(t, qt.IsNil(hmp.Err))
}

func TestSplitHostMaybePortPort(t *testing.T) {
	hmp := SplitHostMaybePort("some.domain:123")
	qt.Check(t, qt.Equals(hmp.Host, "some.domain"))
	qt.Check(t, qt.Equals(hmp.Port, 123))
	qt.Check(t, qt.IsFalse(hmp.NoPort))
	qt.Check(t, qt.IsNil(hmp.Err))
}

func TestSplitHostMaybePortBadPort(t *testing.T) {
	hmp := SplitHostMaybePort("some.domain:wat")
	qt.Check(t, qt.Equals(hmp.Host, "some.domain"))
	qt.Check(t, qt.Equals(hmp.Port, -1))
	qt.Check(t, qt.IsFalse(hmp.NoPort))
	qt.Check(t, qt.IsNotNil(hmp.Err))
}
