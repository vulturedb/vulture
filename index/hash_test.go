package index

import (
	"crypto"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type bytes []byte

func (b bytes) PutBytes(w io.Writer) error {
	_, err := w.Write(b)
	return err
}

func TestHash(t *testing.T) {
	obj := bytes("vulturedb")
	actual := hash(obj, crypto.MD5)
	expected, _ := hex.DecodeString("a5d94fbdd26039282fa3e22ba5b62f02")
	assert.Equal(t, expected, actual)
}
