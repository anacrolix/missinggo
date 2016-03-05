package httptoo

import (
	"crypto/tls"
	"net/http"
)

// Returns the http.Client's TLS Config, traversing and generating any
// defaults along the way to get it.
func ClientTLSConfig(cl *http.Client) *tls.Config {
	rt := cl.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	tr := rt.(*http.Transport)
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}
	ret := tr.TLSClientConfig
	return ret
}
