package missinggo

import (
	"testing"

	"github.com/go-quicktest/qt"
)

// Since GetTestName panics if the test name isn't found, it'll be easy to
// expand the tests if we find weird cases.
func TestGetTestName(t *testing.T) {
	qt.Check(t, qt.Equals(GetTestName(), "TestGetTestName"))
}

func TestGetSubtestName(t *testing.T) {
	t.Run("hello", func(t *testing.T) {
		qt.Check(t, qt.StringContains("TestGetSubtestName", GetTestName()))
	})
	t.Run("world", func(t *testing.T) {
		qt.Check(t, qt.StringContains("TestGetSubtestName", GetTestName()))
	})
}
