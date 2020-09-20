package server

import (
	"context"

	icore "github.com/ipfs/interface-go-ipfs-core"

	"github.com/vulturedb/vulture/index"
)

type IPFSNodeStore struct {
	ctx        context.Context
	dagService icore.APIDagService
}

func NewIPFSNodeStore(ctx context.Context, dagService icore.APIDagService) *IPFSNodeStore {
	return &IPFSNodeStore{ctx, dagService}
}

func (*IPFSNodeStore) Get(k []byte) index.Hashable {
	return nil
}

func (*IPFSNodeStore) Put(n index.Hashable) []byte {
	return nil
}

func (*IPFSNodeStore) Remove(k []byte) {

}

func Copy() index.NodeStore {
	return nil
}

func Size() uint {
	return 0
}
