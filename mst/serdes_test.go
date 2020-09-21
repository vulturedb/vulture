package mst

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPutGetUInt32(t *testing.T) {
	expected := uint32(69)
	buf := bytes.NewBuffer([]byte{})
	err := putUint32(expected, buf)
	assert.NoError(t, err)
	// Little endian over here
	assert.Equal(t, buf.Bytes(), []byte{0b01000101, 0b00000000, 0b00000000, 0b00000000})
}
