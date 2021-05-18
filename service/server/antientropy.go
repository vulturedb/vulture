package server

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/vulturedb/vulture/mst"
	"github.com/vulturedb/vulture/service/rpc"
)

// EndRoundFunc is the signature for ending the anti entropy round
// When called, it tells the caller that the round is over so that it can clean
// up whatever needs to be cleaned up
type EndRoundFunc func()

// AntiEntropyRound represents a round of anti-entropy with another vulture
// host.
type AntiEntropyRound struct {
	roundUUID uuid.UUID
	ctx       context.Context
	peer      Peer
	tree      *mst.MerkleSearchTree
	cancelFn  context.CancelFunc
}

// NewAntiEntropyRound creates new AntiEntropyRound
func NewAntiEntropyRound(peer Peer, tree *mst.MerkleSearchTree) AntiEntropyRound {
	ctx, cancelFn := context.WithCancel(context.Background())
	roundUUID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return AntiEntropyRound{roundUUID, ctx, peer, tree, cancelFn}
}

func (r AntiEntropyRound) runRound(endRoundFunc EndRoundFunc) {
	address := fmt.Sprintf("%s:%d", r.peer.Hostname, r.peer.Port)
	rootHash := r.tree.RootHash()
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("Error connecting to %s when gossiping: %s", address, err)
		endRoundFunc()
		return
	}
	client := rpc.NewMSTManagerServiceClient(conn)
	_, err = client.Gossip(r.ctx, &rpc.MSTGossipRequest{RootHash: rootHash})
	if err != nil {
		log.Printf("Error gossiping to %s: %s", address, err)
		endRoundFunc()
		return
	}
	endRoundFunc()
}
