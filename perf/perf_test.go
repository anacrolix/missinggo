package perf

import (
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/bradfitz/iter"
)

func TestTimer(t *testing.T) {
	tr := NewTimer()
	tr.Mark("hiyo")
	tr.Mark("hiyo")
	WriteEventsTable(ioutil.Discard)
}

func BenchmarkStopWarm(b *testing.B) {
	tr := NewTimer()
	for range iter.N(b.N) {
		tr.Mark("a")
	}
}

func BenchmarkStopCold(b *testing.B) {
	tr := NewTimer()
	for i := range iter.N(b.N) {
		tr.Mark(strconv.FormatInt(int64(i), 10))
	}
}
