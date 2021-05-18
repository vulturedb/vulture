package server

import "github.com/vulturedb/vulture/mst"

// AntiEntropyRound represents a round of anti-entropy with another vulture
// host.
type AntiEntropyRound struct {
  peer Peer
}

// NewAntiEntropyRound creates new AntiEntropyRound
func NewAntiEntropyRound(peer Peer, tree *mst.MerkleSearchTree) AntiEntropyRound {
  return AntiEntropyRound{peer}
}

func (r AntiEntropyRound) runRound() {
}
