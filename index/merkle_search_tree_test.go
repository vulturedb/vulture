package index

import (
	"crypto"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func putAndGetRunner(t *testing.T, base Base, iters, elems, keyMod int) {
	rand.Seed(42)
	for i := 0; i < iters; i++ {
		index := NewLocalMST(base, crypto.SHA256)
		collected := map[UInt32]Value{}
		for j := 0; j < elems; j++ {
			key := UInt32(rand.Uint32() % uint32(keyMod))
			val := UInt32(rand.Uint32())
			index.Put(key, val)
			existingVal, exists := collected[key]
			if exists {
				collected[key] = existingVal.Merge(val)
			} else {
				collected[key] = val
			}
			assert.Equal(t, index.Get(key), collected[key])
		}

		for key, val := range collected {
			assert.Equal(t, index.Get(key), val)
		}
	}
}

func TestMSTPutAndGetIterBase32(t *testing.T) {
	putAndGetRunner(t, Base32, 50, 1000, 100)
}

func TestMSTPutAndGetIterBase2(t *testing.T) {
	putAndGetRunner(t, Base2, 50, 1000, 100)
}

func mergeRunner(t *testing.T, base Base) {
	lInd := NewLocalMST(base, crypto.SHA256)
	rInd := NewLocalMST(base, crypto.SHA256)
	for i := 0; i < 50; i++ {
		lInd.Put(UInt32(i), UInt32(i))
		rInd.Put(UInt32(i+25), UInt32(i+50))
	}
	lIndCopy := lInd.Copy()

	err := lInd.Merge(rInd)
	assert.NoError(t, err)
	for i := 0; i < 25; i++ {
		assert.Equal(t, lInd.Get(UInt32(i)), UInt32(i))
	}
	for i := 25; i < 75; i++ {
		assert.Equal(t, lInd.Get(UInt32(i)), UInt32(i+25))
		assert.Equal(t, rInd.Get(UInt32(i)), UInt32(i+25))
	}

	err = rInd.Merge(lIndCopy)
	assert.NoError(t, err)
	for i := 0; i < 50; i++ {
		assert.Equal(t, lIndCopy.Get(UInt32(i)), UInt32(i))
	}
	for i := 0; i < 25; i++ {
		assert.Equal(t, rInd.Get(UInt32(i)), UInt32(i))
	}
	for i := 25; i < 75; i++ {
		assert.Equal(t, rInd.Get(UInt32(i)), UInt32(i+25))
	}
}

func TestMSTMergeBase32(t *testing.T) {
	mergeRunner(t, Base32)
}

func TestMSTMergeBase2(t *testing.T) {
	mergeRunner(t, Base2)
}
