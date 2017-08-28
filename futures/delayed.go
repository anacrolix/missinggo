package futures

import "time"

func timeoutFuture(timeout time.Duration) *F {
	return Start(func() (interface{}, error) {
		time.Sleep(timeout)
		return nil, nil
	})
}

type Delayed struct {
	Delay time.Duration
	Fs    []*F
}
