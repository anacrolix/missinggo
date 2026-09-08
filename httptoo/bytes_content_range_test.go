package httptoo

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestParseHTTPContentRange(t *testing.T) {
	for _, _case := range []struct {
		h  string
		cr *BytesContentRange
	}{
		{"", nil},
		{"1-2/*", nil},
		{"bytes=1-2/3", &BytesContentRange{1, 2, 3}},
		{"bytes=12-34/*", &BytesContentRange{12, 34, -1}},
		{" bytes=12-34/*", &BytesContentRange{12, 34, -1}},
		{"  bytes 12-34/56", &BytesContentRange{12, 34, 56}},
		{"  bytes=*/56", &BytesContentRange{-1, -1, 56}},
	} {
		ret, ok := ParseBytesContentRange(_case.h)
		qt.Check(t, qt.Equals(ok, _case.cr != nil))
		if _case.cr != nil {
			qt.Check(t, qt.Equals(ret, *_case.cr))
		}
	}
}
