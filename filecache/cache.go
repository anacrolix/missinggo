package filecache

import (
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/anacrolix/missinggo/pproffd"
	"github.com/anacrolix/missinggo/resource"
)

const (
	dirPerm  = 0755
	filePerm = 0644
)

type Cache struct {
	root     string
	mu       sync.Mutex
	capacity int64
	filled   int64
	policy   Policy
	paths    map[key]ItemInfo
}

type CacheInfo struct {
	Capacity int64
	Filled   int64
	NumItems int
}

type ItemInfo struct {
	Accessed time.Time
	Size     int64
	Path     key
}

// Calls the function for every item known to the cache. The ItemInfo should
// not be modified.
func (me *Cache) WalkItems(cb func(ItemInfo)) {
	me.mu.Lock()
	defer me.mu.Unlock()
	for _, ii := range me.paths {
		cb(ii)
	}
}

func (me *Cache) Info() (ret CacheInfo) {
	me.mu.Lock()
	defer me.mu.Unlock()
	ret.Capacity = me.capacity
	ret.Filled = me.filled
	ret.NumItems = len(me.paths)
	return
}

// Setting a negative capacity means unlimited.
func (me *Cache) SetCapacity(capacity int64) {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.capacity = capacity
}

func NewCache(root string) (ret *Cache, err error) {
	root, err = filepath.Abs(root)
	ret = &Cache{
		root:     root,
		capacity: -1, // unlimited
	}
	ret.mu.Lock()
	go func() {
		defer ret.mu.Unlock()
		ret.rescan()
	}()
	return
}

// An empty return path is an error.
func sanitizePath(p string) (ret key) {
	if p == "" {
		return
	}
	ret = key(path.Clean("/" + p))
	if ret[0] == '/' {
		ret = ret[1:]
	}
	return
}

// Leaf is a descendent of root.
func pruneEmptyDirs(root string, leaf string) (err error) {
	rootInfo, err := os.Stat(root)
	if err != nil {
		return
	}
	for {
		var leafInfo os.FileInfo
		leafInfo, err = os.Stat(leaf)
		if os.IsNotExist(err) {
			goto parent
		}
		if err != nil {
			return
		}
		if !leafInfo.IsDir() {
			return
		}
		if os.SameFile(rootInfo, leafInfo) {
			return
		}
		if os.Remove(leaf) != nil {
			return
		}
	parent:
		leaf = filepath.Dir(leaf)
	}
}

func (me *Cache) Remove(path string) error {
	me.mu.Lock()
	defer me.mu.Unlock()
	return me.remove(sanitizePath(path))
}

var (
	ErrBadPath = errors.New("bad path")
	ErrIsDir   = errors.New("is directory")
)

func (me *Cache) StatFile(path string) (os.FileInfo, error) {
	return os.Stat(me.realpath(sanitizePath(path)))
}

func (me *Cache) OpenFile(path string, flag int) (ret *File, err error) {
	key := sanitizePath(path)
	if key == "" {
		err = ErrIsDir
		return
	}
	f, err := os.OpenFile(me.realpath(key), flag, filePerm)
	if flag&os.O_CREATE != 0 && os.IsNotExist(err) {
		// Ensure intermediate directories and try again.
		os.MkdirAll(filepath.Dir(me.realpath(key)), dirPerm)
		f, err = os.OpenFile(me.realpath(key), flag, filePerm)
		if err != nil {
			go me.pruneEmptyDirs(key)
		}
	}
	if err != nil {
		return
	}
	ret = &File{
		path: key,
		f:    pproffd.WrapOSFile(f),
		onRead: func(n int) {
			me.mu.Lock()
			defer me.mu.Unlock()
			me.updateItem(key, func(i *ItemInfo, ok bool) bool {
				i.Accessed = time.Now()
				return ok
			})
		},
		afterWrite: func(endOff int64) {
			me.mu.Lock()
			defer me.mu.Unlock()
			me.updateItem(key, func(i *ItemInfo, ok bool) bool {
				i.Accessed = time.Now()
				if endOff > i.Size {
					i.Size = endOff
				}
				return ok
			})
		},
	}
	accessed := time.Now()
	me.mu.Lock()
	go func() {
		defer me.mu.Unlock()
		me.updateItem(key, func(i *ItemInfo, ok bool) bool {
			if !ok {
				*i, ok = me.statKey(key)
			}
			i.Accessed = accessed
			return ok
		})
	}()
	return
}

func (me *Cache) rescan() {
	me.filled = 0
	me.policy = new(lru)
	me.paths = make(map[key]ItemInfo)
	err := filepath.Walk(me.root, func(path string, info os.FileInfo, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		path, err = filepath.Rel(me.root, path)
		if err != nil {
			log.Print(err)
			return nil
		}
		key := sanitizePath(path)
		me.updateItem(key, func(i *ItemInfo, ok bool) bool {
			if ok {
				panic("scanned duplicate items")
			}
			*i, ok = me.statKey(key)
			return ok
		})
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (me *Cache) statKey(k key) (i ItemInfo, ok bool) {
	fi, err := os.Stat(me.realpath(k))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		panic(err)
	}
	i.FromFileInfo(fi, k)
	ok = true
	return
}

func (me *Cache) updateItem(k key, u func(*ItemInfo, bool) bool) {
	ii, ok := me.paths[k]
	me.filled -= ii.Size
	if u(&ii, ok) {
		me.filled += ii.Size
		me.policy.Used(k, ii.Accessed)
		me.paths[k] = ii
	} else {
		me.policy.Forget(k)
		delete(me.paths, k)
	}
	me.trimToCapacity()
}

func (me *Cache) realpath(path key) string {
	return filepath.Join(me.root, filepath.FromSlash(string(path)))
}

func (me *Cache) TrimToCapacity() {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.trimToCapacity()
}

func (me *Cache) pruneEmptyDirs(path key) {
	pruneEmptyDirs(me.root, me.realpath(path))
}

func (me *Cache) remove(path key) error {
	err := os.Remove(me.realpath(path))
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return err
	}
	me.pruneEmptyDirs(path)
	me.updateItem(path, func(*ItemInfo, bool) bool {
		return false
	})
	return nil
}

func (me *Cache) trimToCapacity() {
	if me.capacity < 0 {
		return
	}
	for me.filled > me.capacity {
		me.remove(me.policy.Choose().(key))
	}
}

func (me *Cache) pathInfo(p string) ItemInfo {
	return me.paths[sanitizePath(p)]
}

func (me *Cache) Rename(from, to string) (err error) {
	_from := sanitizePath(from)
	_to := sanitizePath(to)
	me.mu.Lock()
	defer me.mu.Unlock()
	err = os.MkdirAll(filepath.Dir(me.realpath(_to)), dirPerm)
	if err != nil {
		return
	}
	err = os.Rename(me.realpath(_from), me.realpath(_to))
	if err != nil {
		return
	}
	me.updateItem(_from, func(i *ItemInfo, ok bool) bool {
		if ok {
			i.Path = _to
		} else {
			*i, ok = me.statKey(_to)
		}
		return ok
	})
	return
}

func (me *Cache) Stat(path string) (os.FileInfo, error) {
	return os.Stat(me.realpath(sanitizePath(path)))
}

func (me *Cache) AsResourceProvider() resource.Provider {
	return &uniformResourceProvider{me}
}
