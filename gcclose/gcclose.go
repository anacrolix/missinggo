// Package gcclose allows GC-triggered closing of objects.
package gcclose

import (
	"log"
	"runtime"
	"sync"

	"github.com/anacrolix/missinggo"
)

var (
	Mu       sync.Mutex
	wrappers = map[interface{}]*Wrapper{}
)

// Wrapper manages the lifetime of the contained object. Refs hold pointers to
// this.
type Wrapper struct {
	Object  interface{}
	Closer  func()
	NumRefs int
	Closed  missinggo.Event
}

// Ref decrements its object's reference count when garbage collected. When
// the reference count reaches zero, the object is released using the function
// provided when the first ref was created.
type Ref struct {
	w *Wrapper
}

// Returns the Ref's Wrapper.
func (r *Ref) Wrapper() *Wrapper {
	return r.w
}

func newWrapper(obj interface{}, closer func()) *Wrapper {
	return &Wrapper{
		Object: obj,
		Closer: closer,
	}
}

// Returns a new reference to obj, creating a wrapper if one does not exist.
// This must be called with Mu locked, to prevent multiple wrappers for the
// same object.
func NewRef(obj interface{}, closer func()) *Ref {
	w, ok := wrappers[obj]
	if !ok {
		w = newWrapper(obj, closer)
		wrappers[obj] = w
	}
	w.NumRefs++
	r := &Ref{w}
	runtime.SetFinalizer(r, refFinalizer)
	return r
}

func refFinalizer(r *Ref) {
	Mu.Lock()
	defer Mu.Unlock()
	r.w.NumRefs--
	if r.w.NumRefs != 0 {
		return
	}
	log.Printf("closing object with no refs: %s", r.w.Object)
	r.w.Closer()
	r.w.Closed.Set()
	delete(wrappers, r.w)
}
