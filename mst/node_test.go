package mst

import (
	"crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindMissingNodesEmpty(t *testing.T) {
	ns := NewLocalNodeStore(crypto.SHA256)
	assert.Equal(t, [][]byte{}, FindMissingNodes(ns, nil))
	assert.Equal(t, [][]byte{{1, 2, 3}}, FindMissingNodes(ns, []byte{1, 2, 3}))
}

func TestFindMissingNodesSomeChildren(t *testing.T) {
	ns := NewLocalNodeStore(crypto.SHA256)
	nChild := &Node{0, nil, []Child{{UInt32(1), UInt32(2), []byte{1, 2, 3}}}}
	hChild := ns.Put(nChild)
	nRoot := &Node{1, []byte{2, 3, 4}, []Child{{UInt32(3), UInt32(4), hChild}}}
	hRoot := ns.Put(nRoot)
	assert.Equal(t, [][]byte{{2, 3, 4}, {1, 2, 3}}, FindMissingNodes(ns, hRoot))
}
