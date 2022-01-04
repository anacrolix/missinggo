package missinggo

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyValue(t *testing.T) {
	assert.True(t, IsEmptyValue(reflect.ValueOf(false)))
	assert.False(t, IsEmptyValue(reflect.ValueOf(true)))
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
