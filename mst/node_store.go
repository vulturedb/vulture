package mst

import (
	"crypto"
	"io"

	"github.com/benbjohnson/immutable"
)

// Only used for computing the hash for a given node
type HashableNode Node

func hashToWrite(hash []byte, hashSize int) []byte {
	if hash == nil {
		return make([]byte, hashSize)
	}
	return hash
}

func (n *HashableNode) Write(w io.Writer) error {
	// TODO: Test this if it's ever used in something production related. It really shouldn't be since
	// this node store is mainly meant for testing. Persistent node stores will need to think more
	// about the serialization format.
	err := putUint32(n.level, w)
	if err != nil {
		return err
	}

	// Write key/values first
	err = putUint32(uint32(len(n.children)), w)
	if err != nil {
		return err
	}

	for _, child := range n.children {
		err = child.key.Write(w)
		if err != nil {
			return err
		}
		err = child.value.Write(w)
		if err != nil {
			return err
		}
	}

	// Assumes all hash sizes are exactly the same
	hashSize := 0
	if n.low != nil {
		hashSize = len(n.low)
	} else {
		for _, child := range n.children {
			if child.high != nil {
				hashSize = len(child.high)
				break
			}
		}
	}

	// Don't waste space on zeroed out hashes
	if hashSize > 0 {
		if n.low != nil {
			_, err = w.Write(hashToWrite(n.low, hashSize))
			if err != nil {
				return err
			}
		}

		for _, child := range n.children {
			_, err = w.Write(hashToWrite(child.high, hashSize))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type NodeStore interface {
	Get([]byte) *Node
	Put(*Node) (NodeStore, []byte)
	Remove([]byte) NodeStore
	Size() uint
}

type LocalNodeStore struct {
	dict *immutable.Map
	hash crypto.Hash
}

func NewLocalNodeStore(hash crypto.Hash) NodeStore {
	return &LocalNodeStore{dict: immutable.NewMap(nil), hash: hash}
}

func (ns *LocalNodeStore) withDict(dict *immutable.Map) NodeStore {
	return &LocalNodeStore{dict: dict, hash: ns.hash}
}

func (ns *LocalNodeStore) Get(k []byte) *Node {
	val, ok := ns.dict.Get(string(k))
	if ok {
		return val.(*Node)
	} else {
		return nil
	}
}

func (ns *LocalNodeStore) Put(n *Node) (NodeStore, []byte) {
	wn := HashableNode(*n)
	k := HashWritable(&wn, ns.hash)
	return ns.withDict(ns.dict.Set(string(k), n)), k
}

func (ns *LocalNodeStore) Remove(k []byte) NodeStore {
	return ns.withDict(ns.dict.Delete(string(k)))
}

func (ns *LocalNodeStore) Size() uint {
	return uint(ns.dict.Len())
}

func FindMissingNodes(ns NodeStore, hash []byte) [][]byte {
	if hash == nil {
		return [][]byte{}
	}
	n := ns.Get(hash)
	if n == nil {
		return [][]byte{hash}
	}
	missingNodes := [][]byte{}
	missingNodes = append(missingNodes, FindMissingNodes(ns, n.low)...)
	for _, child := range n.children {
		missingNodes = append(missingNodes, FindMissingNodes(ns, child.high)...)
	}
	return missingNodes
}
