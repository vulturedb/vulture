package index

type Hashable interface {
	Bytes() []byte
}

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
