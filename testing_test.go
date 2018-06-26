package missinggo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Since GetTestName panics if the test name isn't found, it'll be easy to
// expand the tests if we find weird cases.
func TestGetTestName(t *testing.T) {
	assert.EqualValues(t, "TestGetTestName", GetTestName())
}

func TestGetSubtestName(t *testing.T) {
	t.Run("hello", func(t *testing.T) {
		assert.Contains(t, "TestGetSubtestName", GetTestName())
	})
	t.Run("world", func(t *testing.T) {
		assert.Contains(t, "TestGetSubtestName", GetTestName())
	})
}
