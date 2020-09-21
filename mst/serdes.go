package mst

import (
	"encoding/binary"
	"io"
)

type Writable interface {
	Write(io.Writer) error
}

func putUint32(n uint32, w io.Writer) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, n)
	_, err := w.Write(buf)
	return err
}
