package mst

import (
	"crypto"
)

func HashWritable(obj Writable, h crypto.Hash) []byte {
	hasher := h.New()
	err := obj.Write(hasher)
	if err != nil {
		// This should never error for any hash function
		panic(err)
	}
	return hasher.Sum(nil)
}
