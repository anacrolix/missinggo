package missinggo

import (
	"log"
	"reflect"

	"github.com/bradfitz/iter"
)

type Inheriter interface {
	Base() interface{}
	Visible() interface{}
}

func Dispatch(dest, source interface{}) bool {
	destValue := reflect.ValueOf(dest).Elem()
	destType := destValue.Type()
	sourceValue := reflect.ValueOf(source)
	sourceType := sourceValue.Type()
	if implements(destType, sourceType) {
		destValue.Set(sourceValue)
		return true
	}
	class, ok := source.(Inheriter)
	if !ok {
		return false
	}
	log.Print("inheriting")
	if !visible(destType, reflect.TypeOf(class.Visible())) {
		return false
	}
	return Dispatch(dest, class.Base())
}

func visible(wanted reflect.Type, mask reflect.Type) bool {
	if mask.Kind() == reflect.Ptr {
		mask = mask.Elem()
	}
	switch mask.Kind() {
	case reflect.Interface:
		return mask.ConvertibleTo(wanted)
	case reflect.Struct:
		for fieldIndex := range iter.N(mask.NumField()) {
			field := mask.Field(fieldIndex)
			if field.Type.ConvertibleTo(wanted) {
				return true
			}
		}
	default:
		panic(mask)
	}
	return false
}

func implements(wanted, source reflect.Type) bool {
	return source.ConvertibleTo(wanted)
}
