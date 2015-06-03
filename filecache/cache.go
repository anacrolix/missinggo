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

func (me *Cache) WalkItems(cb func(ItemInfo)) {
	for e := me.items.Front(); e != nil; e = e.Next() {
		cb(e.Value.(ItemInfo))
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

func sanitizePath(p string) (ret string) {
	ret = path.Clean(p)
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
	if flag&os.O_CREATE != 0 {
		os.MkdirAll(me.root, 0755)
		os.MkdirAll(filepath.Dir(me.realpath(path)), 0755)
	}
	f, err := os.OpenFile(filepath.Join(me.root, path), flag, 0644)
	if err != nil {
		me.pruneEmptyDirs(path)
		return
	}
	fi, err := f.Stat()
	if err != nil {
		me.pruneEmptyDirs(path)
		return
	}
	if fi.IsDir() {
		err = ErrIsDir
		return
	}
	ret = &File{
		c:    me,
		path: path,
		f:    f,
	}
	me.AccessedItem(path)
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
		log.Print(path)
		me.probeItem(path)
		return nil
	})
	if err != nil {
		panic(err)
	}
}

// Inserts the item into its sorted position.
func (me *Cache) insertItem(i ItemInfo) *list.Element {
	for e := me.items.Front(); e != nil; e = e.Next() {
		if i.Accessed.Before(e.Value.(ItemInfo).Accessed) {
			return me.items.InsertBefore(i, e)
		}
	}
	return me.items.PushBack(i)
}

func (me *Cache) AccessedItem(path string) {
	me.UpdateItem(path, time.Now())
}

func (me *Cache) accessedItem(path string) {
	me.updateItem(path, time.Now())
}

func (me *Cache) UpdateItem(path string, access time.Time) {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.updateItem(path, access)
}

func (me *Cache) probeItem(path string) {
	me.updateItem(path, time.Time{})
}

func (me *Cache) ProbeItem(path string) {
	me.UpdateItem(path, time.Time{})
}

// Triggers the item for path to be updated. If access is non-zero, set the
// item's access time to that value, otherwise deduce it appropriately.
func (me *Cache) updateItem(path string, access time.Time) {
	// If the item is known, remove it.
	if e, ok := me.paths[path]; ok {
		v := me.items.Remove(e).(ItemInfo)
		me.filled -= v.Size
		if access.IsZero() {
			// Reuse the last access time.
			access = v.Accessed
		}
		delete(me.paths, path)
	}
	fi, err := os.Stat(me.realpath(path))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		panic(err)
	}
	if access.IsZero() {
		access = missinggo.FileInfoAccessTime(fi)
	}
	// Insert the item.
	ii := ItemInfo{
		access,
		fi.Size(),
		path,
	}
	me.filled += ii.Size
	me.paths[path] = me.insertItem(ii)
}

func (me *Cache) itemRealPath(info *ItemInfo) string {
	return filepath.Join(me.root, info.Path)
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
	me.probeItem(path)
	me.pruneEmptyDirs(path)
	return
}

func (me *Cache) trimToCapacity() {
	if me.capacity < 0 {
		return
	}
	for me.filled > me.capacity {
		item := me.items.Front().Value.(ItemInfo)
		me.remove(item.Path)
	}
}
