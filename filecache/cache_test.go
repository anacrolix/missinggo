package filecache

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bradfitz/iter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anacrolix/missinggo"
)

func TestCache(t *testing.T) {
	td, err := ioutil.TempDir("", "gotest")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	c, err := NewCache(filepath.Join(td, "cache"))
	require.NoError(t, err)
	assert.EqualValues(t, CacheInfo{
		Filled:   0,
		Capacity: -1,
		NumItems: 0,
	}, c.Info())

	c.WalkItems(func(i ItemInfo) {})

	_, err = c.OpenFile("/", os.O_CREATE)
	assert.NotNil(t, err)

	_, err = c.OpenFile("", os.O_CREATE)
	assert.NotNil(t, err)

	c.WalkItems(func(i ItemInfo) {})

	require.Equal(t, CacheInfo{
		Filled:   0,
		Capacity: -1,
		NumItems: 0,
	}, c.Info())

	_, err = c.OpenFile("notexist", 0)
	assert.True(t, os.IsNotExist(err), err)

	_, err = c.OpenFile("/notexist", 0)
	assert.True(t, os.IsNotExist(err), err)

	_, err = c.OpenFile("/dir/notexist", 0)
	assert.True(t, os.IsNotExist(err), err)

	f, err := c.OpenFile("dir/blah", os.O_CREATE)
	require.NoError(t, err)
	defer f.Close()
	require.Equal(t, CacheInfo{
		Filled:   0,
		Capacity: -1,
		NumItems: 1,
	}, c.Info())

	c.WalkItems(func(i ItemInfo) {})

	assert.True(t, missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("cache/dir/blah"))))
	assert.True(t, missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("cache/dir/"))))
	assert.Equal(t, 1, c.Info().NumItems)

	c.Remove("dir/blah")
	assert.False(t, missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("cache/dir/blah"))))
	assert.False(t, missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("cache/dir/"))))
	_, err = f.ReadAt(nil, 0)
	assert.NotEqual(t, io.EOF, err)

	a, err := c.OpenFile("/a", os.O_CREATE|os.O_WRONLY)
	defer a.Close()
	require.NoError(t, err)
	b, err := c.OpenFile("b", os.O_CREATE|os.O_WRONLY)
	defer b.Close()
	require.NoError(t, err)
	c.mu.Lock()
	assert.False(t, c.pathInfo("a").Accessed.After(c.pathInfo("b").Accessed))
	c.mu.Unlock()
	n, err := a.WriteAt([]byte("hello"), 0)
	assert.NoError(t, err)
	assert.EqualValues(t, 5, n)
	assert.EqualValues(t, CacheInfo{
		Filled:   5,
		Capacity: -1,
		NumItems: 2,
	}, c.Info())
	assert.False(t, c.pathInfo("b").Accessed.After(c.pathInfo("a").Accessed))

	// Reopen a, to check that the info values remain correct.
	assert.NoError(t, a.Close())
	a, err = c.OpenFile("a", 0)
	require.NoError(t, err)
	require.EqualValues(t, CacheInfo{
		Filled:   5,
		Capacity: -1,
		NumItems: 2,
	}, c.Info())

	c.SetCapacity(5)
	require.EqualValues(t, CacheInfo{
		Filled:   5,
		Capacity: 5,
		NumItems: 2,
	}, c.Info())

	n, err = a.WriteAt([]byte(" world"), 5)
	assert.Error(t, err)
	n, err = b.WriteAt([]byte("boom!"), 0)
	// "a" and "b" have been evicted.
	require.NoError(t, err)
	require.EqualValues(t, 5, n)
	require.EqualValues(t, CacheInfo{
		Filled:   5,
		Capacity: 5,
		NumItems: 1,
	}, c.Info())
}

func TestSanitizePath(t *testing.T) {
	assert.EqualValues(t, "", sanitizePath("////"))
	assert.EqualValues(t, "", sanitizePath("/../.."))
	assert.EqualValues(t, "a", sanitizePath("/a//b/.."))
	assert.EqualValues(t, "a", sanitizePath("../a"))
	assert.EqualValues(t, "a", sanitizePath("./a"))
}

func BenchmarkCacheOpenFile(t *testing.B) {
	td, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(td)
	c, err := NewCache(td)
	for range iter.N(t.N) {
		func() {
			f, err := c.OpenFile("a", os.O_CREATE|os.O_RDWR)
			require.NoError(t, err)
			assert.NoError(t, f.Close())
		}()
	}
}

func TestFileReadWrite(t *testing.T) {
	td, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	c, err := NewCache(td)
	require.NoError(t, err)

	a, err := c.OpenFile("a", os.O_CREATE|os.O_EXCL|os.O_RDWR)
	require.NoError(t, err)
	defer a.Close()

	for off, c := range []byte("herp") {
		n, err := a.WriteAt([]byte{c}, int64(off))
		assert.NoError(t, err)
		require.EqualValues(t, 1, n)
	}
	for off, c := range []byte("herp") {
		var b [1]byte
		n, err := a.ReadAt(b[:], int64(off))
		require.EqualValues(t, 1, n)
		require.NoError(t, err)
		assert.EqualValues(t, []byte{c}, b[:])
	}

}
