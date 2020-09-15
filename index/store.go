package index

import "crypto"

type NodeStore interface {
	Get([]byte) Hashable
	Put(Hashable) []byte
	Remove([]byte)
	Copy() NodeStore
}

type LocalNodeStore struct {
	dict map[string]Hashable
	hash crypto.Hash
}

func NewLocalNodeStore(hash crypto.Hash) *LocalNodeStore {
	return &LocalNodeStore{dict: map[string]Hashable{}, hash: hash}
}

func (ns *LocalNodeStore) Get(k []byte) Hashable {
	return ns.dict[string(k)]
}

func (ns *LocalNodeStore) Put(n Hashable) []byte {
	k := HashHashable(n, ns.hash)
	ns.dict[string(k)] = n
	return k
}

func (ns *LocalNodeStore) Remove(k []byte) {
	delete(ns.dict, string(k))
}

func (ns *LocalNodeStore) Copy() NodeStore {
	newDict := map[string]Hashable{}
	for k, v := range ns.dict {
		newDict[k] = v
	}
	return &LocalNodeStore{newDict, ns.hash}
}
