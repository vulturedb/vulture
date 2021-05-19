package mst

import (
	"io"
)

type Key interface {
	Writable
	Less(than Key) bool
}

type Value interface {
	Writable
	Merge(with Value) Value
}

type KeyReader interface {
	FromBytes([]byte) (Key, error)
}

type ValueReader interface {
	FromBytes([]byte) (Value, error)
}

func keysEqual(k1 Key, k2 Key) bool {
	return !k1.Less(k2) && !k2.Less(k1)
}

// Used mainly for testing
type UInt32 uint32

func (f UInt32) Less(than Key) bool {
	return f < than.(UInt32)
}

func (f UInt32) Write(w io.Writer) error {
	return putUint32(uint32(f), w)
}

func (f UInt32) Merge(with Value) Value {
	if with.(UInt32).Less(f) {
		return f
	} else {
		return with
	}
}
