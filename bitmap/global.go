package bitmap

import "github.com/RoaringBitmap/roaring"

func Sub(left, right Bitmap) Bitmap {
	return Bitmap{roaring.AndNot(left.rb, right.rb)}
}

func Flip(bm Bitmap, start, end int) Bitmap {
	return Bitmap{roaring.FlipInt(bm.rb, start, end)}
}
