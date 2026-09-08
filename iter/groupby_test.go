package iter

import (
	"testing"

	"github.com/anacrolix/missinggo/slices"

	"github.com/go-quicktest/qt"
)

func TestGroupByKey(t *testing.T) {
	var ks []byte
	gb := GroupBy(StringIterator("AAAABBBCCDAABBB"), nil)
	for gb.Next() {
		ks = append(ks, gb.Value().(Group).Key().(byte))
	}
	t.Log(ks)
	qt.Assert(t, qt.Equals(string(ks), "ABCDAB"))
}

func TestGroupByList(t *testing.T) {
	var gs []string
	gb := GroupBy(StringIterator("AAAABBBCCD"), nil)
	for gb.Next() {
		i := gb.Value().(Iterator)
		var g string
		for i.Next() {
			g += string(i.Value().(byte))
		}
		gs = append(gs, g)
	}
	t.Log(gs)
}

func TestGroupByNiladicKey(t *testing.T) {
	const s = "AAAABBBCCD"
	gb := GroupBy(StringIterator(s), func(interface{}) interface{} { return nil })
	gb.Next()
	var ss []byte
	g := ToSlice(ToFunc(gb.Value().(Iterator)))
	slices.MakeInto(&ss, g)
	qt.Check(t, qt.Equals(string(ss), s))
}

func TestNilEqualsNil(t *testing.T) {
	qt.Check(t, qt.IsFalse(nil == uniqueKey))
}
