package index

type Key interface {
	Hashable
	Less(than Key) bool
}

type Value interface {
	Hashable
}

type Index interface {
	Put(Key, Value)
	Get(Key) Value
}

func keysEqual(k1 Key, k2 Key) bool {
	return !k1.Less(k2) && !k2.Less(k1)
}
