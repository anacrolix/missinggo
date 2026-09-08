package missinggo

import (
	"testing"

	"github.com/go-quicktest/qt"
)

func TestEmptyValue(t *testing.T) {
	qt.Check(t, qt.IsTrue(IsZeroValue(false)))
	qt.Check(t, qt.IsFalse(IsZeroValue(true)))
}

func TestUnexportedField(t *testing.T) {
	type FooType1 struct {
		Bar  int
		Dog  bool
		fish string
	}
	fooInstance := FooType1{}

	qt.Check(t, qt.IsTrue(IsZeroValue(fooInstance)))

	fooInstance2 := FooType1{fish: "fishy"}
	qt.Check(t, qt.IsFalse(IsZeroValue(fooInstance2)))

	fooInstance3 := FooType1{Bar: 5}

	qt.Check(t, qt.IsFalse(IsZeroValue(fooInstance3)))
}
