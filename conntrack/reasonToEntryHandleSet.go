package conntrack

type reasonToEntryHandleSet struct {
	m map[reason]*entryHandleSet
}

func NewreasonToEntryHandleSet() reasonToEntryHandleSet {
	return reasonToEntryHandleSet{
		m: make(map[reason]*entryHandleSet),
	}
}

func (me reasonToEntryHandleSet) Contains(k reason) bool {
	_, ok := me.m[k]
	return ok
}

func (me reasonToEntryHandleSet) Get(k reason) *entryHandleSet {
	return me.m[k]
}

func (me reasonToEntryHandleSet) DeleteFromSet(k reason, sk *EntryHandle) bool {
	s, ok := me.m[k]
	if !ok {
		return true
	}
	s = s.Delete(sk)
	if s.Len() == 0 {
		delete(me.m, k)
		return true
	}
	me.m[k] = s
	return false
}

func (me reasonToEntryHandleSet) AddToSet(k reason, sk *EntryHandle) {
	s, ok := me.m[k]
	if ok {
		s = s.Add(sk)
	} else {
		s = NewentryHandleSet(sk)
	}
	me.m[k] = s
}

func (me reasonToEntryHandleSet) Len() int {
	return len(me.m)
}

func (me reasonToEntryHandleSet) Range(f func(reason, *entryHandleSet) bool) {
	for k, s := range me.m {
		if !f(k, s) {
			break
		}
	}
}
