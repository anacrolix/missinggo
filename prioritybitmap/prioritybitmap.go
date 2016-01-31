package prioritybitmap

import (
	"github.com/anacrolix/missinggo/orderedmap"
)

type PriorityBitmap struct {
	om         *orderedmap.OrderedMap
	priorities map[int]int
}

func New() *PriorityBitmap {
	return &PriorityBitmap{
		om:         orderedmap.New(),
		priorities: make(map[int]int),
	}
}

func (me *PriorityBitmap) delete(key int) {
	p, ok := me.priorities[key]
	if !ok {
		return
	}
	keys := me.om.Get(p).(map[int]struct{})
	delete(keys, key)
}

func (me *PriorityBitmap) getMake(priority int) map[int]struct{} {
	ret, ok := me.om.GetOk(priority)
	if !ok {
		ret = make(map[int]struct{})
		me.om.Set(priority, ret)
	}
	return ret.(map[int]struct{})
}

func (me *PriorityBitmap) Set(key int, priority int) {
	me.delete(key)
	me.getMake(priority)[key] = struct{}{}
	me.priorities[key] = priority
}
