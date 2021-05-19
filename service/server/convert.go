package server

import (
	"bytes"
	"fmt"

	"github.com/vulturedb/vulture/mst"
	"github.com/vulturedb/vulture/service/rpc"
)

// TODO: Write tests for this

func childToRPC(child mst.Child) *rpc.MSTChild {
	keyBuf := new(bytes.Buffer)
	valBuf := new(bytes.Buffer)
	err := child.Key().Write(keyBuf)
	if err != nil {
		panic(fmt.Errorf("Couldn't write key to byte buffer: %s", err))
	}
	err = child.Value().Write(valBuf)
	if err != nil {
		panic(fmt.Errorf("Couldn't write value to byte buffer: %s", err))
	}
	return &rpc.MSTChild{
		Key:   keyBuf.Bytes(),
		Value: valBuf.Bytes(),
		High:  child.High(),
	}
}

// NodeToRPC converts a native mst.Node type into the transport layer
func NodeToRPC(node *mst.Node) *rpc.MSTNode {
	children := node.Children()
	rpcChildren := make([]*rpc.MSTChild, 0, len(children))
	for _, child := range children {
		rpcChildren = append(rpcChildren, childToRPC(child))
	}
	return &rpc.MSTNode{
		Level:    node.Level(),
		Low:      node.Low(),
		Children: rpcChildren,
	}
}

func childFromRPC(child *rpc.MSTChild, kr mst.KeyReader, vr mst.ValueReader) (mst.Child, error) {
	k, err := kr.FromBytes(child.GetKey())
	if err != nil {
		return mst.Child{}, err
	}
	v, err := vr.FromBytes(child.GetValue())
	if err != nil {
		return mst.Child{}, err
	}
	return mst.NewChild(k, v, child.GetHigh()), nil
}

// NodeFromRPC creates a native mst.Node type from the transport layer
func NodeFromRPC(node *rpc.MSTNode, kr mst.KeyReader, vr mst.ValueReader) (*mst.Node, error) {
	children := node.GetChildren()
	mstChildren := make([]mst.Child, 0, len(children))
	for _, child := range children {
		mstChild, err := childFromRPC(child, kr, vr)
		if err != nil {
			return nil, err
		}
		mstChildren = append(mstChildren, mstChild)
	}
	return mst.NewNode(node.GetLevel(), node.GetLow(), mstChildren), nil
}
