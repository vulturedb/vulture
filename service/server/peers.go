package server

import "sync"

type Peer struct {
	Hostname string
	Port     int
}

type PeerSelectionStrategy interface {
	Select(peers map[Peer]bool) []Peer
}

type SelectAll struct{}

func (s *SelectAll) Select(peerSet map[Peer]bool) []Peer {
	peers := make([]Peer, 0, len(peerSet))
	for peer := range peerSet {
		peers = append(peers, peer)
	}
	return peers
}

type Peers struct {
	peerSet           map[Peer]bool
	mutex             sync.RWMutex
	selectionStrategy PeerSelectionStrategy
}

func NewPeers(selectionStrategy PeerSelectionStrategy) *Peers {
	return &Peers{peerSet: map[Peer]bool{}, selectionStrategy: selectionStrategy}
}

func (ps *Peers) exists(p Peer) bool {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.peerSet[p]
}

func (ps *Peers) Add(hostname string, port int) {
	p := Peer{hostname, port}
	if ps.exists(p) {
		return
	}
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	ps.peerSet[p] = true
}

func (ps *Peers) Select() []Peer {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	return ps.selectionStrategy.Select(ps.peerSet)
}
