package gcclose

import (
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/bradfitz/iter"
)

type refTest struct {
	obj    interface{}
	t      *testing.T
	closed chan struct{}
}

func (me *refTest) closer() {
	close(me.closed)
}

func (me *refTest) failIfClosed() {
	select {
	case <-me.closed:
		me.t.FailNow()
	default:
	}
}

func (me *refTest) requireClosed() {
	select {
	case <-me.closed:
	case <-time.After(time.Second):
		panic("not closed")
	}
}

func (me *refTest) newRef() *Ref {
	return NewRef(me.obj, me.closer)
}

func (me refTest) run() {
	me.closed = make(chan struct{})
	r0 := me.newRef()
	for range iter.N(10000) {
		r1 := me.newRef()
		r0 = r1
		log.Print(r0.w.NumRefs)
		me.failIfClosed()
	}
	runtime.GC()
	me.requireClosed()
}

func Test(t *testing.T) {
	refTest{
		obj: 3,
		t:   t,
	}.run()
}
