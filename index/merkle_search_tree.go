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

func (t *MerkleSearchTree) put(nodeHash []byte, key Key, val Value, atLevel uint32) []byte {
	var n *merkleSearchNode = nil
	if nodeHash == nil {
		n = &merkleSearchNode{
			level:    atLevel,
			low:      nil,
			children: []merkleSearchChild{{key, val, nil}},
		}
	} else {
		// TODO: Fill in
	}
	return t.store.Put(n)
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
