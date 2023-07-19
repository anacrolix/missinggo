package httpmux

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	"go.opencensus.io/trace"
)

type pathParamContextKeyType struct{}

var pathParamContextKey pathParamContextKeyType

type Mux struct {
	handlers []Handler
}

func New() *Mux {
	return new(Mux)
}

type Handler struct {
	path        *regexp.Regexp
	userHandler http.Handler
}

func (h Handler) Pattern() string {
	return h.path.String()
}

func (mux *Mux) GetHandler(r *http.Request) *Handler {
	matches := mux.matchingHandlers(r)
	if len(matches) == 0 {
		return nil
	}
	return &matches[0].Handler
}

func (me *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	matches := me.matchingHandlers(r)
	if len(matches) == 0 {
		http.NotFound(w, r)
		return
	}
	m := matches[0]
	ctx := context.WithValue(r.Context(), pathParamContextKey, &PathParams{m})
	ctx, span := trace.StartSpan(ctx, m.Handler.path.String(), trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()
	r = r.WithContext(ctx)
	defer func() {
		r := recover()
		if r == http.ErrAbortHandler {
			panic(r)
		}
		if r == nil {
			return
		}
		panic(fmt.Sprintf("while handling %q: %s", m.Handler.path.String(), r))
	}()
	m.Handler.userHandler.ServeHTTP(w, r)
}

type match struct {
	Handler    Handler
	submatches []string
}

func (me *Mux) matchingHandlers(r *http.Request) (ret []match) {
	for _, h := range me.handlers {
		subs := h.path.FindStringSubmatch(r.URL.Path)
		if subs == nil {
			continue
		}
		ret = append(ret, match{h, subs})
	}
	return
}

func (me *Mux) distinctHandlerRegexp(r *regexp.Regexp) bool {
	for _, h := range me.handlers {
		if h.path.String() == r.String() {
			return false
		}
	}
	return true
}

func (me *Mux) Handle(path string, h http.Handler) {
	expr := "^" + path
	if !strings.HasSuffix(expr, "$") {
		expr += "$"
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		panic(err)
	}
	if !me.distinctHandlerRegexp(re) {
		panic(fmt.Sprintf("path %q is not distinct", path))
	}
	me.handlers = append(me.handlers, Handler{re, h})
}

func (me *Mux) HandleFunc(path string, hf func(http.ResponseWriter, *http.Request)) {
	me.Handle(path, http.HandlerFunc(hf))
}

func Path(parts ...string) string {
	return path.Join(parts...)
}

type PathParams struct {
	match match
}

func (me *PathParams) ByName(name string) string {
	for i, sn := range me.match.Handler.path.SubexpNames()[1:] {
		if sn == name {
			return me.match.submatches[i+1]
		}
	}
	return ""
}

func RequestPathParams(r *http.Request) *PathParams {
	ctx := r.Context()
	return ctx.Value(pathParamContextKey).(*PathParams)
}

func PathRegexpParam(name string, re string) string {
	return fmt.Sprintf("(?P<%s>%s)", name, re)
}

func Param(name string) string {
	return fmt.Sprintf("(?P<%s>[^/]+)", name)
}

func RestParam(name string) string {
	return fmt.Sprintf("(?P<%s>.*)$", name)
}

func NonEmptyRestParam(name string) string {
	return fmt.Sprintf("(?P<%s>.+)$", name)
}
