package futures

import (
	"fmt"
	"reflect"
	"sync"
)

func Start(fn func() (interface{}, error)) *F {
	f := &F{
		done: make(chan struct{}),
	}
	go func() {
		f.setResult(fn())
	}()
	return f
}

func StartNoError(fn func() interface{}) *F {
	return Start(func() (interface{}, error) {
		return fn(), nil
	})
}

type F struct {
	name   string
	mu     sync.Mutex
	result interface{}
	err    error
	done   chan struct{}
}

func (f *F) String() string {
	if f.name != "" {
		return f.name
	}
	return fmt.Sprintf("future %p", f)
}

func (f *F) SetName(s string) {
	f.name = s
}

func (f *F) Result() (interface{}, error) {
	<-f.done
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.result, f.err
}

func (f *F) MustResult() interface{} {
	val, err := f.Result()
	if err != nil {
		panic(err)
	}
	return val
}

func (f *F) Done() <-chan struct{} {
	return f.done
}

func (f *F) setResult(result interface{}, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.result = result
	f.err = err
	close(f.done)
}

func (f *F) ScanResult(res interface{}) error {
	_res, err := f.Result()
	reflect.ValueOf(res).Elem().Set(reflect.ValueOf(_res))
	return err
}
