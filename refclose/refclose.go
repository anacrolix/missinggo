package refclose

import "sync"

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

func (me *RefPool) dec(key interface{}, closer Closer) {
	me.mu.Lock()
	defer me.mu.Unlock()
	r := me.rs[key]
	if closer != nil && r.closer == nil {
		r.closer = closer
	}
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
	return &Ref{
		pool: me,
		key:  key,
	}
}

type Ref struct {
	mu     sync.Mutex
	pool   *RefPool
	key    interface{}
	closed bool
}

func (me *Ref) Abort() {
	me.Release(nil)
}

func (me *Ref) Release(closer Closer) {
	me.mu.Lock()
	defer me.mu.Unlock()
	if me.closed {
		panic("already closed ref")
	}
	me.pool.dec(me.key, closer)
}
