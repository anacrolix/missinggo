package filecache

import (
	"container/list"
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/anacrolix/missinggo"
)

type Cache struct {
	mu       sync.Mutex
	capacity int64
	filled   int64
	items    *list.List
	paths    map[string]*list.Element
	root     string
}

type CacheInfo struct {
	Capacity int64
	Filled   int64
	NumItems int
}

type ItemInfo struct {
	Accessed time.Time
	Size     int64
	Path     string
}

// Calls the function for every item known to the cache. The ItemInfo should
// not be modified.
func (me *Cache) WalkItems(cb func(*ItemInfo)) {
	me.mu.Lock()
	defer me.mu.Unlock()
	for e := me.items.Front(); e != nil; e = e.Next() {
		cb(e.Value.(*ItemInfo))
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

func (me *Cache) SetCapacity(capacity int64) {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.capacity = capacity
}

func NewCache(root string) (ret *Cache, err error) {
	if !filepath.IsAbs(root) {
		err = errors.New("root is not an absolute filepath")
		return
	}
	ret = &Cache{
		root:     root,
		capacity: -1,
	}
	ret.mu.Lock()
	go func() {
		defer ret.mu.Unlock()
		ret.rescan()
	}()
	return
}

// An empty return path is an error.
func sanitizePath(p string) (ret string) {
	if p == "" {
		return
	}
	ret = path.Clean("/" + p)
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

func (me *Cache) Remove(path string) (err error) {
	path = sanitizePath(path)
	err = me.remove(path)
	return
}

var (
	ErrBadPath = errors.New("bad path")
	ErrIsDir   = errors.New("is directory")
)

func (me *Cache) OpenFile(path string, flag int) (ret *File, err error) {
	path = sanitizePath(path)
	if path == "" {
		err = ErrIsDir
		return
	}
	f, err := os.OpenFile(me.realpath(path), flag, 0644)
	if flag&os.O_CREATE != 0 && os.IsNotExist(err) {
		os.MkdirAll(me.root, 0755)
		os.MkdirAll(filepath.Dir(me.realpath(path)), 0755)
		f, err = os.OpenFile(me.realpath(path), flag, 0644)
	}
	if err != nil {
		me.pruneEmptyDirs(path)
		return
	}
	ret = &File{
		c:    me,
		path: path,
		f:    f,
	}
	me.mu.Lock()
	go func() {
		defer me.mu.Unlock()
		me.statItem(path, time.Now())
	}()
	return
}

func (me *Cache) rescan() {
	me.filled = 0
	me.items = list.New()
	me.paths = make(map[string]*list.Element)
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
		me.statItem(path, time.Time{})
		return nil
	})
	if err != nil {
		panic(err)
	}
}

// Inserts the item into its sorted position. Items are usually being updated
// so start at the back.
func (me *Cache) insertItem(i *ItemInfo) *list.Element {
	for e := me.items.Back(); e != nil; e = e.Prev() {
		if i.Accessed.After(e.Value.(*ItemInfo).Accessed) {
			return me.items.InsertAfter(i, e)
		}
	}
	return me.items.PushFront(i)
}

func (me *Cache) removeInfo(path string) (ret *ItemInfo) {
	e, ok := me.paths[path]
	if !ok {
		return
	}
	ret = me.items.Remove(e).(*ItemInfo)
	me.filled -= ret.Size
	delete(me.paths, path)
	return
}

// Triggers the item for path to be updated. If access is non-zero, set the
// item's access time to that value, otherwise deduce it appropriately.
func (me *Cache) statItem(path string, access time.Time) {
	info := me.removeInfo(path)
	fi, err := os.Stat(me.realpath(path))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		panic(err)
	}
	if info == nil {
		info = &ItemInfo{
			Path: path,
		}
	}
	if !access.IsZero() {
		info.Accessed = access
	}
	if info.Accessed.IsZero() {
		access = missinggo.FileInfoAccessTime(fi)
	}
	info.Size = fi.Size()
	me.filled += info.Size
	me.paths[path] = me.insertItem(info)
}

func (me *Cache) realpath(path string) string {
	return filepath.Join(me.root, filepath.FromSlash(path))
}

func (me *Cache) TrimToCapacity() {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.trimToCapacity()
}

func (me *Cache) pruneEmptyDirs(path string) {
	pruneEmptyDirs(me.root, me.realpath(path))
}

func (me *Cache) remove(path string) (err error) {
	err = os.Remove(me.realpath(path))
	if os.IsNotExist(err) {
		err = nil
	}
	me.pruneEmptyDirs(path)
	me.removeInfo(path)
	return
}

func (me *Cache) trimToCapacity() {
	if me.capacity < 0 {
		return
	}
	for me.filled > me.capacity {
		item := me.items.Front().Value.(*ItemInfo)
		me.remove(item.Path)
	}
}
