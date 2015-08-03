package missinggo

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

const debug = false

type Wolf struct {
	cries uint64
}

func (me *Wolf) CryHeard() bool {
	n := atomic.AddUint64(&me.cries, 1)
	return n&(n-1) == 0
}

var (
	mu     sync.Mutex
	wolves map[uintptr]*Wolf
)

func CryHeard() bool {
	pc, file, line, ok := runtime.Caller(1)
	if debug {
		log.Println(pc, file, line, ok)
	}
	if !ok {
		return true
	}
	mu.Lock()
	if wolves == nil {
		wolves = make(map[uintptr]*Wolf)
	}
	w, ok := wolves[pc]
	if !ok {
		w = new(Wolf)
		wolves[pc] = w
	}
	mu.Unlock()
	return w.CryHeard()
}
