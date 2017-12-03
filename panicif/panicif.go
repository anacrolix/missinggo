package panicif

func NotNil(x interface{}) {
	if x != nil {
		panic(x)
	}
}
