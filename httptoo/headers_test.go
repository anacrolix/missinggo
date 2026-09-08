package httptoo

import (
	"testing"
	"time"

	"github.com/go-quicktest/qt"
)

func TestCacheControlHeaderString(t *testing.T) {
	qt.Check(t, qt.Equals(CacheControlHeader{
		MaxAge:  12 * time.Hour,
		Caching: Public,
	}.String(), "public, max-age=43200"))
}
