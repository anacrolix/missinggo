package conntrack

import (
	"math"
	"sync"
	"testing"
	"time"

	_ "github.com/anacrolix/envpprof"
	"github.com/bradfitz/iter"
)

func TestWaitingForSameEntry(t *testing.T) {
	i := NewInstance()
	i.SetMaxEntries(1)
	i.Timeout = func(Entry) time.Duration {
		return 0
	}
	e1 := Entry{"", "", "1"}
	e1h1 := i.Wait(e1, "1", 0)
	gotE2s := make(chan struct{})
	for range iter.N(2) {
		go func() {
			i.Wait(Entry{"", "", "2"}, "", 0)
			gotE2s <- struct{}{}
		}()
	}
	gotE1 := make(chan struct{})
	var e1h2 *EntryHandle
	go func() {
		e1h2 = i.Wait(e1, "2", 0)
		gotE1 <- struct{}{}
	}()
	select {
	case <-gotE1:
	case <-gotE2s:
		t.FailNow()
	}
	go e1h1.Done()
	go e1h2.Done()
	<-gotE2s
	<-gotE2s
}

func TestInstanceSetNoMaxEntries(t *testing.T) {
	i := NewInstance()
	i.SetMaxEntries(0)
	var wg sync.WaitGroup
	wait := func(e Entry, p priority) {
		i.Wait(e, "", p)
		wg.Done()
	}
	for _, e := range []Entry{{"", "", ""}, {"", "", "1"}} {
		for _, p := range []priority{math.MinInt32, math.MaxInt32} {
			wg.Add(1)
			go wait(e, p)
		}
	}
	waitForNumWaiters := func(num int) {
		i.mu.Lock()
		for i.numWaiters != num {
			i.numWaitersChanged.Wait()
		}
		i.mu.Unlock()
	}
	waitForNumWaiters(4)
	i.SetNoMaxEntries()
	waitForNumWaiters(0)
	wg.Wait()
}
