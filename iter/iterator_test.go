package iter

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestIterator(t *testing.T) {
	const s = "AAAABBBCCDAABBB"
	si := StringIterator(s)
	for i := range s {
		qt.Assert(t, qt.IsTrue(si.Next()))
		qt.Assert(t, qt.Equals(si.Value().(byte), s[i]))
	}
	qt.Assert(t, qt.IsFalse(si.Next()))
}
