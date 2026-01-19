package gossip

import (
	"encoding/json"
	"math/rand"
	"net"
	"sync"
	"time"
)

// Gossiper is the main entry point for using the gossip protocol.
type Gossiper struct {
	name      string
	members   *MembershipList
	transport Transport
	stop      chan struct{}
	self      *Node
	wg        sync.WaitGroup
}

// NewGossiper creates a new gossiper.
func NewGossiper(name, listenAddr string, peers []string, transport Transport) (*Gossiper, error) {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	self := &Node{
		Addr:  addr,
		State: Alive,
		LastUpdated: time.Now(),
	}

	g := &Gossiper{
		name:      name,
		members:   NewMembershipList(),
		transport: transport,
		stop:      make(chan struct{}),
		self:      self,
	}

	g.members.Add(self)

	// Add initial peers
	for _, peerAddr := range peers {
		peerUDPAddr, err := net.ResolveUDPAddr("udp", peerAddr)
		if err != nil {
			return nil, err
		}
		peerNode := &Node{
			Addr:  peerUDPAddr,
			State: Alive,
			LastUpdated: time.Now(),
		}
		g.members.Add(peerNode)
	}

	return g, nil
}

// Start starts the gossip loops.
func (g *Gossiper) Start() {
	g.wg.Add(3)
	go g.pingLoop()
	go g.syncLoop()
	go g.listen()
}

// Stop stops the gossip loops.
func (g *Gossiper) Stop() {
	close(g.stop)
	g.wg.Wait()
}

// SetPayload sets the payload for the local node.
func (g *Gossiper) SetPayload(payload string) {
	g.self.Payload = payload
	g.self.LastUpdated = time.Now()
}

// Members returns all nodes in the membership list.
func (g *Gossiper) Members() []*Node {
	return g.members.All()
}

func (g *Gossiper) listen() {
	defer g.wg.Done()
	for {
		select {
		case data := <-g.transport.Read():
			g.handleMessage(data)
		case <-g.stop:
			return
		}
	}
}

func (g *Gossiper) pingLoop() {
	defer g.wg.Done()
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
	defer g.wg.Done()
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
		return
	}

	// Send the message.
	if err := g.transport.Write(data, node.Addr.String()); err != nil {
		// Log.Printf("[%s] failed to send message to %s: %v", g.name, node.Addr.String(), err)
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
		return
	}

	msg := &Message{
		Type:    Sync,
		Payload: payload,
	}

	// Encode the message.
	data, err := msg.Encode()
	if err != nil {
		return
	}

	// Send the message.
	if err := g.transport.Write(data, node.Addr.String()); err != nil {
		// Log.Printf("[%s] failed to send message to %s: %v", g.name, node.Addr.String(), err)
	}
}

func (g *Gossiper) handleMessage(data []byte) {
	msg, err := Decode(data)
	if err != nil {
		return
	}

	switch msg.Type {
	case Ping:
		// Do nothing for now.
		// In the future, this could be used to request a sync.
	case Sync:
		var nodes []*Node
		if err := json.Unmarshal(msg.Payload, &nodes); err != nil {
			return
		}
		g.members.Merge(nodes)
	}
}