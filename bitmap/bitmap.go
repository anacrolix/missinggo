package bitmap

import "github.com/RoaringBitmap/roaring"

type Bitmap struct {
	rb *roaring.RoaringBitmap
}

func New() *Bitmap {
	return &Bitmap{roaring.NewRoaringBitmap()}
}

func (me *Bitmap) Iter() *Iter {
	if me == nil {
		return nil
	}
	return &Iter{me.rb.Iterator()}
}

func (me *Bitmap) Add(i int) {
	me.rb.AddInt(i)
}

func (me *Bitmap) Remove(i int) {
	if me == nil {
		return
	}
	me.rb.Remove(uint32(i))
}

func (me *Bitmap) Contains(i int) bool {
	if me == nil {
		return false
	}
	return me.rb.Contains(uint32(i))
}

type Iter struct {
	ii roaring.IntIterable
}

func (me *Iter) Next() bool {
	if me == nil {
		return false
	}
	return me.ii.HasNext()
}

func (me *Iter) Value() int {
	return int(me.ii.Next())
}
