package panicif

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"reflect"
)

func NotNil(x any) {
	if x == nil {
		return
	}
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		var herp *reflect.ValueError
		herp, ok := r.(*reflect.ValueError)
		if !ok {
			panic(r)
		}
		if herp.Method != "reflect.Value.IsNil" {
			panic(r)
		}
	}()
	if !reflect.ValueOf(x).IsNil() {
		panic(x)
	}
}

func Nil(a any) {
	if a == nil {
		panic(a)
	}
}

func Err(err error) {
	if err != nil {
		panic(err)
	}
}

func NotEq(a, b any) {
	if a != b {
		panic(fmt.Sprintf("%v != %v", a, b))
	}
}

func Eq(a, b any) {
	if a == b {
		panic(fmt.Sprintf("%v == %v", a, b))
	}
}

func True(x bool) {
	if x {
		panic(x)
	}
}

func False(x bool) {
	if !x {
		panic(x)
	}
}

func SendBlocks[T any](ch chan<- T, t T) {
	select {
	case ch <- t:
	default:
		panic("send blocked")
	}
}

func GreaterThan[T constraints.Ordered](a, b T) {
	if a > b {
		panic(fmt.Sprintf("%v > %v", a, b))
	}
}

func LessThan[T constraints.Ordered](a, b T) {
	if a < b {
		panic(fmt.Sprintf("%v < %v", a, b))
	}
}
