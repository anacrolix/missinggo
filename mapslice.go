package missinggo

import (
	"reflect"
)

// Creates a []struct{Key K; Value V} for map[K]V.
func MapAsSlice(m interface{}) interface{} {
	mapValue := reflect.ValueOf(m)
	sliceElemType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Key",
			Type: mapValue.Type().Key(),
		},
		{
			Name: "Value",
			Type: mapValue.Type().Elem(),
		},
	})
	sliceValue := reflect.New(reflect.SliceOf(sliceElemType)).Elem()
	for _, key := range mapValue.MapKeys() {
		sliceElem := reflect.New(sliceElemType).Elem()
		sliceElem.Field(0).Set(key)
		sliceElem.Field(1).Set(mapValue.MapIndex(key))
		sliceValue = reflect.Append(sliceValue, sliceElem)
	}
	return sliceValue.Interface()
}
