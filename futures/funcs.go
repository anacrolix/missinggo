package futures

import (
	"context"
	"sync"
	"time"

	"github.com/anacrolix/missinggo/slices"
	"github.com/bradfitz/iter"
)

// Sends each future as it completes on the returned chan, closing it when
// everything has been sent.
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

// Additional state maintained for each delayed element.
type delayedState struct {
	timeout *F
	added   bool
}

func AsCompletedDelayed(ctx context.Context, initial []*F, delayed []Delayed) <-chan *F {
	ret := make(chan *F)
	go func() {
		defer close(ret)
		var (
			dss      []delayedState
			timeouts = map[*F]struct{}{} // Pending timeouts
		)
		for i := range delayed {
			func(i int) {
				f := Start(func() (interface{}, error) {
					select {
					case <-time.After(delayed[i].Delay):
						return i, nil
					case <-ctx.Done():
						return nil, ctx.Err()
					}
				})
				timeouts[f] = struct{}{}
				dss = append(dss, delayedState{timeout: f})
			}(i)
		}
		// Number of pending sends for a future.
		results := map[*F]int{}
		for _, f := range initial {
			results[f]++
		}
	start:
		// A slice of futures we want to send when they complete.
		resultsSlice := func() (ret []*F) {
			for f, left := range results {
				for range iter.N(left) {
					ret = append(ret, f)
				}
			}
			return
		}()
		if len(resultsSlice) == 0 {
			for i, ds := range dss {
				if ds.added {
					continue
				}
				// Add this delayed block prematurely.
				delete(timeouts, ds.timeout)
				for _, f := range delayed[i].Fs {
					results[f]++
				}
				dss[i].added = true
				// We need to recompute the results slice.
				goto start
			}
		}
		as := AsCompleted(append(
			resultsSlice,
			slices.FromMapKeys(timeouts).([]*F)...,
		)...)
		for {
			select {
			case <-ctx.Done():
				return
			case f, ok := <-as:
				if !ok {
					return
				}
				if _, ok := timeouts[f]; ok {
					if ctx.Err() != nil {
						break
					}
					i := f.MustResult().(int)
					for _, f := range delayed[i].Fs {
						results[f]++
					}
					delete(timeouts, f)
					dss[i].added = true
					goto start
				}
				select {
				case ret <- f:
					results[f]--
					if results[f] == 0 {
						delete(results, f)
					}
					if len(results) == 0 {
						goto start
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ret
}
