package index

import (
	"crypto"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
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
	store nodeStore
}

func NewLocalMST(base Base, hash crypto.Hash) *MerkleSearchTree {
	return &MerkleSearchTree{
		root:  nil,
		base:  base,
		hash:  hash,
		store: newLocalNodeStore(hash),
	}
}

func (t *MerkleSearchTree) split(nodeHash []byte, key Key) ([]byte, []byte) {
	if nodeHash == nil {
		return nil, nil
	}
	n := t.store.Get(nodeHash).(*merkleSearchNode)
	child, i := n.findChild(key)
	if i > 0 && keysEqual(key, n.children[i-1].key) {
		panic(fmt.Errorf("Trying to get split node but key matches. Key: %v, Level: %d", key, n.level))
	}
	t.store.Remove(nodeHash)
	lChildren := make([]merkleSearchChild, i)
	rChildren := make([]merkleSearchChild, uint(len(n.children))-i)
	copy(lChildren, n.children[:i])
	copy(rChildren, n.children[i:])
	l, r := t.split(child, key)
	lNode := &merkleSearchNode{
		level:    n.level,
		low:      n.low,
		children: lChildren,
	}
	rNode := &merkleSearchNode{
		level:    n.level,
		low:      r,
		children: rChildren,
	}
	lNode = lNode.withHashAt(l, i)
	return t.store.Put(lNode), t.store.Put(rNode)
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
		n := t.store.Get(nodeHash).(*merkleSearchNode)
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

func (t *MerkleSearchTree) leadingZeros(key Key) uint32 {
	return t.base.LeadingZeros(hash(key, t.hash))
}

func (t *MerkleSearchTree) Put(key Key, val Value) {
	t.root = t.put(t.root, key, val, t.leadingZeros(key))
}

func (t *MerkleSearchTree) get(nodeHash []byte, key Key) Value {
	if nodeHash == nil {
		return nil
	}
	n := t.store.Get(nodeHash).(*merkleSearchNode)
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

func (t *MerkleSearchTree) Get(key Key) Value {
	return t.get(t.root, key)
}

func (t *MerkleSearchTree) RootHash() []byte {
	return t.root
}
