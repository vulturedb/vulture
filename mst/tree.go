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
	store NodeStore2
}

func NewLocalMST(base Base, hash crypto.Hash) *MerkleSearchTree {
	return &MerkleSearchTree{
		root:  nil,
		base:  base,
		hash:  hash,
		store: NewLocalNodeStore2(hash),
	}
}

func NewMSTWithRoot(root []byte, base Base, hash crypto.Hash, store NodeStore2) *MerkleSearchTree {
	return &MerkleSearchTree{
		root:  root,
		base:  base,
		hash:  hash,
		store: store,
	}
}

func NewMST(base Base, hash crypto.Hash, store NodeStore2) *MerkleSearchTree {
	return NewMSTWithRoot(nil, base, hash, store)
}

func (t *MerkleSearchTree) getNodeMaybe(hash []byte, store NodeStore2) *Node {
	res := t.store.Get(hash)
	if res == nil {
		res = store.Get(hash)
	}
	return res
}

// Gets nodes from t.store but modifies store
func (t *MerkleSearchTree) splitInto(
	store NodeStore2,
	nodeHash []byte,
	key Key,
) (NodeStore2, []byte, []byte) {
	if nodeHash == nil {
		return store, nil, nil
	}
	n := t.getNodeMaybe(nodeHash, store)
	child, i := n.findChild(key)
	if i > 0 && keysEqual(key, n.children[i-1].key) {
		panic(fmt.Errorf("Trying to get split node but key matches. Key: %v, Level: %d", key, n.level))
	}
	store = store.Remove(nodeHash)
	lChildren := make([]Child, i)
	rChildren := make([]Child, uint(len(n.children))-i)
	copy(lChildren, n.children[:i])
	copy(rChildren, n.children[i:])
	store, l, r := t.splitInto(store, child, key)
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
		store, lHash = store.Put(lNode)
	}
	if len(rChildren) == 0 {
		rHash = r
	} else {
		rNode := &Node{
			level:    n.level,
			low:      r,
			children: rChildren,
		}
		store, rHash = store.Put(rNode)
	}
	return store, lHash, rHash
}

func (t *MerkleSearchTree) split(
	store NodeStore2,
	nodeHash []byte,
	key Key,
) (NodeStore2, []byte, []byte) {
	return t.splitInto(store, nodeHash, key)
}

func (t *MerkleSearchTree) leadingZeros(key Key) uint32 {
	return t.base.LeadingZeros(HashWritable(key, t.hash))
}

func (t *MerkleSearchTree) put(
	store NodeStore2,
	nodeHash []byte,
	key Key,
	val Value,
	atLevel uint32,
) (NodeStore2, []byte) {
	var newNode *Node = nil
	if nodeHash == nil {
		newNode = &Node{
			level:    atLevel,
			low:      nil,
			children: []Child{{key, val, nil}},
		}
	} else {
		n := store.Get(nodeHash)
		if atLevel < n.level {
			store = store.Remove(nodeHash)
			childHash, i := n.findChild(key)
			store, childHash = t.put(store, childHash, key, val, atLevel)
			newNode = n.withHashAt(childHash, i)
		} else if atLevel == n.level {
			store = store.Remove(nodeHash)
			i := n.find(key)
			if i > 0 && keysEqual(key, n.children[i-1].key) {
				newNode = n.withMergedValueAt(val, i-1)
			} else {
				var l, r []byte
				store, l, r = t.split(store, n.childAt(i), key)
				newNode = n.withChildInsertedAt(key, val, r, i)
				newNode = newNode.withHashAt(l, i)
			}
		} else {
			var l, r []byte
			store, l, r = t.split(store, nodeHash, key)
			newNode = &Node{
				level:    atLevel,
				low:      l,
				children: []Child{{key, val, r}},
			}
		}
	}
	return store.Put(newNode)
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
		recurNode = n.children[i-1].high
	} else {
		recurNode = n.low
	}
	return t.get(recurNode, key)
}

