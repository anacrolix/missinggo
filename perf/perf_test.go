package perf

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/bradfitz/iter"
	"github.com/stretchr/testify/assert"
)

func TestTimer(t *testing.T) {
	tr := NewTimer()
	tr.Mark("hiyo")
	tr.Mark("hiyo")
	t.Log(em.Get("hiyo").(*buckets))
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

func TestExponent(t *testing.T) {
	for _, c := range []struct {
		e             int
		d             time.Duration
		humanInterval string
	}{
		{0, 0, "<1ns"},
		{31, 1 * time.Second, "[1s,2s)"},
		{30, 1*time.Second - 1, "[500ms,1s)"},
	} {
		assert.Equal(t, c.e, bucketExponent(c.d))
		assert.Equal(t, c.humanInterval, humanExponent(c.e))
		tr := NewTimer()
		tr.addDuration(fmt.Sprintf("%d", c.e), c.d)
		assert.Equal(t, c.e, bucketExponent(c.d))
	}
}
