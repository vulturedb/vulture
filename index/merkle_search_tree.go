package index

import (
	"bytes"
	"crypto"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"strings"
)

type merkleSearchChild struct {
	key   Key
	value Value
	node  []byte
}

type merkleSearchNode struct {
	level    uint32
	low      []byte
	children []merkleSearchChild
}

func (n *merkleSearchNode) PutBytes(w io.Writer) error {
	// This can probably be improved...
	// I still don't even know if there are any cases where this easily breaks
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, n.level)
	_, err := w.Write(buf)
	if err != nil {
		return err
	}
	// Write key/values first
	if n.children != nil {
		for _, child := range n.children {
			err = child.key.PutBytes(w)
			if err != nil {
				return err
			}
			err = child.value.PutBytes(w)
			if err != nil {
				return err
			}
		}
	}
	// Links second
	if n.low != nil {
		_, err = w.Write(n.low)
		if err != nil {
			return err
		}
	}
	if n.children != nil {
		for _, child := range n.children {
			if child.node != nil {
				_, err = w.Write(child.node)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (n *merkleSearchNode) find(key Key) uint {
	i := sort.Search(len(n.children), func(i int) bool {
		return key.Less(n.children[i].key)
	})
	if i < 0 {
		panic(fmt.Sprintf("i cannot be < 0, got %d", i))
	}
	return uint(i)
}

func (n *merkleSearchNode) childAt(i uint) []byte {
	if i == 0 {
		return n.low
	} else {
		return n.children[i-1].node
	}
}

func (n *merkleSearchNode) findChild(key Key) ([]byte, uint) {
	i := n.find(key)
	if i > 0 && keysEqual(key, n.children[i-1].key) {
		panic(fmt.Errorf("Trying to get childHash but key matches. Key: %v, Level: %d", key, n.level))
	}
	return n.childAt(i), i
}

func (n *merkleSearchNode) withHashAt(hash []byte, at uint) *merkleSearchNode {
	if at == 0 {
		return &merkleSearchNode{
			level:    n.level,
			low:      hash,
			children: n.children,
		}
	} else {
		newChildren := make([]merkleSearchChild, len(n.children))
		copy(newChildren, n.children)
		newChildren[at-1].node = hash
		return &merkleSearchNode{
			level:    n.level,
			low:      n.low,
			children: newChildren,
		}
	}
}

func (n *merkleSearchNode) withMergedValueAt(val Value, at uint) *merkleSearchNode {
	newChildren := make([]merkleSearchChild, len(n.children))
	copy(newChildren, n.children)
	newChildren[at].value = n.children[at].value.Merge(val)
	return &merkleSearchNode{
		level:    n.level,
		low:      n.low,
		children: newChildren,
	}
}

func (n *merkleSearchNode) withChildInsertedAt(
	key Key,
	val Value,
	node []byte,
	at uint,
) *merkleSearchNode {
	newChildren := make([]merkleSearchChild, len(n.children)+1)
	copy(newChildren[:at], n.children[:at])
	copy(newChildren[at+1:], n.children[at:])
	newChildren[at] = merkleSearchChild{key, val, node}
	return &merkleSearchNode{
		level:    n.level,
		low:      n.low,
		children: newChildren,
	}
}

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

func (t *MerkleSearchTree) getNode(hash []byte) *merkleSearchNode {
	return t.store.Get(hash).(*merkleSearchNode)
}

func (t *MerkleSearchTree) getNodeMaybe(hash []byte, store NodeStore) *merkleSearchNode {
	res := t.store.Get(hash)
	if res == nil {
		res = store.Get(hash)
	}
	return res.(*merkleSearchNode)
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
	lChildren := make([]merkleSearchChild, i)
	rChildren := make([]merkleSearchChild, uint(len(n.children))-i)
	copy(lChildren, n.children[:i])
	copy(rChildren, n.children[i:])
	l, r := t.splitInto(store, child, key)
	var lHash, rHash []byte
	if len(lChildren) == 0 {
		lHash = l
	} else {
		lNode := &merkleSearchNode{
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
		rNode := &merkleSearchNode{
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
	var newNode *merkleSearchNode = nil
	if nodeHash == nil {
		newNode = &merkleSearchNode{
			level:    atLevel,
			low:      nil,
			children: []merkleSearchChild{{key, val, nil}},
		}
	} else {
		n := t.getNode(nodeHash)
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
			newNode = &merkleSearchNode{
				level:    atLevel,
				low:      l,
				children: []merkleSearchChild{{key, val, r}},
			}
		}
	}
	return t.store.Put(newNode)
}

func (t *MerkleSearchTree) get(nodeHash []byte, key Key) Value {
	if nodeHash == nil {
		return nil
	}
	n := t.getNode(nodeHash)
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
		rNode := with.getNode(r)
		t.merge(with, nil, rNode.low)
		for _, rChild := range rNode.children {
			t.merge(with, nil, rChild.node)
		}
		return t.store.Put(rNode)
	} else if r == nil || bytes.Equal(l, r) {
		return l
	}

	lNode := t.getNode(l)
	// When we split in certain cases, the split results are only added to t.store and not store
	// (since we never want to mutate with.store).
	rNode := with.getNodeMaybe(r, t.store)

	var level uint32
	var lLow, rLow []byte = l, r
	var lChildren, rChildren []merkleSearchChild = []merkleSearchChild{}, []merkleSearchChild{}
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
	children := []merkleSearchChild{}

	lCur, rCur := uint(0), uint(0)
	for i := 0; lCur <= lN && rCur <= rN; i++ {
		var nextNode, interNode []byte = nil, nil
		if lCur == lN && rCur == rN {
			nextNode = t.merge(with, lLow, rLow)
			lCur++
			rCur++
		} else if lCur == lN {
			rChild := rNode.children[rCur]
			children = append(children, merkleSearchChild{rChild.key, rChild.value, nil})
			interNode, lLow = t.split(lLow, rChild.key)
			nextNode = t.merge(with, interNode, rLow)
			rLow = rChild.node
			rCur++
		} else if rCur == rN {
			lChild := lNode.children[lCur]
			children = append(children, merkleSearchChild{lChild.key, lChild.value, nil})
			interNode, rLow = with.splitInto(t.store, rLow, lChild.key)
			nextNode = t.merge(with, lLow, interNode)
			lLow = lChild.node
			lCur++
		} else {
			lChild := lNode.children[lCur]
			rChild := rNode.children[rCur]
			if lChild.key.Less(rChild.key) {
				children = append(children, merkleSearchChild{lChild.key, lChild.value, nil})
				interNode, rLow = with.splitInto(t.store, rLow, lChild.key)
				nextNode = t.merge(with, lLow, interNode)
				lLow = lChild.node
				lCur++
			} else if rChild.key.Less(lChild.key) {
				children = append(children, merkleSearchChild{rChild.key, rChild.value, nil})
				interNode, lLow = t.split(lLow, rChild.key)
				nextNode = t.merge(with, interNode, rLow)
				rLow = rChild.node
				rCur++
			} else {
				nextNode = t.merge(with, lLow, rLow)
				mergedValue := lChild.value.Merge(rChild.value)
				children = append(children, merkleSearchChild{lChild.key, mergedValue, nil})
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
	return t.store.Put(&merkleSearchNode{
		level:    level,
		low:      low,
		children: children,
	})
}

func (t *MerkleSearchTree) printInOrder(nodeHash []byte, height uint32) {
	if nodeHash == nil {
		return
	}
	n := t.getNode(nodeHash)
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
	node := t.getNode(n)
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
		return fmt.Errorf("Mismatching hash funcitons. %d vs %d", t.hash, with.hash)
	}
	t.root = t.merge(with, t.root, with.root)
	return nil
}

func (t *MerkleSearchTree) PrintInOrder() {
	if t.root != nil {
		height := t.store.Get(t.root).(*merkleSearchNode).level
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
