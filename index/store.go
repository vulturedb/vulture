package index

import "crypto"

type nodeStore interface {
	Get([]byte) Hashable
	Put(Hashable) []byte
	Remove([]byte)
}

type localNodeStore struct {
	dict map[string]Hashable
	hash crypto.Hash
}

func newLocalNodeStore(hash crypto.Hash) *localNodeStore {
	return &localNodeStore{dict: map[string]Hashable{}, hash: hash}
}

func (ns *localNodeStore) Get(k []byte) Hashable {
	return ns.dict[string(k)]
}

func (ns *localNodeStore) Put(n Hashable) []byte {
	k := hash(n, ns.hash)
	ns.dict[string(k)] = n
	return k
}

func (ns *localNodeStore) Remove(k []byte) {
	delete(ns.dict, string(k))
}
