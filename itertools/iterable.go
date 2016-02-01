package itertools

type Iterable interface {
	Iter() Iterator
}
