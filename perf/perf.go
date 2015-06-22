package perf

import (
	"expvar"
	"strconv"
	"sync"
	"time"
)

var (
	em = expvar.NewMap("perfBuckets")
	mu sync.RWMutex
)

type Timer struct {
	started time.Time
}

func NewTimer() Timer {
	return Timer{time.Now()}
}

// var es = [...]string{"-9", "-8", "-7", "-6", "-5", "-4", "-3", "-2", "-1", "0", "1", "2"}

func bucket(d time.Duration) string {
	e := int64(-9)
	for d != 0 {
		d /= 10
		e++
	}
	return strconv.FormatInt(e, 10)
	// return es[e+9]
}

func (t *Timer) Stop(desc string) string {
	d := time.Since(t.started)
	mu.RLock()
	_m := em.Get(desc)
	mu.RUnlock()
	var m *expvar.Map
	if _m == nil {
		mu.Lock()
		_m = em.Get(desc)
		if _m == nil {
			m = new(expvar.Map)
			m.Init()
			em.Set(desc, m)
		}
		mu.Unlock()
	} else {
		m = _m.(*expvar.Map)
	}
	b := bucket(d)
	m.Add(b, 1)
	return b
}
