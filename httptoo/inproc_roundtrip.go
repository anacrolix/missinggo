package httptoo

import (
	"io"
	"net/http"

	"github.com/anacrolix/missinggo"
)

type responseWriter struct {
	r             http.Response
	headerWritten missinggo.Event
	bodyWriter    io.Writer
}

func (me *responseWriter) Header() http.Header {
	if me.r.Header == nil {
		me.r.Header = make(http.Header)
	}
	return me.r.Header
}

func (me *responseWriter) Write(b []byte) (int, error) {
	if !me.headerWritten.IsSet() {
		me.WriteHeader(200)
	}
	return me.bodyWriter.Write(b)
}

func (me *responseWriter) WriteHeader(status int) {
	if me.headerWritten.IsSet() {
		return
	}
	me.r.StatusCode = status
	me.r.Body, me.bodyWriter = io.Pipe()
	me.headerWritten.Set()
}

func RoundTripHandler(req *http.Request, h http.Handler) (*http.Response, error) {
	rw := responseWriter{}
	go h.ServeHTTP(&rw, req)
	rw.headerWritten.Wait()
	return &rw.r, nil
}
