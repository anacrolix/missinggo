package slices

import "reflect"

func FilterInPlace(sl interface{}, f func(interface{}) bool) {
	v := reflect.ValueOf(sl).Elem()
	j := 0
	for i := 0; i < v.Len(); i++ {
		e := v.Index(i)
		if f(e.Interface()) {
			v.Index(j).Set(e)
			j++
		}
	}
	v.SetLen(j)
}
