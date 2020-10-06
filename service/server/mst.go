package server

import (
	"context"
	"io"
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
	var primVal uint32
	if val == nil {
		primVal = 0
	} else {
		primVal = uint32(val.(mst.UInt32))
	}
	log.Printf("Get %d: %d", key, primVal)
	return &rpc.MSTGetResponse{Value: primVal}, nil
}

func (s *MSTServer) Put(ctx context.Context, in *rpc.MSTPutRequest) (*empty.Empty, error) {
	key := in.GetKey()
	val := in.GetValue()
	s.tree.Put(mst.UInt32(key), mst.UInt32(val))
	log.Printf("Put %d: %d", key, val)
	return &empty.Empty{}, nil
}

type MSTManagerServer struct {
}

func NewMSTManagerServer() *MSTManagerServer {
	return &MSTManagerServer{}
}

func (s *MSTManagerServer) Manage(stream rpc.MSTManagerService_ManageServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Printf("Received %s", in.GetVal())
		err = stream.Send(&rpc.MSTManageCommand{Val: in.GetVal()})
		if err != nil {
			return err
		}
	}
}
