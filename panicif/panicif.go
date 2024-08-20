package panicif

import (
	"fmt"
)

func NotNil(x any) {
	if x != nil {
		panic(x)
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
