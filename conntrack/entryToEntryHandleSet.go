package conntrack

type entryToEntryHandleSet struct {
	m map[Entry]*entryHandleSet
}

func NewentryToEntryHandleSet() entryToEntryHandleSet {
	return entryToEntryHandleSet{
		m: make(map[Entry]*entryHandleSet),
	}
}

func (me entryToEntryHandleSet) Contains(k Entry) bool {
	_, ok := me.m[k]
	return ok
}

func (me entryToEntryHandleSet) Get(k Entry) *entryHandleSet {
	return me.m[k]
}

func (me entryToEntryHandleSet) DeleteFromSet(k Entry, sk *EntryHandle) bool {
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

func (me entryToEntryHandleSet) AddToSet(k Entry, sk *EntryHandle) {
	s, ok := me.m[k]
	if ok {
		s = s.Add(sk)
	} else {
		s = NewentryHandleSet(sk)
	}
	me.m[k] = s
}

func (me entryToEntryHandleSet) Len() int {
	return len(me.m)
}

func (me entryToEntryHandleSet) Range(f func(Entry, *entryHandleSet) bool) {
	for k, s := range me.m {
		if !f(k, s) {
			break
		}
	}
}
