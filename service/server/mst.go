package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/vulturedb/vulture/mst"
	"github.com/vulturedb/vulture/service/rpc"
)

type MSTServer struct {
	tree  *mst.MerkleSearchTree
	peers *Peers
}

func NewMSTServer(tree *mst.MerkleSearchTree, peers *Peers) *MSTServer {
	return &MSTServer{tree, peers}
}

func (s *MSTServer) gossip(ctx context.Context, rootHash []byte) {
	peers := s.peers.Select()
	for _, peer := range peers {
		// TODO: should keep these connections around to save on round-trips, maybe there's a connection
		// pool library we can use
		address := fmt.Sprintf("%s:%d", peer.Hostname, peer.Port)
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Printf("Error connecting to %s when gossiping: %s", address, err)
		}
		client := rpc.NewMSTManagerServiceClient(conn)
		_, err = client.Gossip(ctx, &rpc.MSTGossipRequest{RootHash: rootHash})
		if err != nil {
			log.Printf("Error gossiping to %s: %s", address, err)
		}
	}
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
	initialRootHash := s.tree.RootHash()
	s.tree.Put(mst.UInt32(key), mst.UInt32(val))
	log.Printf("Put %d: %d", key, val)
	newRootHash := s.tree.RootHash()
	if bytes.Compare(newRootHash, initialRootHash) != 0 {
		s.gossip(ctx, newRootHash)
	}
	return &empty.Empty{}, nil
}

type MSTManagerServer struct {
}

func NewMSTManagerServer() *MSTManagerServer {
	return &MSTManagerServer{}
}

func (s *MSTManagerServer) Gossip(
	ctx context.Context,
	in *rpc.MSTGossipRequest,
) (*empty.Empty, error) {
	log.Printf("Received gossip for %s", hex.EncodeToString(in.GetRootHash()))
	return &empty.Empty{}, nil
}

func (s *MSTManagerServer) GetNodes(
	ctx context.Context,
	in *rpc.MSTGetNodesRequest,
) (*rpc.MSTGetNodesResponse, error) {
	return nil, nil
}
