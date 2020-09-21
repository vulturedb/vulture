package server

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	node "github.com/ipfs/go-ipld-format"
	icore "github.com/ipfs/interface-go-ipfs-core"

	"github.com/vulturedb/vulture/mst"
)

type iPFSMSTChild struct {
	Key   []byte     `mapstructure:"key"`
	Value []byte     `mapstructure:"value"`
	High  *node.Link `mapstructure:"high"`
}

type iPFSMSTNode struct {
	Level    uint32         `mapstructure:"level"`
	Low      *node.Link     `mapstructure:"low"`
	Children []iPFSMSTChild `mapstructure:"children"`
}

func hashToLink(hash []byte) *node.Link {
	if hash == nil {
		return nil
	}
	_, cid, err := cid.CidFromBytes(hash)
	if err != nil {
		panic(fmt.Errorf("Couldn't convert bytes to cid (%s): %s", hash, err))
	}
	// TODO: Do we need to provide size and name
	return &node.Link{Cid: cid}

}

func newIPFSMSTChild(c mst.Child) iPFSMSTChild {
	var keyBuf, valBuf *bytes.Buffer
	err := c.Key().Write(keyBuf)
	if err != nil {
		panic(fmt.Errorf("Couldn't write key to byte buffer: %s", err))
	}
	err = c.Value().Write(valBuf)
	if err != nil {
		panic(fmt.Errorf("Couldn't write key to byte buffer: %s", err))
	}
	return iPFSMSTChild{keyBuf.Bytes(), valBuf.Bytes(), hashToLink(c.High())}
}

func newIPFSMSTNode(n *mst.Node) *iPFSMSTNode {
	nChildren := n.Children()
	children := make([]iPFSMSTChild, len(nChildren))
	for i, nChild := range nChildren {
		children[i] = newIPFSMSTChild(nChild)
	}
	return &iPFSMSTNode{n.Level(), hashToLink(n.Low()), children}
}

type IPFSMSTNodeStore struct {
	ctx        context.Context
	dagService icore.APIDagService
}

func NewIPFSMSTNodeStore(ctx context.Context, dagService icore.APIDagService) *IPFSMSTNodeStore {
	return &IPFSMSTNodeStore{ctx, dagService}
}

func (*IPFSMSTNodeStore) Get(k []byte) *mst.Node {
	return nil
}

func (*IPFSMSTNodeStore) Put(n *mst.Node) []byte {
	return nil
}

func (*IPFSMSTNodeStore) Remove(k []byte) {

}

func Copy() mst.NodeStore {
	return nil
}

func Size() uint {
	return 0
}
