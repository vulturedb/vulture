package core

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Int int

func (s Int) Less(than IndexElem) bool {
	return s < than.(Int)
}

func TestBTreeGet(t *testing.T) {
	val1 := Int(1)
	val2 := Int(2)
	val3 := Int(3)
	val4 := Int(4)
	val5 := Int(5)
	val6 := Int(6)
	val7 := Int(7)
	index := BTreeIndex{
		root: &bTreeNode{
			elems: []IndexElem{val3, val5},
			children: []*bTreeNode{
				{
					elems:    []IndexElem{val2},
					children: nil,
				},
				{
					elems:    []IndexElem{val4},
					children: nil,
				},
				{
					elems:    []IndexElem{val6},
					children: nil,
				},
			},
		},
		t: 2,
	}
	assert.Nil(t, index.Get(val1))
	assert.Equal(t, val2, index.Get(val2))
	assert.Equal(t, val3, index.Get(val3))
	assert.Equal(t, val4, index.Get(val4))
	assert.Equal(t, val5, index.Get(val5))
	assert.Equal(t, val6, index.Get(val6))
	assert.Nil(t, index.Get(val7))
}

func TestBTreeGetNilRoot(t *testing.T) {
	assert.Nil(t, NewBTreeIndex(2).Get(Int(1)))
}

func TestBTreePutAndGetIter(t *testing.T) {
	rand.Seed(42)
	iters := 50
	elems := 1000
	mod := 100
	for i := 0; i < iters; i++ {
		index := NewBTreeIndex(uint(3))
		collected := map[int]interface{}{}
		for j := 0; j < elems; j++ {
			elem := rand.Int() % mod
			_, exists := collected[elem]
			added := index.Put(Int(elem))
			assert.Equal(t, exists, !added)
			collected[elem] = nil
		}

		for elem := range collected {
			assert.Equal(t, Int(elem), index.Get(Int(elem)))
		}
	}
}
