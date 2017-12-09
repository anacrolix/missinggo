package reqctx

import (
	"context"
	"net/http"

	"github.com/anacrolix/missinggo/futures"
)

func NewApp() *App {
	ret := &App{}
	ret.contextKey = ret
	return ret
}

type App struct {
	contextKey interface{}
}

func (app App) Middleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, app.Install(r))
		})
	}
}

func (app App) Install(r *http.Request) *http.Request {
	return SetNewValue(r, app.contextKey, &LazyValues{r: r})
}

func (app *App) LazyValues(ctx context.Context) *LazyValues {
	return ctx.Value(app.contextKey).(*LazyValues)
}

type LazyValues struct {
	values map[interface{}]*futures.F
	r      *http.Request
}

func (me *LazyValues) Get(val *requestContextValue) *futures.F {
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

func NewLazyValue(get func(r *http.Request) (interface{}, error)) *requestContextValue {
	val := &requestContextValue{
		get: get,
	}
	val.key = val
	return val
}

type requestContextValue struct {
	key interface{}
	get func(r *http.Request) (interface{}, error)
}
