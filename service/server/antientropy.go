package server

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

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
	store := r.tree.NodeStore()
	roundUUIDBytes, err := r.roundUUID.MarshalBinary()
	if err != nil {
		panic(err)
	}

	// Create connection to other node
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("Error connecting to %s when gossiping: %s", address, err)
		endRoundFunc()
		return
	}
	client := rpc.NewMSTManagerServiceClient(conn)

	// Start the round
	res, err := client.RoundStart(r.ctx, &rpc.MSTRoundStartRequest{
		RootHash:  rootHash,
		RoundUuid: roundUUIDBytes,
	})
	if err != nil {
		log.Printf("Error starting round to %s: %s", address, err)
		endRoundFunc()
		return
	}

	// Iterate over steps of the anti-entropy algorithm
	hashes := res.GetHashes()
	for len(hashes) > 0 {
		rpcNodes := make([]*rpc.MSTNode, 0, len(hashes))
		hashStrs := make([]string, 0, len(hashes))
		for _, hash := range hashes {
			node := store.Get(hash)
			if node == nil {
				panic(fmt.Sprintf("Missing node for hash %s", hex.EncodeToString(hash)))
			}
			rpcNodes = append(rpcNodes, nodeToRPC(node))
			hashStrs = append(hashStrs, hex.EncodeToString(hash))
		}
		log.Printf("Sending nodes for hashes: %s", strings.Join(hashStrs, ", "))
		res, err := client.RoundStep(r.ctx, &rpc.MSTRoundStepRequest{
			RoundUuid: roundUUIDBytes,
			Nodes:     rpcNodes,
		})
		if err != nil {
			log.Printf("Error stepping round to %s: %s", address, err)
			endRoundFunc()
			return
		}
		hashes = res.GetHashes()
	}
	endRoundFunc()
}
