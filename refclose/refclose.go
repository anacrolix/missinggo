package refclose

import (
	"runtime/pprof"
	"sync"
)

var profile = pprof.NewProfile("refs")

type RefPool struct {
	mu sync.Mutex
	rs map[interface{}]*Resource
}

type Closer func()

func (me *RefPool) inc(key interface{}) {
	me.mu.Lock()
	defer me.mu.Unlock()
	r := me.rs[key]
	if r == nil {
		r = new(Resource)
		if me.rs == nil {
			me.rs = make(map[interface{}]*Resource)
		}
		me.rs[key] = r
	}
	r.numRefs++
}

func (me *RefPool) dec(key interface{}) {
	me.mu.Lock()
	defer me.mu.Unlock()
	r := me.rs[key]
	r.numRefs--
	if r.numRefs > 0 {
		return
	}
	if r.numRefs < 0 {
		panic(r.numRefs)
	}
	r.closer()
	delete(me.rs, key)
}

type Resource struct {
	closer  Closer
	numRefs int
}

func (me *RefPool) NewRef(key interface{}) (ret *Ref) {
	me.inc(key)
	ret = &Ref{
		pool: me,
		key:  key,
	}
	profile.Add(ret, 0)
	return
}

type Ref struct {
	mu     sync.Mutex
	pool   *RefPool
	key    interface{}
	closed bool
}

func (me *Ref) SetCloser(closer Closer) {
	me.pool.mu.Lock()
	defer me.pool.mu.Unlock()
	me.pool.rs[me.key].closer = closer
}

func (me *Ref) Release() {
	me.mu.Lock()
	defer me.mu.Unlock()
	if me.closed {
		panic("already closed ref")
	}
	profile.Remove(me)
	me.pool.dec(me.key)
}
