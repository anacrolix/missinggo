package futures

import (
	"context"
	"sync"
	"time"

	"github.com/anacrolix/missinggo/slices"
	"github.com/bradfitz/iter"
)

func AsCompleted(fs ...*F) <-chan *F {
	ret := make(chan *F, len(fs))
	var wg sync.WaitGroup
	for _, f := range fs {
		wg.Add(1)
		go func(f *F) {
			defer wg.Done()
			<-f.Done()
			ret <- f
		}(f)
	}
	go func() {
		wg.Wait()
		close(ret)
	}()
	return ret
}

func AsCompletedDelayed(ctx context.Context, initial []*F, delayed []Delayed) <-chan *F {
	ret := make(chan *F)
	go func() {
		defer close(ret)
		timeouts := make(map[*F]struct{}, len(delayed))
		for _, d := range delayed {
			func(d Delayed) {
				timeouts[Start(func() (interface{}, error) {
					select {
					case <-time.After(d.Delay):
						return d.Fs, nil
					case <-ctx.Done():
						return nil, ctx.Err()
					}
				})] = struct{}{}
			}(d)
		}
		// Number of pending sends for a future.
		results := map[*F]int{}
		for _, f := range initial {
			results[f]++
		}
	start:
		for f := range AsCompleted(append(
			func() (ret []*F) {
				for f, left := range results {
					for range iter.N(left) {
						ret = append(ret, f)
					}
				}
				return
			}(),
			slices.FromMapKeys(timeouts).([]*F)...,
		)...) {
			if _, ok := timeouts[f]; ok {
				for _, f := range f.MustResult().([]*F) {
					results[f]++
				}
				delete(timeouts, f)
				goto start
			}
			select {
			case ret <- f:
				results[f]--
				if results[f] == 0 {
					delete(results, f)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return ret
}
