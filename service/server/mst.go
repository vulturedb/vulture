package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/vulturedb/vulture/mst"
	"github.com/vulturedb/vulture/service/rpc"
)

// MSTServer stores all local data required for running the Vulture server
type MSTServer struct {
	tree                  *mst.MerkleSearchTree
	peers                 *Peers
	antiEntropyRounds     map[Peer]AntiEntropyRound
	treeLock              sync.RWMutex
	antiEntropyRoundsLock sync.RWMutex
}

// NewMSTServer creates a new Vulture server
func NewMSTServer(tree *mst.MerkleSearchTree, peers *Peers) *MSTServer {
	return &MSTServer{tree: tree, peers: peers, antiEntropyRounds: make(map[Peer]AntiEntropyRound)}
}

func (s *MSTServer) getTree() *mst.MerkleSearchTree {
	s.treeLock.RLock()
	defer s.treeLock.RUnlock()
	return s.tree
}

func (s *MSTServer) gossip(ctx context.Context, rootHash []byte) {
	peers := s.peers.Select()
	for _, peer := range peers {
		// TODO: should keep these connections around to save on round-trips, maybe there's a connection
		// pool library we can use.
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

func (s *MSTServer) createEndRoundFunc(peer Peer) EndRoundFunc {
	return func() {
		s.antiEntropyRoundsLock.Lock()
		defer s.antiEntropyRoundsLock.Unlock()
		_, hasRound := s.antiEntropyRounds[peer]
		if hasRound {
			delete(s.antiEntropyRounds, peer)
		}
	}
}

// Get returns the value for a given key
func (s *MSTServer) Get(ctx context.Context, in *rpc.MSTGetRequest) (*rpc.MSTGetResponse, error) {
	key := in.GetKey()
	tree := s.getTree()
	val := tree.Get(mst.UInt32(key))
	var primVal uint32
	if val == nil {
		primVal = 0
	} else {
		primVal = uint32(val.(mst.UInt32))
	}
	log.Printf("Get %d: %d", key, primVal)
	return &rpc.MSTGetResponse{Value: primVal}, nil
}

func (s *MSTServer) runAntiEntropy() {
	peers := s.peers.Select()
	rounds := make([]AntiEntropyRound, 0)
	s.antiEntropyRoundsLock.Lock()
	for _, peer := range peers {
		_, hasRound := s.antiEntropyRounds[peer]
		if !hasRound {
			round := NewAntiEntropyRound(peer, s.getTree())
			s.antiEntropyRounds[peer] = round
			rounds = append(rounds, round)
		}
	}
	s.antiEntropyRoundsLock.Unlock()
	for _, round := range rounds {
		go round.runRound(s.createEndRoundFunc(round.peer))
	}
}

// Put inserts the given value for the given key
func (s *MSTServer) Put(ctx context.Context, in *rpc.MSTPutRequest) (*empty.Empty, error) {
	key := in.GetKey()
	val := in.GetValue()
	initialRootHash := s.tree.RootHash()
	s.treeLock.Lock()
	s.tree = s.tree.Put(mst.UInt32(key), mst.UInt32(val))
	s.treeLock.Unlock()
	log.Printf("Put %d: %d", key, val)
	newRootHash := s.tree.RootHash()
	if bytes.Compare(newRootHash, initialRootHash) != 0 {
		go s.runAntiEntropy()
	}
	return &empty.Empty{}, nil
}

// MSTManagerServer stores data required for managing the Vulture server
type MSTManagerServer struct {
}

// NewMSTManagerServer creates a new Vulture management server
func NewMSTManagerServer() *MSTManagerServer {
	return &MSTManagerServer{}
}

// Gossip defines the handler for the Gossip endpoint of the Vulture management
// server
func (s *MSTManagerServer) Gossip(
	ctx context.Context,
	in *rpc.MSTGossipRequest,
) (*empty.Empty, error) {
	log.Printf("Received gossip for %s", hex.EncodeToString(in.GetRootHash()))
	return &empty.Empty{}, nil
}

// GetNodes returns the nodes from the local tree
func (s *MSTManagerServer) GetNodes(
	ctx context.Context,
	in *rpc.MSTGetNodesRequest,
) (*rpc.MSTGetNodesResponse, error) {
	return nil, nil
}
