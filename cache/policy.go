package cache

type Usage interface {
	Less(Usage) bool
}

type Policy interface {
	Candidate() (Key, bool)
	Update(Key, Usage)
	Forget(Key)
	NumItems() int
}
