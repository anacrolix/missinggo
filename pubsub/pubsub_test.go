package pubsub

import (
	"sync"
	"testing"

	"github.com/bradfitz/iter"

	"github.com/go-quicktest/qt"
)

func TestDoubleClose(t *testing.T) {
	var ps PubSub[any]
	ps.Close()
	ps.Close()
}

func testBroadcast(t testing.TB, subs, vals int) {
	var ps PubSub[int]
	var wg sync.WaitGroup
	for range iter.N(subs) {
		wg.Add(1)
		s := ps.Subscribe()
		go func() {
			defer wg.Done()
			var e int
			for i := range s.Values {
				qt.Check(t, qt.Equals(i, e))
				e++
			}
			qt.Check(t, qt.Equals(e, vals))
		}()
	}
	for i := range iter.N(vals) {
		ps.Publish(i)
	}
	ps.Close()
	wg.Wait()
}

func TestBroadcast(t *testing.T) {
	testBroadcast(t, 100, 10)
}

func BenchmarkBroadcast(b *testing.B) {
	for range iter.N(b.N) {
		testBroadcast(b, 10, 1000)
	}
}

func BenchmarkPublishNoSubscribers(b *testing.B) {
	b.ReportAllocs()
	var ps PubSub[int]
	for b.Loop() {
		ps.Publish(0)
	}
}

func TestCloseSubscription(t *testing.T) {
	var ps PubSub[int]
	ps.Publish(1)
	s := ps.Subscribe()
	select {
	case <-s.Values:
		t.FailNow()
	default:
	}
	ps.Publish(2)
	s2 := ps.Subscribe()
	ps.Publish(3)
	qt.Assert(t, qt.Equals(<-s.Values, 2))
	qt.Assert(t, qt.Equals(<-s.Values, 3))
	s.Close()
	_, ok := <-s.Values
	qt.Assert(t, qt.IsFalse(ok))
	ps.Publish(4)
	ps.Close()
	qt.Assert(t, qt.Equals(<-s2.Values, 3))
	qt.Assert(t, qt.Equals(<-s2.Values, 4))
	qt.Assert(t, qt.Equals(<-s2.Values, 0))
	s2.Close()
}
