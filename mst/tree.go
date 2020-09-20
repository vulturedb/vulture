package mst

import (
	"bytes"
	"crypto"
	"fmt"
	"strings"
)

type MerkleSearchTree struct {
	root  []byte
	base  Base
	hash  crypto.Hash
	store NodeStore
}

func NewLocalMST(base Base, hash crypto.Hash) *MerkleSearchTree {
	return &MerkleSearchTree{
		root:  nil,
		base:  base,
		hash:  hash,
		store: NewLocalNodeStore(hash),
	}
}

func (t *MerkleSearchTree) getNodeMaybe(hash []byte, store NodeStore) *Node {
	res := t.store.Get(hash)
	if res == nil {
		res = store.Get(hash)
	}
	return res
}

// Gets nodes from t.store but modifies store
func (t *MerkleSearchTree) splitInto(store NodeStore, nodeHash []byte, key Key) ([]byte, []byte) {
	if nodeHash == nil {
		return nil, nil
	}
	n := t.getNodeMaybe(nodeHash, store)
	child, i := n.findChild(key)
	if i > 0 && keysEqual(key, n.children[i-1].key) {
		panic(fmt.Errorf("Trying to get split node but key matches. Key: %v, Level: %d", key, n.level))
	}
	store.Remove(nodeHash)
	lChildren := make([]Child, i)
	rChildren := make([]Child, uint(len(n.children))-i)
	copy(lChildren, n.children[:i])
	copy(rChildren, n.children[i:])
	l, r := t.splitInto(store, child, key)
	var lHash, rHash []byte
	if len(lChildren) == 0 {
		lHash = l
	} else {
		lNode := &Node{
			level:    n.level,
			low:      n.low,
			children: lChildren,
		}
		lNode = lNode.withHashAt(l, i)
		lHash = store.Put(lNode)
	}
	if len(rChildren) == 0 {
		rHash = r
	} else {
		rNode := &Node{
			level:    n.level,
			low:      r,
			children: rChildren,
		}
		rHash = store.Put(rNode)
	}
	return lHash, rHash
}

func (t *MerkleSearchTree) split(nodeHash []byte, key Key) ([]byte, []byte) {
	return t.splitInto(t.store, nodeHash, key)
}

func (t *MerkleSearchTree) leadingZeros(key Key) uint32 {
	return t.base.LeadingZeros(HashHashable(key, t.hash))
}

func (t *MerkleSearchTree) put(nodeHash []byte, key Key, val Value, atLevel uint32) []byte {
	var newNode *Node = nil
	if nodeHash == nil {
		newNode = &Node{
			level:    atLevel,
			low:      nil,
			children: []Child{{key, val, nil}},
		}
	} else {
		n := t.store.Get(nodeHash)
		if atLevel < n.level {
			t.store.Remove(nodeHash)
			childHash, i := n.findChild(key)
			childHash = t.put(childHash, key, val, atLevel)
			newNode = n.withHashAt(childHash, i)
		} else if atLevel == n.level {
			t.store.Remove(nodeHash)
			i := n.find(key)
			if i > 0 && keysEqual(key, n.children[i-1].key) {
				newNode = n.withMergedValueAt(val, i-1)
			} else {
				l, r := t.split(n.childAt(i), key)
				newNode = n.withChildInsertedAt(key, val, r, i)
				newNode = newNode.withHashAt(l, i)
			}
		} else {
			l, r := t.split(nodeHash, key)
			newNode = &Node{
				level:    atLevel,
				low:      l,
				children: []Child{{key, val, r}},
			}
		}
	}
	return t.store.Put(newNode)
}

func (t *MerkleSearchTree) get(nodeHash []byte, key Key) Value {
	if nodeHash == nil {
		return nil
	}
	n := t.store.Get(nodeHash)
	i := n.find(key)
	var recurNode []byte = nil
	if i > 0 {
		if keysEqual(key, n.children[i-1].key) {
			return n.children[i-1].value
		}
		recurNode = n.children[i-1].node
	} else {
		recurNode = n.low
	}
	return t.get(recurNode, key)
}

