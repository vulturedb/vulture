package index

type Node interface{}

type NodeStore interface {
	Get(string) Node
	Put(string, Node)
	Remove(string)
}

type LocalNodeStore struct {
	dict map[string]Node
}

func NewLocalNodeStore() *LocalNodeStore {
	return &LocalNodeStore{dict: map[string]Node{}}
}

func (ns *LocalNodeStore) Get(k string) Node {
	return ns.dict[k]
}

func (ns *LocalNodeStore) Put(k string, v Node) {
	ns.dict[k] = v
}

func (ns *LocalNodeStore) Remove(k string) {
	delete(ns.dict, k)
}
