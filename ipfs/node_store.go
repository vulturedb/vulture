package ipfs

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	node "github.com/ipfs/go-ipld-format"
	icore "github.com/ipfs/interface-go-ipfs-core"

	"github.com/vulturedb/vulture/mst"
)

type MSTKeyReader interface {
	FromBytes([]byte) (mst.Key, error)
}

type MSTValueReader interface {
	FromBytes([]byte) (mst.Value, error)
}

func hashToLink(hash []byte) (*node.Link, error) {
	if hash == nil {
		return nil, nil
	}
	_, cid, err := cid.CidFromBytes(hash)
	if err != nil {
		return nil, err
	}
	// TODO: Do we need to provide size and name
	return &node.Link{Cid: cid}, nil
}

func linkToHash(l *node.Link) []byte {
	if l == nil {
		return nil
	}
	return l.Cid.Bytes()
}

type iPFSMSTChild struct {
	Key   []byte     `mapstructure:"key"`
	Value []byte     `mapstructure:"value"`
	High  *node.Link `mapstructure:"high"`
}

func newIPFSMSTChild(c mst.Child) iPFSMSTChild {
	keyBuf := new(bytes.Buffer)
	valBuf := new(bytes.Buffer)
	err := c.Key().Write(keyBuf)
	if err != nil {
		panic(fmt.Errorf("Couldn't write key to byte buffer: %s", err))
	}
	err = c.Value().Write(valBuf)
	if err != nil {
		panic(fmt.Errorf("Couldn't write value to byte buffer: %s", err))
	}
	high, err := hashToLink(c.High())
	if err != nil {
		panic(fmt.Errorf("Couldn't create link from high hash: %s", err))
	}
	return iPFSMSTChild{keyBuf.Bytes(), valBuf.Bytes(), high}
}

func (c iPFSMSTChild) toMSTChild(kr MSTKeyReader, vr MSTValueReader) (mst.Child, error) {
	k, err := kr.FromBytes(c.Key)
	if err != nil {
		return mst.Child{}, err
	}
	v, err := vr.FromBytes(c.Value)
	if err != nil {
		return mst.Child{}, err
	}
	return mst.NewChild(k, v, linkToHash(c.High)), nil
}

type iPFSMSTNode struct {
	Level    uint32         `mapstructure:"level"`
	Low      *node.Link     `mapstructure:"low"`
	Children []iPFSMSTChild `mapstructure:"children"`
}

func newIPFSMSTNode(n *mst.Node) *iPFSMSTNode {
	nChildren := n.Children()
	children := make([]iPFSMSTChild, len(nChildren))
	for i, nChild := range nChildren {
		children[i] = newIPFSMSTChild(nChild)
	}
	low, err := hashToLink(n.Low())
	if err != nil {
		panic(fmt.Errorf("Couldn't create link from low hash: %s", err))
	}
	return &iPFSMSTNode{n.Level(), low, children}
}

func (n *iPFSMSTNode) toMSTNode(kr MSTKeyReader, vr MSTValueReader) (*mst.Node, error) {
	children := make([]mst.Child, len(n.Children))
	for i, nChild := range n.Children {
		var err error
		children[i], err = nChild.toMSTChild(kr, vr)
		if err != nil {
			return nil, err
		}
	}
	return mst.NewNode(n.Level, linkToHash(n.Low), children), nil
}

type IPFSMSTNodeStore struct {
	ctx           context.Context
	dagService    icore.APIDagService
	multihashType uint64
	keyReader     MSTKeyReader
	valReader     MSTValueReader
}

func NewIPFSMSTNodeStore(
	ctx context.Context,
	dagService icore.APIDagService,
	multihashType uint64,
	keyReader MSTKeyReader,
	valReader MSTValueReader,
) *IPFSMSTNodeStore {
	return &IPFSMSTNodeStore{ctx, dagService, multihashType, keyReader, valReader}
}

// TODO: Potentially change the node store interface to return errors and have the mst methods also
// return errors. Panicing here is pretty dangerous since it's likely that stuff could go wrong.

func (s *IPFSMSTNodeStore) Get(k []byte) *mst.Node {
	_, ndCid, err := cid.CidFromBytes(k)
	if err != nil {
		panic(fmt.Errorf("Couldn't create cid: %s", err))
	}
	nd, err := s.dagService.Get(s.ctx, ndCid)
	if err != nil {
		panic(fmt.Errorf("Couldn't get node: %s", err))
	}
	raw, _, err := nd.Resolve([]string{})
	if err != nil {
		panic(fmt.Errorf("Couldn't resolve node: %s", err))
	}
	node := &iPFSMSTNode{}
	err = unmarshal(node, raw)
	if err != nil {
		panic(fmt.Errorf("Couldn't unmarshal node: %s", err))
	}
	mstNode, err := node.toMSTNode(s.keyReader, s.valReader)
	if err != nil {
		panic(fmt.Errorf("Couldn't convert to mst node: %s", err))
	}
	return mstNode
}

func (s *IPFSMSTNodeStore) Put(n *mst.Node) []byte {
	nd, err := cbor.WrapObject(newIPFSMSTNode(n), s.multihashType, -1)
	if err != nil {
		panic(fmt.Errorf("Couldn't wrap object: %s", err))
	}
	err = s.dagService.Add(s.ctx, nd)
	if err != nil {
		panic(fmt.Errorf("Couldn't add node: %s", err))
	}
	return nd.Cid().Bytes()
}

func (s *IPFSMSTNodeStore) Remove(k []byte) {
	_, ndCid, err := cid.CidFromBytes(k)
	if err != nil {
		panic(fmt.Errorf("Couldn't create cid: %s", err))
	}
	err = s.dagService.Remove(s.ctx, ndCid)
	if err != nil {
		panic(fmt.Errorf("Couldn't remove node: %s", err))
	}
}

func (s *IPFSMSTNodeStore) Copy() mst.NodeStore {
	panic("Cannot call Copy on an IPFSMSTNodeStore")
}

func (s *IPFSMSTNodeStore) Size() uint {
	panic("Cannot call Size on an IPFSMSTNodeStore")
}
