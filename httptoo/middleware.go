package httptoo

import "net/http"

type Middleware func(http.Handler) http.Handler

func WrapHandler(middleware []Middleware, h http.Handler) (ret http.Handler) {
	ret = h
	for i := range middleware {
		ret = middleware[len(middleware)-1-i](ret)
	}
	return
}

func WrapHandlerFunc(middleware []Middleware, hf func(http.ResponseWriter, *http.Request)) (ret http.Handler) {
	return WrapHandler(middleware, http.HandlerFunc(hf))
}

func RunHandler(h http.Handler, w http.ResponseWriter, r *http.Request, middleware ...Middleware) {
	WrapHandler(middleware, h).ServeHTTP(w, r)
}

func RunHandlerFunc(h func(http.ResponseWriter, *http.Request), w http.ResponseWriter, r *http.Request, middleware ...Middleware) {
	WrapHandlerFunc(middleware, h).ServeHTTP(w, r)
}
