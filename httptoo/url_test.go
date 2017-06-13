package httptoo

import (
	"net/url"

	"github.com/stretchr/testify/assert"

	"testing"
)

func TestAppendURL(t *testing.T) {
	assert.EqualValues(t, "http://localhost:8080/trailing/slash/", AppendURL(
		&url.URL{Scheme: "http", Host: "localhost:8080"},
		&url.URL{Path: "/trailing/slash/"},
	).String())
	assert.EqualValues(t, "ws://localhost:8080/events?ih=harpdarp", AppendURL(
		&url.URL{Scheme: "http", Host: "localhost:8080"},
		&url.URL{Scheme: "ws", Path: "/events", RawQuery: "ih=harpdarp"},
	).String())
}
