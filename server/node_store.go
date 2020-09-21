package server

import (
	"context"

	icore "github.com/ipfs/interface-go-ipfs-core"

	"github.com/vulturedb/vulture/mst"
)

type IPFSMSTNode struct {
	level uint32
}

type IPFSMSTNodeStore struct {
	ctx        context.Context
	dagService icore.APIDagService
}

func NewIPFSMSTNodeStore(ctx context.Context, dagService icore.APIDagService) *IPFSMSTNodeStore {
	return &IPFSMSTNodeStore{ctx, dagService}
}

func (*IPFSMSTNodeStore) Get(k []byte) mst.Hashable {
	return nil
}

func (*IPFSMSTNodeStore) Put(n mst.Hashable) []byte {
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