func (t *MerkleSearchTree) merge(
	with *MerkleSearchTree,
	store NodeStore2,
	l []byte,
	r []byte,
) (NodeStore2, []byte) {
	if l == nil && r != nil {
		// Recursively insert entire subtree into store
		rNode := with.store.Get(r)
		store, _ = t.merge(with, store, nil, rNode.low)
		for _, rChild := range rNode.children {
			store, _ = t.merge(with, store, nil, rChild.high)
		}
		return store.Put(rNode)
	} else if r == nil || bytes.Equal(l, r) {
		return store, l
	}

	lNode := store.Get(l)
	// When we split in certain cases, the split results are only added to t.store and not store
	// (since we never want to mutate with.store).
	rNode := with.getNodeMaybe(r, store)

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
			store, nextNode = t.merge(with, store, lLow, rLow)
			lCur++
			rCur++
		} else if lCur == lN {
			rChild := rNode.children[rCur]
			children = append(children, Child{rChild.key, rChild.value, nil})
			store, interNode, lLow = t.split(store, lLow, rChild.key)
			store, nextNode = t.merge(with, store, interNode, rLow)
			rLow = rChild.high
			rCur++
		} else if rCur == rN {
			lChild := lNode.children[lCur]
			children = append(children, Child{lChild.key, lChild.value, nil})
			store, interNode, rLow = with.splitInto(store, rLow, lChild.key)
			store, nextNode = t.merge(with, store, lLow, interNode)
			lLow = lChild.high
			lCur++
		} else {
			lChild := lNode.children[lCur]
			rChild := rNode.children[rCur]
			if lChild.key.Less(rChild.key) {
				children = append(children, Child{lChild.key, lChild.value, nil})
				store, interNode, rLow = with.splitInto(store, rLow, lChild.key)
				store, nextNode = t.merge(with, store, lLow, interNode)
				lLow = lChild.high
				lCur++
			} else if rChild.key.Less(lChild.key) {
				children = append(children, Child{rChild.key, rChild.value, nil})
				store, interNode, lLow = t.split(store, lLow, rChild.key)
				store, nextNode = t.merge(with, store, interNode, rLow)
				rLow = rChild.high
				rCur++
			} else {
				store, nextNode = t.merge(with, store, lLow, rLow)
				mergedValue := lChild.value.Merge(rChild.value)
				children = append(children, Child{lChild.key, mergedValue, nil})
				lLow = lChild.high
				rLow = rChild.high
				lCur++
				rCur++
			}
		}
		if i == 0 {
			low = nextNode
		} else {
			children[i-1].high = nextNode
		}
		if interNode != nil {
			store = store.Remove(interNode)
		}
	}

	if len(children) == 0 {
		panic("If the number of children is zero here, it means a tree is misformed")
	}

	store = store.Remove(l)
	store = store.Remove(r)
	return store.Put(&Node{
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
		t.printInOrder(child.high, height)
	}
}

func (t *MerkleSearchTree) numNodes(n []byte) uint {
	if n == nil {
		return 0
	}
	node := t.store.Get(n)
	nNodes := t.numNodes(node.low)
	for _, child := range node.children {
		nNodes += t.numNodes(child.high)
	}
	return nNodes + 1
}

func (t *MerkleSearchTree) withStoreAndRoot(store NodeStore2, root []byte) *MerkleSearchTree {
	return &MerkleSearchTree{
		root:  root,
		base:  t.base,
		hash:  t.hash,
		store: store,
	}
}

func (t *MerkleSearchTree) Put(key Key, val Value) *MerkleSearchTree {
	atLevel := t.leadingZeros(key)
	newStore, newRoot := t.put(t.store, t.root, key, val, atLevel)
	return t.withStoreAndRoot(newStore, newRoot)
}

func (t *MerkleSearchTree) Get(key Key) Value {
	return t.get(t.root, key)
}

func (t *MerkleSearchTree) Merge(with *MerkleSearchTree) (*MerkleSearchTree, error) {
	if t.base != with.base {
		return nil, fmt.Errorf("Mismatching bases. 2^%d vs 2^%d", t.base, with.base)
	} else if t.hash != with.hash {
		return t, fmt.Errorf("Mismatching hash functions. %s vs %s", t.hash, with.hash)
	}
	newStore, newRoot := t.merge(with, t.store, t.root, with.root)
	return t.withStoreAndRoot(newStore, newRoot), nil
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

func (t *MerkleSearchTree) NumNodes() uint {
	return t.numNodes(t.root)
}
