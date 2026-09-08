package httptoo

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-quicktest/qt"
)

const helloWorld = "hello, world\n"

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Length", strconv.FormatInt(int64(len(helloWorld)), 10))
	w.Write([]byte(helloWorld))
}

func requestResponse(h http.Handler, r *http.Request) (*http.Response, error) {
	s := httptest.NewServer(h)
	defer s.Close()
	return http.DefaultClient.Do(r)
}

func TestGzipHandler(t *testing.T) {
	rr := httptest.NewRecorder()
	helloWorldHandler(rr, nil)
	qt.Check(t, qt.Equals(rr.Body.String(), helloWorld))

	rr = httptest.NewRecorder()
	GzipHandler(http.HandlerFunc(helloWorldHandler)).ServeHTTP(rr, new(http.Request))
	qt.Check(t, qt.Equals(rr.Body.String(), helloWorld))

	rr = httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)
	qt.Assert(t, qt.IsNil(err))
	r.Header.Set("Accept-Encoding", "gzip")
	GzipHandler(http.HandlerFunc(helloWorldHandler)).ServeHTTP(rr, r)
	gr, err := gzip.NewReader(rr.Body)
	qt.Assert(t, qt.IsNil(err))
	defer gr.Close()
	b, err := ioutil.ReadAll(gr)
	qt.Assert(t, qt.IsNil(err))
	qt.Check(t, qt.Equals(string(b), helloWorld))

	s := httptest.NewServer(nil)
	s.Config.Handler = GzipHandler(http.HandlerFunc(helloWorldHandler))
	req, err := http.NewRequest("GET", s.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := http.DefaultClient.Do(req)
	qt.Assert(t, qt.IsNil(err))
	gr.Close()
	gr, err = gzip.NewReader(resp.Body)
	qt.Assert(t, qt.IsNil(err))
	defer gr.Close()
	b, err = ioutil.ReadAll(gr)
	qt.Assert(t, qt.IsNil(err))
	qt.Check(t, qt.Equals(string(b), helloWorld))
	qt.Check(t, qt.Equals(resp.Header.Get("Content-Type"), "text/plain; charset=utf-8"))
	qt.Check(t, qt.Equals(resp.Header.Get("Content-Encoding"), "gzip"))
}
