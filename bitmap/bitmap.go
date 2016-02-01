package bitmap

import "github.com/RoaringBitmap/roaring"

// Bitmaps store the existence of values in [0,math.MaxUint32] more
// efficiently than []bool. The empty value starts with no bits set.
type Bitmap struct {
	inited bool
	rb     *roaring.RoaringBitmap
}

func (me *Bitmap) Iter() *Iter {
	if me.rb == nil {
		return nil
	}
	return &Iter{me.rb.Iterator()}
}

func (me *Bitmap) lazyInit() {
	if me.inited {
		return
	}
	me.rb = roaring.NewRoaringBitmap()
	me.inited = true
}

func (me *Bitmap) Add(i int) {
	me.lazyInit()
	me.rb.AddInt(i)
}

func (me *Bitmap) Remove(i int) {
	if me.rb == nil {
		return
	}
	me.rb.Remove(uint32(i))
}

func (me *Bitmap) Contains(i int) bool {
	if me.rb == nil {
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
