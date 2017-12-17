package missinggo

import (
	"net/http"
	"time"

	"github.com/anacrolix/missinggo/assert"
)

// A http.ResponseWriter that tracks the status of the response. The status
// code, and number of bytes written for example.
type StatusResponseWriter struct {
	http.ResponseWriter
	Code         int
	BytesWritten int64
	Started      time.Time
	Ttfb         time.Duration
}

func (me *StatusResponseWriter) Write(b []byte) (n int, err error) {
	if me.BytesWritten == 0 && len(b) > 0 {
		assert.False(me.Started.IsZero())
		me.Ttfb = time.Since(me.Started)
	}
	n, err = me.ResponseWriter.Write(b)
	me.BytesWritten += int64(n)
	if me.Code == 0 {
		me.Code = 200
	}
	return
}

func (me *StatusResponseWriter) WriteHeader(code int) {
	me.ResponseWriter.WriteHeader(code)
	me.Code = code
}
