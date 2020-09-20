package mst

import "crypto"

type NodeStore interface {
	Get([]byte) *Node
	Put(*Node) []byte
	Remove([]byte)
	Copy() NodeStore
	Size() uint
}

type LocalNodeStore struct {
	dict map[string]*Node
	hash crypto.Hash
}

func NewLocalNodeStore(hash crypto.Hash) *LocalNodeStore {
	return &LocalNodeStore{dict: map[string]*Node{}, hash: hash}
}

func (ns *LocalNodeStore) Get(k []byte) *Node {
	return ns.dict[string(k)]
}

func (ns *LocalNodeStore) Put(n *Node) []byte {
	k := HashHashable(n, ns.hash)
	ns.dict[string(k)] = n
	return k
}

func (ns *LocalNodeStore) Remove(k []byte) {
	delete(ns.dict, string(k))
}

func (ns *LocalNodeStore) Copy() NodeStore {
	newDict := map[string]*Node{}
	for k, v := range ns.dict {
		newDict[k] = v
	}
	return &LocalNodeStore{newDict, ns.hash}
}

func (ns *LocalNodeStore) Size() uint {
	return uint(len(ns.dict))
}
