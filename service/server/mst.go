package server

import (
	"context"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/vulturedb/vulture/mst"
	"github.com/vulturedb/vulture/service/rpc"
)

type MSTServer struct {
	tree *mst.MerkleSearchTree
}

func NewMSTServer(tree *mst.MerkleSearchTree) *MSTServer {
	return &MSTServer{tree}
}

func (s *MSTServer) Get(ctx context.Context, in *rpc.MSTGetRequest) (*rpc.MSTGetResponse, error) {
	key := in.GetKey()
	val := s.tree.Get(mst.UInt32(key))
	log.Printf("Get %d: %d", key, val)
	if val == nil {
		return nil, fmt.Errorf("not found")
	} else {
		return &rpc.MSTGetResponse{Value: uint32(val.(mst.UInt32))}, nil
	}
}

func (s *MSTServer) Put(ctx context.Context, in *rpc.MSTPutRequest) (*empty.Empty, error) {
	key := in.GetKey()
	val := in.GetValue()
	s.tree.Put(mst.UInt32(key), mst.UInt32(val))
	log.Printf("Put %d: %d", key, val)
	return &empty.Empty{}, nil
}
