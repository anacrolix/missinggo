package missinggo

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestSetEvent(t *testing.T) {
	var e Event
	e.Set()
}

func TestEventIsSet(t *testing.T) {
	var e Event
	qt.Check(t, qt.IsFalse(e.IsSet()))
	e.Set()
	qt.Check(t, qt.IsTrue(e.IsSet()))
}
