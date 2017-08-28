package futures

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bradfitz/iter"
	"github.com/stretchr/testify/assert"
)

func TestAsCompletedDelayed(t *testing.T) {
	const u = 10 * time.Millisecond
	var fs []*F
	s := time.Now()
	for i := range iter.N(10) {
		f := timeoutFuture(time.Duration(i) * u)
		f.SetName(fmt.Sprintf("%d", i))
		fs = append(fs, f)
	}
	as := AsCompletedDelayed(
		context.Background(),
		[]*F{fs[0], fs[2]},
		[]Delayed{
			{u, []*F{fs[1]}},
			{3 * u, []*F{fs[0]}},
		},
	)
	a := func(f, when time.Duration) {
		t.Helper()
		assert.Equal(t, fs[f], <-as)
		if time.Since(s) < when*u {
			t.Errorf("%d completed too soon", f)
		}
		if time.Since(s) >= (when+1)*u {
			t.Errorf("%d completed too late", f)
		}
	}
	a(0, 0)
	a(1, 1)
	a(2, 2)
	a(0, 3)
	_, ok := <-as
	assert.False(t, ok)
	assert.True(t, time.Since(s) < 4*u)
}
