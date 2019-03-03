package cache

import (
	"fmt"
	"log"
	"sync"

	humanize "github.com/dustin/go-humanize"
)

type Key = string

type Cache struct {
	mu     sync.Mutex
	filled int64
	Policy Policy
	Items  map[Key]ItemMeta
}

type ItemMeta struct {
	Size     int64
	CanEvict bool
	Usage
}

type Item struct {
	Key
	ItemMeta
}

func (me *Cache) Remove(k Key) {
	me.mu.Lock()
	i := me.Items[k]
	me.filled -= i.Size
	delete(me.Items, k)
	me.Policy.Forget(k)
	me.mu.Unlock()
}

func (me *Cache) Update(i Item) {
	me.mu.Lock()
	m := me.Items[i.Key]
	me.filled -= m.Size
	me.filled += i.Size
	if me.Items == nil {
		me.Items = make(map[Key]ItemMeta)
	}
	me.Items[i.Key] = i.ItemMeta
	if i.CanEvict {
		me.Policy.Update(i.Key, i.Usage)
	} else {
		me.Policy.Forget(i.Key)
	}
	me.mu.Unlock()
}

func (me *Cache) logState() {
	log.Print(me)
}

func (me *Cache) String() string {
	me.mu.Lock()
	defer me.mu.Unlock()
	return fmt.Sprintf(
		"%p: %d items, %v bytes used, lru: %s",
		me, len(me.Items), humanize.Bytes(uint64(me.filled)),
		func() string {
			k, ok := me.Policy.Candidate()
			if ok {
				i := me.Items[k]
				return fmt.Sprintf(
					"%q (%v: %v)",
					k, humanize.Bytes(uint64(i.Size)), i.Usage)
			}
			return "none"
		}(),
	)
}

func (me *Cache) Used() int64 {
	return me.filled
}

func (me *Cache) NumItems() int {
	return len(me.Items)
}

func (me *Cache) Clear() {
	for k := range me.Items {
		delete(me.Items, k)
		me.Policy.Forget(k)
	}
	me.Items = nil
	me.filled = 0
}

func (me *Cache) Filled() int64 {
	me.mu.Lock()
	defer me.mu.Unlock()
	return me.filled
}

func (me *Cache) Candidate() (Item, bool) {
	me.mu.Lock()
	defer me.mu.Unlock()
	k, ok := me.Policy.Candidate()
	return Item{
		Key:      k,
		ItemMeta: me.Items[k],
	}, ok
}
