package reqctx

import (
	"context"
	"net/http"

	"github.com/anacrolix/missinggo/assert"
)

func NewValue() *contextValue {
	return &contextValue{new(byte)}
}

type contextValue struct {
	key interface{}
}

func (me contextValue) Get(ctx context.Context) interface{} {
	return ctx.Value(me.key)
}

func (me contextValue) SetRequestOnce(r *http.Request, val interface{}) *http.Request {
	assert.Nil(me.Get(r.Context()))
	return r.WithContext(context.WithValue(r.Context(), me.key, val))
}

func (me contextValue) SetMiddleware(val interface{}) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = me.SetRequestOnce(r, val)
			h.ServeHTTP(w, r)
		})
	}
}
