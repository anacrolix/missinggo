package httptoo

import (
	"net/url"
	"testing"

	"github.com/go-quicktest/qt"
)

func TestAppendURL(t *testing.T) {
	qt.Check(t, qt.Equals(AppendURL(
		&url.URL{Scheme: "http", Host: "localhost:8080"},
		&url.URL{Path: "/trailing/slash/"},
	).String(), "http://localhost:8080/trailing/slash/"))
	qt.Check(t, qt.Equals(AppendURL(
		&url.URL{Scheme: "http", Host: "localhost:8080"},
		&url.URL{Scheme: "ws", Path: "/events", RawQuery: "ih=harpdarp"},
	).String(), "ws://localhost:8080/events?ih=harpdarp"))
}
