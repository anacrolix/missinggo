package futures

import "sync"

func Start(fn func() (interface{}, error)) *F {
	f := &F{
		done: make(chan struct{}),
	}
	go func() {
		f.setResult(fn())
	}()
	return f
}

type F struct {
	mu     sync.Mutex
	result interface{}
	err    error
	done   chan struct{}
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
