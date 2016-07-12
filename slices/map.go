package slices

import "reflect"

type MapKeyValue struct {
	Key, Value interface{}
}

// Creates a []struct{Key K; Value V} for map[K]V.
func FromMap(m interface{}) (slice []MapKeyValue) {
	mapValue := reflect.ValueOf(m)
	for _, key := range mapValue.MapKeys() {
		slice = append(slice, MapKeyValue{key.Interface(), mapValue.MapIndex(key).Interface()})
	}
	return
}

func FromElems(m interface{}) interface{} {
// Returns all the elements []T, from m where m is map[K]T.
	inValue := reflect.ValueOf(m)
	outValue := reflect.MakeSlice(reflect.SliceOf(inValue.Type().Elem()), inValue.Len(), inValue.Len())
	for i, key := range inValue.MapKeys() {
		outValue.Index(i).Set(inValue.MapIndex(key))
	}
	return outValue.Interface()
}
