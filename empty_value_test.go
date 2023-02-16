package missinggo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyValue(t *testing.T) {
	assert.True(t, IsZeroValue(false))
	assert.False(t, IsZeroValue(true))
}

func TestUnexportedField(t *testing.T) {
	type FooType1 struct {
		Bar  int
		Dog  bool
		fish string
	}
	fooInstance := FooType1{}

	assert.True(t, IsZeroValue(fooInstance))

	fooInstance2 := FooType1{fish: "fishy"}
	assert.False(t, IsZeroValue(fooInstance2))

	fooInstance3 := FooType1{Bar: 5}

	assert.False(t, IsZeroValue(fooInstance3))
}

func TestUnexportedField(t *testing.T) {
	type FooType1 struct {
		Bar  int
		Dog  bool
		fish string
	}
	fooInstance := FooType1{}

	assert.True(t, IsEmptyValue(reflect.ValueOf(fooInstance)))

	fooInstance2 := FooType1{fish: "fishy"}
	assert.True(t, IsEmptyValue(reflect.ValueOf(fooInstance2)))

	fooInstance3 := FooType1{Bar: 5}

	assert.False(t, IsEmptyValue(reflect.ValueOf(fooInstance3)))
}
