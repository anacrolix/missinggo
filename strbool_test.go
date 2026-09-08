package missinggo

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestStringTruth(t *testing.T) {
	for _, s := range []string{
		"",
		" ",
		"\n",
		"\x00",
		"0",
	} {
		t.Run(s, func(t *testing.T) {
			qt.Check(t, qt.IsFalse(StringTruth(s)))
		})
	}
	for _, s := range []string{
		" 1",
		"t",
	} {
		t.Run(s, func(t *testing.T) {
			qt.Check(t, qt.IsTrue(StringTruth(s)))
		})
	}
}
