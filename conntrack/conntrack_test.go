package conntrack

import (
	"context"
	"math"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_ "github.com/anacrolix/envpprof"
	"github.com/bradfitz/iter"
)

func entry(id int) Entry {
	return Entry{"", "", strconv.FormatInt(int64(id), 10)}
}

func TestWaitingForSameEntry(t *testing.T) {
	i := NewInstance()
	i.SetMaxEntries(1)
	i.Timeout = func(Entry) time.Duration {
		return 0
	}
	e1h1 := i.WaitDefault(context.Background(), entry(1))
	gotE2s := make(chan struct{})
	for range iter.N(2) {
		go func() {
			i.WaitDefault(context.Background(), entry(2))
			gotE2s <- struct{}{}
		}()
	}
	gotE1 := make(chan struct{})
	var e1h2 *EntryHandle
	go func() {
		e1h2 = i.WaitDefault(context.Background(), entry(1))
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
		i.Wait(context.Background(), e, "", p)
		wg.Done()
	}
	for _, e := range []Entry{entry(0), entry(1)} {
		for _, p := range []priority{math.MinInt32, math.MaxInt32} {
			wg.Add(1)
			go wait(e, p)
		}
	}
	waitForNumWaiters := func(num int) {
		i.mu.Lock()
		for len(i.waiters) != num {
			i.numWaitersChanged.Wait()
		}
		i.mu.Unlock()
	}
	waitForNumWaiters(4)
	i.SetNoMaxEntries()
	waitForNumWaiters(0)
	wg.Wait()
}

func TestWaitReturnsNilContextCompleted(t *testing.T) {
	i := NewInstance()
	i.SetMaxEntries(0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.Nil(t, i.WaitDefault(ctx, entry(0)))
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
	assert.Nil(t, i.WaitDefault(ctx, entry(1)))
	cancel()
}

func TestWaitContextCanceledButRoomForEntry(t *testing.T) {
	i := NewInstance()
	i.SetMaxEntries(1)
	ctx, cancel := context.WithCancel(context.Background())
	go cancel()
	eh := i.WaitDefault(ctx, entry(0))
	if eh == nil {
		assert.Error(t, ctx.Err())
	} else {
		eh.Done()
	}
}

func TestUnlimitedInstance(t *testing.T) {
	i := NewInstance()
	i.SetNoMaxEntries()
	i.Timeout = func(Entry) time.Duration { return 0 }
	eh := i.WaitDefault(context.Background(), entry(0))
	assert.NotNil(t, eh)
	i.mu.Lock()
	assert.Len(t, i.entries[eh.e], 1)
	i.mu.Unlock()
	eh.Done()
	i.mu.Lock()
	assert.Nil(t, i.entries[eh.e])
	i.mu.Unlock()
}

func TestUnlimitedInstanceContextCanceled(t *testing.T) {
	i := NewInstance()
	i.SetNoMaxEntries()
	i.Timeout = func(Entry) time.Duration { return 0 }
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	eh := i.WaitDefault(ctx, entry(0))
	assert.NotNil(t, eh)
	i.mu.Lock()
	assert.Len(t, i.entries[eh.e], 1)
	i.mu.Unlock()
	eh.Done()
	i.mu.Lock()
	assert.Nil(t, i.entries[eh.e])
	i.mu.Unlock()
}

func TestContextCancelledWhileWaiting(t *testing.T) {
	i := NewInstance()
	i.SetMaxEntries(0)
	ctx, cancel := context.WithCancel(context.Background())
	i.mu.Lock()
	assert.Len(t, i.waiters, 0)
	i.mu.Unlock()
	waitReturned := make(chan struct{})
	go func() {
		eh := i.WaitDefault(ctx, entry(0))
		assert.Nil(t, eh)
		close(waitReturned)
	}()
	for {
		i.mu.Lock()
		if len(i.waiters) == 1 {
			i.mu.Unlock()
			break
		}
		i.mu.Unlock()
		time.Sleep(time.Millisecond)
	}
	cancel()
	<-waitReturned
	assert.Len(t, i.entries, 0)
	assert.Len(t, i.waiters, 0)
}

func TestRaceWakeAndContextCompletion(t *testing.T) {
	i := NewInstance()
	i.SetMaxEntries(1)
	eh0 := i.WaitDefault(context.Background(), entry(0))
	ctx, cancel := context.WithCancel(context.Background())
	waitReturned := make(chan struct{})
	go func() {
		eh1 := i.WaitDefault(ctx, entry(1))
		if eh1 != nil {
			eh1.Forget()
		}
		close(waitReturned)
	}()
	go cancel()
	go eh0.Forget()
	<-waitReturned
	cancel()
	eh0.Forget()
	i.mu.Lock()
	assert.Len(t, i.entries, 0)
	assert.Len(t, i.waiters, 0)
	i.mu.Unlock()
}
