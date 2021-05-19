package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"log"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/pkg/errors"

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

func (s *MSTServer) mergeTree(tree *mst.MerkleSearchTree) {
	log.Printf("Merging tree")
	s.treeLock.Lock()
	defer s.treeLock.Unlock()
	tree, err := s.tree.Merge(tree)
	if err != nil {
		panic(err)
	}
	s.tree = tree
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

type antiEntropyDestRound struct {
	tree     *mst.MerkleSearchTree
	rootHash []byte
}

// MSTManagerServer stores data required for managing the Vulture server
// TODO: right now we don't do any cleanup of destination rounds, so there may
// be a memory leak if we have stray rounds. We could include a garbage
// collection round at some point here.
type MSTManagerServer struct {
	server                    *MSTServer
	kr                        mst.KeyReader
	vr                        mst.ValueReader
	antiEntropyDestRounds     map[uuid.UUID]antiEntropyDestRound
	antiEntropyDestRoundsLock sync.RWMutex
}

// NewMSTManagerServer creates a new Vulture management server
func NewMSTManagerServer(
	server *MSTServer,
	kr mst.KeyReader,
	vr mst.ValueReader,
) *MSTManagerServer {
	return &MSTManagerServer{
		server:                server,
		kr:                    kr,
		vr:                    vr,
		antiEntropyDestRounds: make(map[uuid.UUID]antiEntropyDestRound),
	}
}

func (s *MSTManagerServer) getMissingHashes(
	roundUUID uuid.UUID,
	round antiEntropyDestRound,
) *rpc.MSTRoundStepResponse {
	hashes := mst.FindMissingNodes(round.tree.NodeStore(), round.rootHash)
	if len(hashes) == 0 {
		s.antiEntropyDestRoundsLock.Lock()
		delete(s.antiEntropyDestRounds, roundUUID)
		s.antiEntropyDestRoundsLock.Unlock()
		s.server.mergeTree(round.tree)
	}
	return &rpc.MSTRoundStepResponse{Hashes: hashes}
}

// RoundStart starts a round of anti entropy
func (s *MSTManagerServer) RoundStart(
	ctx context.Context,
	in *rpc.MSTRoundStartRequest,
) (*rpc.MSTRoundStepResponse, error) {
	rootHash := in.GetRootHash()
	roundUUID, err := uuid.FromBytes(in.GetRoundUuid())
	if err != nil {
		return nil, err
	}

	// Create the round on the destination side
	tree := s.server.getTree().WithRoot(rootHash)
	round := antiEntropyDestRound{tree, rootHash}
	s.antiEntropyDestRoundsLock.Lock()
	s.antiEntropyDestRounds[roundUUID] = round
	s.antiEntropyDestRoundsLock.Unlock()

	log.Printf(
		"Starting round for %s with roundUUID %s",
		hex.EncodeToString(rootHash),
		roundUUID.String(),
	)
	return s.getMissingHashes(roundUUID, round), nil
}

func (s *MSTManagerServer) updateNodes(
	roundUUID uuid.UUID,
	nodes []*mst.Node,
) (antiEntropyDestRound, error) {
	// This locks for a long time, so maybe find a way to have more granular
	// locks.
	s.antiEntropyDestRoundsLock.Lock()
	defer s.antiEntropyDestRoundsLock.Unlock()
	if _, exists := s.antiEntropyDestRounds[roundUUID]; !exists {
		return antiEntropyDestRound{}, errors.Errorf("Missing tree for round %s", roundUUID.String())
	}
	round := s.antiEntropyDestRounds[roundUUID]
	tree := round.tree
	store := tree.NodeStore()
	for _, node := range nodes {
		store, _ = store.Put(node)
	}
	tree = tree.WithNodeStore(store)
	round = antiEntropyDestRound{tree, round.rootHash}
	s.antiEntropyDestRounds[roundUUID] = round
	return round, nil
}

// RoundStep does a step of anti entropy
func (s *MSTManagerServer) RoundStep(
	ctx context.Context,
	in *rpc.MSTRoundStepRequest,
) (*rpc.MSTRoundStepResponse, error) {
	roundUUID, err := uuid.FromBytes(in.GetRoundUuid())
	if err != nil {
		return nil, err
	}

	rpcNodes := in.GetNodes()
	mstNodes := make([]*mst.Node, 0, len(rpcNodes))
	for _, rpcNode := range rpcNodes {
		mstNode, err := nodeFromRPC(rpcNode, s.kr, s.vr)
		if err != nil {
			return nil, err
		}
		mstNodes = append(mstNodes, mstNode)
	}

	round, err := s.updateNodes(roundUUID, mstNodes)
	if err != nil {
		return nil, err
	}

	log.Printf(
		"Stepping round for %s with roundUUID %s",
		hex.EncodeToString(round.rootHash),
		roundUUID.String(),
	)

	return s.getMissingHashes(roundUUID, round), nil
}
