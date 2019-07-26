package xprometheus

import (
	"expvar"
	"fmt"
	"sync"
	"testing"

	"github.com/bradfitz/iter"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
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

func TestCollectInvalidJsonStringChar(t *testing.T) {
	c := collector{
		descs: make(map[int]*prometheus.Desc),
		f: func(m prometheus.Metric) {
			var iom dto.Metric
			err := m.Write(&iom)
			if err != nil {
				t.Fatal(err)
			} else {
				t.Log(iom)
			}
		},
	}
	v := new(expvar.Map).Init()
	v.Add(fmt.Sprintf("received query %q", "find\xdfnode"), 1)
	c.collectVar(v)
}
