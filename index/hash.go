package index

import (
	"crypto"
	"io"
)

type Hashable interface {
	PutBytes(io.Writer) error
}

func HashHashable(obj Hashable, h crypto.Hash) []byte {
	hasher := h.New()
	err := obj.PutBytes(hasher)
	if err != nil {
		// This should never error for any hash function
		panic(err)
	}
	return hasher.Sum(nil)
}
