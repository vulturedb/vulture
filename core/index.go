package core

import "sort"

type IndexElem interface {
	Less(than IndexElem) bool
}

type Index interface {
	Get(IndexElem) IndexElem
	Put(IndexElem)
}

type BTreeIndex struct {
	root *bTreeNode
	t    uint
}

type bTreeNode struct {
	elems    []IndexElem
	children []*bTreeNode
}

func NewBTreeIndex(t uint) *BTreeIndex {
	return &BTreeIndex{nil, t}
}

func newBTreeBranch(t int) *bTreeNode {
	return &bTreeNode{[]IndexElem{}, []*bTreeNode{}}
}

// Returns first index where element is less than the value at the index.
// If no such index is found, return len(n.elems).
func (n *bTreeNode) find(e IndexElem) int {
	return sort.Search(len(n.elems), func(i int) bool {
		return e.Less(n.elems[i])
	})
}

func indexElemsEqual(a IndexElem, b IndexElem) bool {
	return !a.Less(b) && !b.Less(a)
}

func (n *bTreeNode) get(e IndexElem) IndexElem {
	i := n.find(e)
	if i > 0 && indexElemsEqual(n.elems[i-1], e) {
		return n.elems[i-1]
	}
	if n.children == nil {
		return nil
	}
	return n.children[i].get(e)
}

func (i *BTreeIndex) Get(e IndexElem) IndexElem {
	if i.root == nil {
		return nil
	}
	return i.root.get(e)
}

func (i *BTreeIndex) Put(e IndexElem) {
	// TODO: Implement
}
