package filecache

import (
	"testing"
	"time"

	"github.com/go-quicktest/qt"
)

func TestLruDuplicateAccessTimes(t *testing.T) {
	var li Policy = new(lru)
	now := time.Now()
	li.Used(key("a"), now)
	li.Used(key("b"), now)
	qt.Check(t, qt.Equals(li.NumItems(), 2))
}
