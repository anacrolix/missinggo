package reqctx

import (
	"context"
	"net/http"

	"github.com/anacrolix/missinggo/futures"
)

var lazyValuesContextKey = new(byte)

func WithLazyMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = WithLazy(r)
			h.ServeHTTP(w, r)
		})
	}
}

func WithLazy(r *http.Request) *http.Request {
	if r.Context().Value(lazyValuesContextKey) == nil {
		r = r.WithContext(context.WithValue(r.Context(), lazyValuesContextKey, &LazyValues{r: r}))
	}
	return r
}

func GetLazyValues(ctx context.Context) *LazyValues {
	return ctx.Value(lazyValuesContextKey).(*LazyValues)
}

type LazyValues struct {
	values map[interface{}]*futures.F
	r      *http.Request
}

func (me *LazyValues) Get(val *lazyValue) *futures.F {
	f := me.values[val.key]
	if f != nil {
		return f
	}
	f = futures.Start(func() (interface{}, error) {
		return val.get(me.r)
	})
	if me.values == nil {
		me.values = make(map[interface{}]*futures.F)
	}
	me.values[val.key] = f
	return f
}

func NewLazyValue(get func(r *http.Request) (interface{}, error)) *lazyValue {
	val := &lazyValue{
		get: get,
	}
	val.key = val
	return val
}

type lazyValue struct {
	key interface{}
	get func(r *http.Request) (interface{}, error)
}

func (me *lazyValue) Get(r *http.Request) *futures.F {
	return me.GetContext(r.Context())
}

func (me *lazyValue) GetContext(ctx context.Context) *futures.F {
	return GetLazyValues(ctx).Get(me)
}

func (me *lazyValue) Prefetch(r *http.Request) {
	me.Get(r)
}

func (me *lazyValue) PrefetchMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		me.Prefetch(r)
		h.ServeHTTP(w, r)
	})
}
