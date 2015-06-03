package filecache

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anacrolix/missinggo"
)

func TestCache(t *testing.T) {
	td, err := ioutil.TempDir("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	c, err := NewCache(filepath.Join(td, "cache"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.OpenFile("/", os.O_CREATE)
	assert.NotNil(t, err)
	_, err = c.OpenFile("", os.O_CREATE)
	assert.NotNil(t, err)
	require.Equal(t, 0, c.Info().NumItems)
	_, err = c.OpenFile("notexist", 0)
	assert.True(t, os.IsNotExist(err), err)
	_, err = c.OpenFile("/notexist", 0)
	assert.True(t, os.IsNotExist(err), err)
	_, err = c.OpenFile("/dir/notexist", 0)
	assert.True(t, os.IsNotExist(err), err)
	f, err := c.OpenFile("dir/blah", os.O_CREATE)
	require.NoError(t, err)
	defer f.Close()
	assert.True(t, missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("cache/dir/blah"))))
	assert.True(t, missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("cache/dir/"))))
	assert.Equal(t, 1, c.Info().NumItems)
	f.Remove()
	assert.False(t, missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("dir/blah"))))
	assert.False(t, missinggo.FilePathExists(filepath.Join(td, filepath.FromSlash("dir/"))))
	_, err = f.Read(nil)
	assert.NotEqual(t, io.EOF, err)
	a, err := c.OpenFile("/a", os.O_CREATE|os.O_WRONLY)
	defer a.Close()
	require.Nil(t, err)
	b, err := c.OpenFile("b", os.O_CREATE|os.O_WRONLY)
	defer b.Close()
	require.Nil(t, err)
	assert.True(t, c.paths["a"].Value.(ItemInfo).Accessed.Before(c.paths["b"].Value.(ItemInfo).Accessed))
	a.Write([]byte("hello"))
	assert.True(t, c.paths["b"].Value.(ItemInfo).Accessed.Before(c.paths["a"].Value.(ItemInfo).Accessed))
	c.SetCapacity(5)
	n, err := a.Write([]byte(" world"))
	assert.Equal(t, ErrFileTooLarge, err)
	assert.Equal(t, 0, n)
	fi, _ := a.Stat()
	assert.EqualValues(t, 5, fi.Size())
}

func TestSanitizePath(t *testing.T) {
	assert.Equal(t, "", sanitizePath("////"))
	assert.Equal(t, "", sanitizePath("/../.."))
	assert.Equal(t, "a", sanitizePath("/a//b/.."))
}
