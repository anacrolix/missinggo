package xprometheus

import (
	"sync"
	"testing"

	"github.com/bradfitz/iter"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkExpvarCollector_Collect(b *testing.B) {
	ec := NewExpvarCollector()
	ch := make(chan prometheus.Metric)
	n := 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ch {
			n++
		}
	}()
	b.ReportAllocs()
	for range iter.N(b.N) {
		ec.Collect(ch)
	}
	close(ch)
	wg.Wait()
	b.Logf("collected %d metrics (%f per collect)", n, float64(n)/float64(b.N))
}
