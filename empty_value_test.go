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
