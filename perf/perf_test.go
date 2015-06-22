package perf

import (
	"expvar"
	"strconv"
	"testing"
	"time"

	"github.com/bradfitz/iter"
	"github.com/stretchr/testify/assert"
)

func TestTimer(t *testing.T) {
	tr := NewTimer()
	tr.Stop("hiyo")
	tr.Stop("hiyo")
	em.Get("hiyo").(*expvar.Map).Do(func(kv expvar.KeyValue) {
		t.Log(kv.Key, kv.Value)
	})
}

func BenchmarkStopWarm(b *testing.B) {
	tr := NewTimer()
	for range iter.N(b.N) {
		tr.Stop("a")
	}
}

func BenchmarkStopCold(b *testing.B) {
	tr := NewTimer()
	for i := range iter.N(b.N) {
		tr.Stop(strconv.FormatInt(int64(i), 10))
	}
}

func TestExponent(t *testing.T) {
	for _, c := range []struct {
		s string
		d time.Duration
	}{
		{"-1", 10 * time.Millisecond},
		{"-2", 5 * time.Millisecond},
		{"-2", time.Millisecond},
		{"-3", 500 * time.Microsecond},
		{"-3", 100 * time.Microsecond},
		{"-4", 10 * time.Microsecond},
	} {
		tr := NewTimer()
		time.Sleep(c.d)
		assert.Equal(t, c.s, tr.Stop(c.s), "%s", c.d)
	}
	t.Log(em)
}
