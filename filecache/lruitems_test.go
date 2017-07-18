package filecache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLruDuplicateAccessTimes(t *testing.T) {
	var li Policy = new(lru)
	now := time.Now()
	li.Used(key("a"), now)
	li.Used(key("b"), now)
	assert.EqualValues(t, 2, li.NumItems())
}
