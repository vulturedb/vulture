package mst

import (
	"encoding/binary"
	"io"
)

type Key interface {
	Hashable
	Less(than Key) bool
}

type Value interface {
	Hashable
	Merge(with Value) Value
}

func keysEqual(k1 Key, k2 Key) bool {
	return !k1.Less(k2) && !k2.Less(k1)
}

// Used mainly for testing
type UInt32 uint32

func (f UInt32) Less(than Key) bool {
	return f < than.(UInt32)
}

func (f UInt32) PutBytes(w io.Writer) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(f))
	_, err := w.Write(buf)
	return err
}

func (f UInt32) Merge(with Value) Value {
	if f > with.(UInt32) {
		return f
	} else {
		return with
	}
}
