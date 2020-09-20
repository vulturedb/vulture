package mst

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeysEqual(t *testing.T) {
	assert.True(t, keysEqual(UInt32(69), UInt32(69)))
	assert.False(t, keysEqual(UInt32(69), UInt32(70)))
	assert.False(t, keysEqual(UInt32(69), UInt32(68)))
}
