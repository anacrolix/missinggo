package filecache

import (
	"errors"
	"math"
	"os"
	"sync"
	"time"
)

type File struct {
	mu   sync.Mutex
	c    *Cache
	path string
	f    *os.File
}

func (me *File) Seek(offset int64, whence int) (ret int64, err error) {
	me.mu.Lock()
	defer me.mu.Unlock()
	return me.f.Seek(offset, whence)
}

func (me *File) maxWrite() (max int64, err error) {
	if me.c.capacity < 0 {
		max = math.MaxInt64
		return
	}
	pos, err := me.Seek(0, os.SEEK_CUR)
	if err != nil {
		return
	}
	max = me.c.capacity - pos
	return
}

var (
	ErrFileTooLarge    = errors.New("file too large for cache")
	ErrFileDisappeared = errors.New("file disappeared")
)

func (me *File) Write(b []byte) (n int, err error) {
	me.c.mu.Lock()
	mw, err := me.maxWrite()
	me.c.mu.Unlock()
	if err != nil {
		return
	}
	tooLarge := false
	if int64(len(b)) > mw {
		b = b[:mw]
		tooLarge = true
	}
	n, err = me.f.Write(b)
	me.c.mu.Lock()
	me.c.updateItem(me.path, time.Now())
	me.c.trimToCapacity()
	_, ok := me.c.paths[me.path]
	me.c.mu.Unlock()
	if !ok {
		n = 0
		err = ErrFileDisappeared
		me.f.Close()
		return
	}
	if tooLarge && err == nil && n == len(b) {
		err = ErrFileTooLarge
	}
	return
}

func (me *File) Close() (err error) {
	err = me.f.Close()
	if err == nil {
		go me.c.ProbeItem(me.path)
	}
	return
}

func (me *File) Stat() (os.FileInfo, error) {
	return me.f.Stat()
}

func (me *File) Read(b []byte) (int, error) {
	return me.f.Read(b)
}
