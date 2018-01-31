package assert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqualDifferentIntTypes(t *testing.T) {
	var a int = 1
	var b int64 = 1
	assert.EqualValues(t, a, b)
	assert.NotEqual(t, a, b)
	assert.NotPanics(t, func() { Equal(a, b) })
	assert.Panics(t, func() { StrictlyEqual(a, b) })
}