func (t *MerkleSearchTree) merge(with *MerkleSearchTree, l []byte, r []byte) []byte {
	if l == nil && r != nil {
		// Recursively insert entire subtree into store
		rNode := with.store.Get(r)
		t.merge(with, nil, rNode.low)
		for _, rChild := range rNode.children {
			t.merge(with, nil, rChild.node)
		}
		return t.store.Put(rNode)
	} else if r == nil || bytes.Equal(l, r) {
		return l
	}

	lNode := t.store.Get(l)
	// When we split in certain cases, the split results are only added to t.store and not store
	// (since we never want to mutate with.store).
	rNode := with.getNodeMaybe(r, t.store)

	var level uint32
	var lLow, rLow []byte = l, r
	var lChildren, rChildren []Child = []Child{}, []Child{}
	if lNode.level >= rNode.level {
		lLow = lNode.low
		lChildren = lNode.children
		level = lNode.level
	}
	if lNode.level <= rNode.level {
		rLow = rNode.low
		rChildren = rNode.children
		level = rNode.level
	}
	lN, rN := uint(len(lChildren)), uint(len(rChildren))

	var low []byte = nil
	children := []Child{}

	lCur, rCur := uint(0), uint(0)
	for i := 0; lCur <= lN && rCur <= rN; i++ {
		var nextNode, interNode []byte = nil, nil
		if lCur == lN && rCur == rN {
			nextNode = t.merge(with, lLow, rLow)
			lCur++
			rCur++
		} else if lCur == lN {
			rChild := rNode.children[rCur]
			children = append(children, Child{rChild.key, rChild.value, nil})
			interNode, lLow = t.split(lLow, rChild.key)
			nextNode = t.merge(with, interNode, rLow)
			rLow = rChild.node
			rCur++
		} else if rCur == rN {
			lChild := lNode.children[lCur]
			children = append(children, Child{lChild.key, lChild.value, nil})
			interNode, rLow = with.splitInto(t.store, rLow, lChild.key)
			nextNode = t.merge(with, lLow, interNode)
			lLow = lChild.node
			lCur++
		} else {
			lChild := lNode.children[lCur]
			rChild := rNode.children[rCur]
			if lChild.key.Less(rChild.key) {
				children = append(children, Child{lChild.key, lChild.value, nil})
				interNode, rLow = with.splitInto(t.store, rLow, lChild.key)
				nextNode = t.merge(with, lLow, interNode)
				lLow = lChild.node
				lCur++
			} else if rChild.key.Less(lChild.key) {
				children = append(children, Child{rChild.key, rChild.value, nil})
				interNode, lLow = t.split(lLow, rChild.key)
				nextNode = t.merge(with, interNode, rLow)
				rLow = rChild.node
				rCur++
			} else {
				nextNode = t.merge(with, lLow, rLow)
				mergedValue := lChild.value.Merge(rChild.value)
				children = append(children, Child{lChild.key, mergedValue, nil})
				lLow = lChild.node
				rLow = rChild.node
				lCur++
				rCur++
			}
		}
		if i == 0 {
			low = nextNode
		} else {
			children[i-1].node = nextNode
		}
		if interNode != nil {
			t.store.Remove(interNode)
		}
	}

	if len(children) == 0 {
		panic("If the number of children is zero here, it means a tree is misformed")
	}

	t.store.Remove(l)
	t.store.Remove(r)
	return t.store.Put(&Node{
		level:    level,
		low:      low,
		children: children,
	})
}

func (t *MerkleSearchTree) printInOrder(nodeHash []byte, height uint32) {
	if nodeHash == nil {
		return
	}
	n := t.store.Get(nodeHash)
	t.printInOrder(n.low, height)
	for _, child := range n.children {
		fmt.Printf("%s%v -> %v\n", strings.Repeat("\t", int(height-n.level)), child.key, child.value)
		t.printInOrder(child.node, height)
	}
}

func (t *MerkleSearchTree) numNodes(n []byte) uint {
	if n == nil {
		return 0
	}
	node := t.store.Get(n)
	nNodes := t.numNodes(node.low)
	for _, child := range node.children {
		nNodes += t.numNodes(child.node)
	}
	return nNodes + 1
}

func (t *MerkleSearchTree) Put(key Key, val Value) {
	atLevel := t.leadingZeros(key)
	t.root = t.put(t.root, key, val, atLevel)
}

func (t *MerkleSearchTree) Get(key Key) Value {
	return t.get(t.root, key)
}

func (t *MerkleSearchTree) Merge(with *MerkleSearchTree) error {
	if t.base != with.base {
		return fmt.Errorf("Mismatching bases. 2^%d vs 2^%d", t.base, with.base)
	} else if t.hash != with.hash {
		// TODO: go 1.15 has string representation for hash functions so use that instead
		return fmt.Errorf("Mismatching hash functions. %d vs %d", t.hash, with.hash)
	}
	t.root = t.merge(with, t.root, with.root)
	return nil
}

func (t *MerkleSearchTree) PrintInOrder() {
	if t.root != nil {
		height := t.store.Get(t.root).level
		t.printInOrder(t.root, height)
	}
}

func (t *MerkleSearchTree) RootHash() []byte {
	return t.root
}

func (t *MerkleSearchTree) Copy() *MerkleSearchTree {
	return &MerkleSearchTree{
		root:  t.root,
		base:  t.base,
		hash:  t.hash,
		store: t.store.Copy(),
	}
}

func (t *MerkleSearchTree) NumNodes() uint {
	return t.numNodes(t.root)
}
