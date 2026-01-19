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

// Start starts the gossip loop.
func (g *Gossiper) Start() {
	go g.gossipLoop()
	go g.listen()
}

// Stop stops the gossip loop.
func (g *Gossiper) Stop() {
	close(g.stop)
}

// AddNode adds a new node to the membership list.
func (g *Gossiper) AddNode(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	node := &Node{
		Addr:  udpAddr,
		State: Alive,
	}

	g.members.Add(node)
	return nil
}

// SetPayload sets the payload for the local node.
func (g *Gossiper) SetPayload(payload []byte) {
	g.self.Payload = payload
	g.self.LastUpdated = time.Now()
}

// Members returns all nodes in the membership list.
func (g *Gossiper) Members() []*Node {
	return g.members.All()
}

func (g *Gossiper) listen() {
	for {
		select {
		case data := <-g.transport.Read():
			g.handleMessage(data)
		case <-g.stop:
			return
		}
	}
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

	// Create a ping message with the membership list.
	payload, err := json.Marshal(g.members.All())
	if err != nil {
		log.Printf("failed to marshal membership list: %v", err)
		return
	}

	msg := &Message{
		Type:    Ping,
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
		var nodes []*Node
		if err := json.Unmarshal(msg.Payload, &nodes); err != nil {
			log.Printf("failed to unmarshal ping payload: %v", err)
			return
		}
		for _, node := range nodes {
			if node.Addr.String() == g.self.Addr.String() {
				continue
			}
			g.members.Add(node)
		}
	}
}
