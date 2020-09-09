package core

import (
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
