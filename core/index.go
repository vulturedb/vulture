package core

import (
	"fmt"
	"sort"
	"strings"
)

type IndexElem interface {
	Less(than IndexElem) bool
}

type Index interface {
	Get(IndexElem) IndexElem
	Put(IndexElem) bool
	Delete(IndexElem) bool
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

// Returns first index where element is less than the value at the index.
// If no such index is found, return len(n.elems).
func (n *bTreeNode) find(e IndexElem) uint {
	i := sort.Search(len(n.elems), func(i int) bool {
		return e.Less(n.elems[i])
	})
	if i < 0 {
		panic(fmt.Sprintf("i cannot be < 0, got %d", i))
	}
	return uint(i)
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
		// leaf node
		return nil
	}
	return n.children[i].get(e)
}

type bTreeSplitResult struct {
	mElem IndexElem
	lNode *bTreeNode
	rNode *bTreeNode
}

type bTreeSplitBias int

const (
	bTreeLeftBias  = bTreeSplitBias(1)
	bTreeRightBias = bTreeSplitBias(0)
)

func (n *bTreeNode) split(t uint, bias bTreeSplitBias) *bTreeSplitResult {
	var lChildren, rChildren []*bTreeNode
	if n.children == nil {
		lChildren = nil
		rChildren = nil
	} else {
		lChildren = n.children[:t]
		rChildren = n.children[t:]
	}
	return &bTreeSplitResult{
		mElem: n.elems[t-1],
		lNode: &bTreeNode{
			elems:    n.elems[:t-1],
			children: lChildren,
		},
		rNode: &bTreeNode{
			elems:    n.elems[t:],
			children: rChildren,
		},
	}
}

func (n *bTreeNode) insertElemAt(e IndexElem, i uint) {
	newElems := make([]IndexElem, len(n.elems)+1)
	copy(newElems[:i], n.elems[:i])
	copy(newElems[i+1:], n.elems[i:])
	newElems[i] = e
	n.elems = newElems
}

func (n *bTreeNode) insertChildrenAt(lNode *bTreeNode, rNode *bTreeNode, i uint) {
	newChildren := make([]*bTreeNode, len(n.children)+1)
	copy(newChildren[:i], n.children[:i])
	copy(newChildren[i+2:], n.children[i+1:])
	newChildren[i] = lNode
	newChildren[i+1] = rNode
	n.children = newChildren
}

func (n *bTreeNode) put(e IndexElem, t uint) (*bTreeSplitResult, bool) {
	i := n.find(e)
	if i > 0 && indexElemsEqual(n.elems[i-1], e) {
		return nil, false
	}
	var added bool
	if n.children == nil {
		// leaf node
		n.insertElemAt(e, i)
		added = true
	} else {
		var splitRes *bTreeSplitResult
		splitRes, added = n.children[i].put(e, t)
		if splitRes != nil {
			n.insertElemAt(splitRes.mElem, i)
			n.insertChildrenAt(splitRes.lNode, splitRes.rNode, i)
		}
	}

	if uint(len(n.elems)) == 2*t-1 {
		// make sure that the node that was in the middle is the one promoted up
		var bias bTreeSplitBias
		if i < t {
			bias = bTreeLeftBias
		} else {
			bias = bTreeRightBias
		}
		return n.split(t, bias), added
	}
	return nil, added
}

func (n *bTreeNode) traverse(depth uint, f func(IndexElem, uint)) {
	for i := 0; i < len(n.elems)+1; i++ {
		if n.children != nil {
			n.children[i].traverse(depth+1, f)
		}
		if i < len(n.elems) {
			f(n.elems[i], depth)
		}
	}
}

func (i *BTreeIndex) Get(e IndexElem) IndexElem {
	if i.root == nil {
		return nil
	}
	return i.root.get(e)
}

func (i *BTreeIndex) Put(e IndexElem) bool {
	if i.root == nil {
		i.root = &bTreeNode{elems: []IndexElem{e}, children: nil}
		return true
	}

	splitRes, added := i.root.put(e, i.t)
	if splitRes != nil {
		i.root = &bTreeNode{
			elems:    []IndexElem{splitRes.mElem},
			children: []*bTreeNode{splitRes.lNode, splitRes.rNode},
		}
	}
	return added
}

func (i *BTreeIndex) Delete(e IndexElem) {
	// TODO: Implement
}

func (i *BTreeIndex) PrintInOrder() {
	if i.root == nil {
		fmt.Println("nil")
		return
	}
	i.root.traverse(uint(0), func(e IndexElem, depth uint) {
		fmt.Printf("%s%v\n", strings.Repeat("\t", int(depth)), e)
	})
}