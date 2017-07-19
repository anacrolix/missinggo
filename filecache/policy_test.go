package filecache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testChooseForgottenKey(t *testing.T, p Policy) {
	assert.Equal(t, 0, p.NumItems())
	assert.Panics(t, func() { p.Choose() })
	p.Used(key("a"), time.Now())
	assert.Equal(t, 1, p.NumItems())
	p.Used(key("a"), time.Now().Add(1))
	assert.Equal(t, 1, p.NumItems())
	p.Forget(key("a"))
	assert.Equal(t, 0, p.NumItems())
	assert.Panics(t, func() { p.Choose() })
}

func testPolicy(t *testing.T, p Policy) {
	testChooseForgottenKey(t, p)
}
