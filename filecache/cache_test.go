package filecache

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bradfitz/iter"

	"github.com/anacrolix/missinggo/v2"

	"github.com/go-quicktest/qt"
)

func TestCache(t *testing.T) {
	td, err := ioutil.TempDir("", "gotest")
	qt.Assert(t, qt.IsNil(err))
	defer os.RemoveAll(td)

	c, err := NewCache(filepath.Join(td, "cache"))
	qt.Assert(t, qt.IsNil(err))
	qt.Check(t, qt.Equals(c.Info(), CacheInfo{
		Filled:   0,
		Capacity: -1,
		NumItems: 0,
	}))

	c.WalkItems(func(i ItemInfo) {})

	_, err = c.OpenFile("/", os.O_CREATE)
	qt.Check(t, qt.IsNotNil(err))

	_, err = c.OpenFile("", os.O_CREATE)
	qt.Check(t, qt.IsNotNil(err))

	c.WalkItems(func(i ItemInfo) {})

	qt.Assert(t, qt.Equals(c.Info(), CacheInfo{
		Filled:   0,
		Capacity: -1,
		NumItems: 0,
	}))

	_, err = c.OpenFile("notexist", 0)
	qt.Check(t, qt.IsTrue(os.IsNotExist(err)), qt.Commentf("%v", err))

	_, err = c.OpenFile("/notexist", 0)
	qt.Check(t, qt.IsTrue(os.IsNotExist(err)), qt.Commentf("%v", err))

	_, err = c.OpenFile("/dir/notexist", 0)
	qt.Check(t, qt.IsTrue(os.IsNotExist(err)), qt.Commentf("%v", err))

	f, err := c.OpenFile("dir/blah", os.O_CREATE)
	qt.Assert(t, qt.IsNil(err))
	defer f.Close()
	qt.Assert(t, qt.Equals(c.Info(), CacheInfo{
		Filled:   0,
		Capacity: -1,
		NumItems: 1,
	}))

	c.WalkItems(func(i ItemInfo) {})

	qt.Check(t, qt.IsTrue(missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("cache/dir/blah")))))
	qt.Check(t, qt.IsTrue(missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("cache/dir/")))))
	qt.Check(t, qt.Equals(c.Info().NumItems, 1))

	_, err = f.ReadAt(nil, 0)
	qt.Check(t, qt.Not(qt.Equals(err, io.EOF)))
	f.Close()

	qt.Assert(t, qt.IsNil(c.Remove("dir/blah")))
	qt.Check(t, qt.IsFalse(missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("cache/dir/blah")))))
	qt.Check(t, qt.IsFalse(missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("cache/dir/")))))

	a, err := c.OpenFile("/a", os.O_CREATE|os.O_WRONLY)
	defer a.Close()
	qt.Assert(t, qt.IsNil(err))
	b, err := c.OpenFile("b", os.O_CREATE|os.O_WRONLY)
	defer b.Close()
	qt.Assert(t, qt.IsNil(err))
	c.mu.Lock()
	qt.Check(t, qt.IsFalse(c.pathInfo("a").Accessed.After(c.pathInfo("b").Accessed)))
	c.mu.Unlock()
	n, err := a.WriteAt([]byte("hello"), 0)
	qt.Check(t, qt.IsNil(err))
	qt.Check(t, qt.Equals(n, 5))
	qt.Check(t, qt.Equals(c.Info(), CacheInfo{
		Filled:   5,
		Capacity: -1,
		NumItems: 2,
	}))
	qt.Check(t, qt.IsFalse(c.pathInfo("b").Accessed.After(c.pathInfo("a").Accessed)))

	// Reopen a, to check that the info values remain correct.
	qt.Check(t, qt.IsNil(a.Close()))
	a, err = c.OpenFile("a", 0)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(c.Info(), CacheInfo{
		Filled:   5,
		Capacity: -1,
		NumItems: 2,
	}))

	c.SetCapacity(5)
	qt.Assert(t, qt.Equals(c.Info(), CacheInfo{
		Filled:   5,
		Capacity: 5,
		NumItems: 2,
	}))

	n, err = a.WriteAt([]byte(" world"), 5)
	qt.Check(t, qt.IsNotNil(err))
	n, err = b.WriteAt([]byte("boom!"), 0)
	// "a" and "b" have been evicted.
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(n, 5))
	qt.Assert(t, qt.Equals(c.Info(), CacheInfo{
		Filled:   5,
		Capacity: 5,
		NumItems: 1,
	}))
}

func TestSanitizePath(t *testing.T) {
	qt.Check(t, qt.Equals(sanitizePath("////"), ""))
	qt.Check(t, qt.Equals(sanitizePath("/../.."), ""))
	qt.Check(t, qt.Equals(sanitizePath("/a//b/.."), "a"))
	qt.Check(t, qt.Equals(sanitizePath("../a"), "a"))
	qt.Check(t, qt.Equals(sanitizePath("./a"), "a"))
}

func BenchmarkCacheOpenFile(t *testing.B) {
	td, err := ioutil.TempDir("", "")
	qt.Assert(t, qt.IsNil(err))
	defer os.RemoveAll(td)
	c, err := NewCache(td)
	for range iter.N(t.N) {
		func() {
			f, err := c.OpenFile("a", os.O_CREATE|os.O_RDWR)
			qt.Assert(t, qt.IsNil(err))
			qt.Check(t, qt.IsNil(f.Close()))
		}()
	}
}

func TestFileReadWrite(t *testing.T) {
	td, err := ioutil.TempDir("", "")
	qt.Assert(t, qt.IsNil(err))
	defer os.RemoveAll(td)

	c, err := NewCache(td)
	qt.Assert(t, qt.IsNil(err))

	a, err := c.OpenFile("a", os.O_CREATE|os.O_EXCL|os.O_RDWR)
	qt.Assert(t, qt.IsNil(err))
	defer a.Close()

	for off, c := range []byte("herp") {
		n, err := a.WriteAt([]byte{c}, int64(off))
		qt.Check(t, qt.IsNil(err))
		qt.Assert(t, qt.Equals(n, 1))
	}
	for off, c := range []byte("herp") {
		var b [1]byte
		n, err := a.ReadAt(b[:], int64(off))
		qt.Assert(t, qt.Equals(n, 1))
		qt.Assert(t, qt.IsNil(err))
		qt.Check(t, qt.DeepEquals(b[:], []byte{c}))
	}

}
