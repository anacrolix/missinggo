package missinggo

import (
	"net/url"
	"testing"

	"github.com/go-quicktest/qt"
)

func TestURLOpaquePath(t *testing.T) {
	qt.Check(t, qt.Equals((&url.URL{Scheme: "sqlite3", Path: "sqlite3.db"}).String(), "sqlite3://sqlite3.db"))
	u, err := url.Parse("sqlite3:sqlite3.db")
	qt.Check(t, qt.IsNil(err))
	qt.Check(t, qt.Equals(URLOpaquePath(u), "sqlite3.db"))
	qt.Check(t, qt.Equals((&url.URL{Scheme: "sqlite3", Opaque: "sqlite3.db"}).String(), "sqlite3:sqlite3.db"))
	qt.Check(t, qt.Equals((&url.URL{Scheme: "sqlite3", Opaque: "/sqlite3.db"}).String(), "sqlite3:/sqlite3.db"))
	u, err = url.Parse("sqlite3:/sqlite3.db")
	qt.Check(t, qt.IsNil(err))
	qt.Check(t, qt.Equals(u.Path, "/sqlite3.db"))
	qt.Check(t, qt.Equals(URLOpaquePath(u), "/sqlite3.db"))
}

func testSchemePopping(t *testing.T, opaque string, expectedPath string) {
	searchDb := &url.URL{
		Scheme: "caterwaul",
		Opaque: "pebble:" + opaque,
	}
	scheme, poppedUrlStr := PopScheme(searchDb)
	qt.Check(t, qt.Equals(scheme, "caterwaul"))
	poppedUrl, err := url.Parse(poppedUrlStr)
	qt.Assert(t, qt.IsNil(err))
	scheme, poppedUrlStr = PopScheme(poppedUrl)
	qt.Check(t, qt.Equals(scheme, "pebble"))
	qt.Check(t, qt.Equals(poppedUrlStr, expectedPath))
}

func TestSchemePopping(t *testing.T) {
	testSchemePopping(t, "caterwaul-pebble-search", "caterwaul-pebble-search")
	testSchemePopping(t, "/home/derp/cove/caterwaul-pebble-search", "/home/derp/cove/caterwaul-pebble-search")
	testSchemePopping(t, `C:\Users\derp\LocalData\`, `C:\Users\derp\LocalData\`)
}
