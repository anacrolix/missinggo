package missinggo

import (
	"net/url"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stretchr/testify/assert"
)

func TestURLOpaquePath(t *testing.T) {
	assert.Equal(t, "sqlite3://sqlite3.db", (&url.URL{Scheme: "sqlite3", Path: "sqlite3.db"}).String())
	u, err := url.Parse("sqlite3:sqlite3.db")
	assert.NoError(t, err)
	assert.Equal(t, "sqlite3.db", URLOpaquePath(u))
	assert.Equal(t, "sqlite3:sqlite3.db", (&url.URL{Scheme: "sqlite3", Opaque: "sqlite3.db"}).String())
	assert.Equal(t, "sqlite3:/sqlite3.db", (&url.URL{Scheme: "sqlite3", Opaque: "/sqlite3.db"}).String())
	u, err = url.Parse("sqlite3:/sqlite3.db")
	assert.NoError(t, err)
	assert.Equal(t, "/sqlite3.db", u.Path)
	assert.Equal(t, "/sqlite3.db", URLOpaquePath(u))
}

func testSchemePopping(t *testing.T, opaque string, expectedPath string) {
	searchDb := &url.URL{
		Scheme: "caterwaul",
		Opaque: "pebble:" + opaque,
	}
	c := qt.New(t)
	scheme, poppedUrlStr := PopScheme(searchDb)
	c.Check(scheme, qt.Equals, "caterwaul")
	poppedUrl, err := url.Parse(poppedUrlStr)
	c.Assert(err, qt.IsNil)
	scheme, poppedUrlStr = PopScheme(poppedUrl)
	c.Check(scheme, qt.Equals, "pebble")
	c.Check(poppedUrlStr, qt.Equals, expectedPath)
}

func TestSchemePopping(t *testing.T) {
	testSchemePopping(t, "caterwaul-pebble-search", "caterwaul-pebble-search")
	testSchemePopping(t, "/home/derp/cove/caterwaul-pebble-search", "/home/derp/cove/caterwaul-pebble-search")
	testSchemePopping(t, `C:\Users\derp\LocalData\`, `C:\Users\derp\LocalData\`)
}
