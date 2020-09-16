package index

import (
	"crypto"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func genKeyVal(keyMod int) (UInt32, UInt32) {
	key := UInt32(rand.Uint32() % uint32(keyMod))
	val := UInt32(rand.Uint32())
	return key, val
}

func putAndGetRunner(t *testing.T, base Base, iters, elems, keyMod int) {
	rand.Seed(42)
	for i := 0; i < iters; i++ {
		index := NewLocalMST(base, crypto.SHA256)
		collected := map[UInt32]Value{}
		for j := 0; j < elems; j++ {
			key, val := genKeyVal(keyMod)
			index.Put(key, val)
			if oVal, exists := collected[key]; exists {
				collected[key] = val.Merge(oVal)
			} else {
				collected[key] = val
			}
			assert.Equal(t, index.Get(key), collected[key])
		}

		for key, val := range collected {
			assert.Equal(t, index.Get(key), val)
		}
		assert.Equal(t, index.store.Size(), index.NumNodes())
	}
}

func TestMSTPutAndGetIterBase32(t *testing.T) {
	putAndGetRunner(t, Base32, 50, 1000, 100)
}

func TestMSTPutAndGetIterBase2(t *testing.T) {
	putAndGetRunner(t, Base2, 50, 1000, 100)
}

func mergeRunner(t *testing.T, base Base, iters, elems, keyMod int) {
	for i := 0; i < iters; i++ {
		lInd := NewLocalMST(base, crypto.SHA256)
		rInd := NewLocalMST(base, crypto.SHA256)

		lCollected := map[UInt32]Value{}
		rCollected := map[UInt32]Value{}
		for j := 0; j < elems; j++ {
			key, val := genKeyVal(keyMod)
			lInd.Put(key, val)
			if oVal, exists := lCollected[key]; exists {
				lCollected[key] = val.Merge(oVal)
			} else {
				lCollected[key] = val
			}
			key, val = genKeyVal(keyMod)
			rInd.Put(key, val)
			if oVal, exists := lCollected[key]; exists {
				lCollected[key] = val.Merge(oVal)
			} else {
				lCollected[key] = val
			}
			if oVal, exists := rCollected[key]; exists {
				rCollected[key] = val.Merge(oVal)
			} else {
				rCollected[key] = val
			}
		}

		err := lInd.Merge(rInd)
		assert.NoError(t, err)

		for key, val := range lCollected {
			assert.Equal(t, lInd.Get(key), val)
		}
		for key, val := range rCollected {
			assert.Equal(t, rInd.Get(key), val)
		}

		assert.Equal(t, lInd.store.Size(), lInd.NumNodes())
		assert.Equal(t, rInd.store.Size(), rInd.NumNodes())
	}
}

func TestMSTMergeBase32(t *testing.T) {
	mergeRunner(t, Base32, 50, 1000, 100)
}

func TestMSTMergeBase2(t *testing.T) {
	mergeRunner(t, Base2, 50, 1000, 100)
}

func TestMSTMergeConsecutive(t *testing.T) {
	// This used to catch a node leak so keeping the test around to make sure we don't regress.
	lInd := NewLocalMST(Base32, crypto.SHA256)
	rInd := NewLocalMST(Base32, crypto.SHA256)
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

	assert.Equal(t, lInd.store.Size(), lInd.NumNodes())
	assert.Equal(t, rInd.store.Size(), rInd.NumNodes())

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

	assert.Equal(t, lIndCopy.store.Size(), lIndCopy.NumNodes())
	assert.Equal(t, rInd.store.Size(), rInd.NumNodes())
}

func TestMSTMergeDiffBase(t *testing.T) {
	lInd := NewLocalMST(Base2, crypto.SHA256)
	rInd := NewLocalMST(Base32, crypto.SHA256)
	err := lInd.Merge(rInd)
	assert.Error(t, err)
	assert.Equal(t, "Mismatching bases. 2^1 vs 2^5", err.Error())
}

func TestMSTMergeDiffHash(t *testing.T) {
	lInd := NewLocalMST(Base32, crypto.SHA512)
	rInd := NewLocalMST(Base32, crypto.SHA256)
	err := lInd.Merge(rInd)
	assert.Error(t, err)
	assert.Equal(t, "Mismatching hash functions. 7 vs 5", err.Error())
}
