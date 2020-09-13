package index

import (
	"crypto"
	"encoding/binary"
	"io"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestMSTPutAndGetIter(t *testing.T) {
	rand.Seed(42)
	iters := 50
	elems := 1000
	keyMod := uint32(100)
	for i := 0; i < iters; i++ {
		index := NewLocalMST(Base16, crypto.SHA256)
		collected := map[UInt32]Value{}
		for j := 0; j < elems; j++ {
			key := UInt32(rand.Uint32() % keyMod)
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
