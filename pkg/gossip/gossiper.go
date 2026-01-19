package gossip

import (
	"time"
)

// Gossiper is the main entry point for using the gossip protocol.
type Gossiper struct {
	members   *MembershipList
	transport Transport
	stop      chan struct{}
}

// NewGossiper creates a new gossiper.
func NewGossiper(transport Transport) *Gossiper {
	return &Gossiper{
		members:   NewMembershipList(),
		transport: transport,
		stop:      make(chan struct{}),
	}
}

// Start starts the gossip loop.
func (g *Gossiper) Start() {
	go g.gossipLoop()
}

// Stop stops the gossip loop.
func (g *Gossiper) Stop() {
	close(g.stop)
}

func (g *Gossiper) gossipLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.gossip()
		case <-g.stop:
			return
		}
	}
}

func (g *Gossiper) gossip() {
	// For now, we don't have any nodes to gossip to.
	// This will be implemented later.
}
