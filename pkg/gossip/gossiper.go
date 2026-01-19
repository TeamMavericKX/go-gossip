package gossip

import (
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"time"
)

// Gossiper is the main entry point for using the gossip protocol.
type Gossiper struct {
	members   *MembershipList
	transport Transport
	stop      chan struct{}
	self      *Node
}

// NewGossiper creates a new gossiper.
func NewGossiper(listenAddr string, transport Transport) (*Gossiper, error) {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	self := &Node{
		Addr:  addr,
		State: Alive,
	}

	g := &Gossiper{
		members:   NewMembershipList(),
		transport: transport,
		stop:      make(chan struct{}),
		self:      self,
	}

	g.members.Add(self)

	return g, nil
}

// Start starts the gossip loops.
func (g *Gossiper) Start() {
	go g.pingLoop()
	go g.syncLoop()
	go g.listen()
}

// Stop stops the gossip loops.
func (g *Gossiper) Stop() {
	close(g.stop)
}

func (g *Gossiper) pingLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.sendPing()
		case <-g.stop:
			return
		}
	}
}

func (g *Gossiper) syncLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.sendSync()
		case <-g.stop:
			return
		}
	}
}

func (g *Gossiper) sendPing() {
	nodes := g.members.All()
	if len(nodes) <= 1 {
		return
	}

	// Select a random node to gossip to (excluding self).
	var node *Node
	for {
		node = nodes[rand.Intn(len(nodes))]
		if node.Addr.String() != g.self.Addr.String() {
			break
		}
	}

	msg := &Message{
		Type: Ping,
	}

	// Encode the message.
	data, err := msg.Encode()
	if err != nil {
		log.Printf("failed to encode message: %v", err)
		return
	}

	// Send the message.
	if err := g.transport.Write(data, node.Addr.String()); err != nil {
		log.Printf("failed to send message to %s: %v", node.Addr.String(), err)
	}
}

func (g *Gossiper) sendSync() {
	g.self.LastUpdated = time.Now()

	nodes := g.members.All()
	if len(nodes) <= 1 {
		return
	}

	// Select a random node to gossip to (excluding self).
	var node *Node
	for {
		node = nodes[rand.Intn(len(nodes))]
		if node.Addr.String() != g.self.Addr.String() {
			break
		}
	}

	// Create a sync message with the membership list.
	payload, err := json.Marshal(g.members.All())
	if err != nil {
		log.Printf("failed to marshal membership list: %v", err)
		return
	}

	msg := &Message{
		Type:    Sync,
		Payload: payload,
	}

	// Encode the message.
	data, err := msg.Encode()
	if err != nil {
		log.Printf("failed to encode message: %v", err)
		return
	}

	// Send the message.
	if err := g.transport.Write(data, node.Addr.String()); err != nil {
		log.Printf("failed to send message to %s: %v", node.Addr.String(), err)
	}
}

func (g *Gossiper) handleMessage(data []byte) {
	msg, err := Decode(data)
	if err != nil {
		log.Printf("failed to decode message: %v", err)
		return
	}

	switch msg.Type {
	case Ping:
		// Do nothing for now.
		// In the future, this could be used to request a sync.
	case Sync:
		var nodes []*Node
		if err := json.Unmarshal(msg.Payload, &nodes); err != nil {
			log.Printf("failed to unmarshal sync payload: %v", err)
			return
		}
		g.members.Merge(nodes)
	}
}
