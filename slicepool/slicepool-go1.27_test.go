//go:build go1.27

package slicepool

import "testing"

func TestGetMinCapMethod(t *testing.T) {
	var p Pool[byte]
	b := p.GetMinCap(uint16(64))
	if cap(*b) < 64 || len(*b) != 0 {
		t.Fatalf("got len %v cap %v", len(*b), cap(*b))
	}
	*b = append(*b, 1, 2, 3)
	p.Put(b)
	// Different integer type for the cap, to exercise the method's own type parameter.
	b = p.GetMinCap(int64(8))
	if cap(*b) < 8 || len(*b) != 0 {
		t.Fatalf("got len %v cap %v", len(*b), cap(*b))
	}
}
