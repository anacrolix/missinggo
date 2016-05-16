package uniform

import (
	"io"
	"os"
)

type OSFileProvider struct{}

var _ Provider = &OSFileProvider{}

func (me *OSFileProvider) NewResource(filePath string) (r Resource, err error) {
	return &OSFileResource{filePath}, nil
}

type OSFileResource struct {
	path string
}

var _ Resource = &OSFileResource{}

func (me *OSFileResource) Get() (ret io.ReadCloser, err error) {
	return os.Open(me.path)
}

func (me *OSFileResource) Put(r io.Reader) (err error) {
	f, err := os.OpenFile(me.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return
}

func (me *OSFileResource) ReadAt(b []byte, off int64) (n int, err error) {
	f, err := os.Open(me.path)
	if err != nil {
		return
	}
	defer f.Close()
	return f.ReadAt(b, off)
}

func (me *OSFileResource) WriteAt(b []byte, off int64) (n int, err error) {
	f, err := os.OpenFile(me.path, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return
	}
	defer f.Close()
	return f.WriteAt(b, off)
}

func (me *OSFileResource) Stat() (fi os.FileInfo, err error) {
	return os.Stat(me.path)
}

func (me *OSFileResource) Delete() error {
	return os.Remove(me.path)
}
