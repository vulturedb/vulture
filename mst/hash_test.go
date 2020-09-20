package mst

import (
	"crypto"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type byteSlice []byte

func (b byteSlice) PutBytes(w io.Writer) error {
	_, err := w.Write(b)
	return err
}

func TestHashHashable(t *testing.T) {
	obj := byteSlice("vulturedb")
	actual := HashHashable(obj, crypto.MD5)
	expected, _ := hex.DecodeString("a5d94fbdd26039282fa3e22ba5b62f02")
	assert.Equal(t, expected, actual)
}
