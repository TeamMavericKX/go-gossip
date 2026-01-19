package gossip

import (
	"encoding/json"
	"log"
	"math/rand"
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
	go g.listen()
}

// Stop stops the gossip loop.
func (g *Gossiper) Stop() {
	close(g.stop)
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
	nodes := g.members.All()
	if len(nodes) == 0 {
		return
	}

	// Select a random node to gossip to.
	node := nodes[rand.Intn(len(nodes))]

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
			g.members.Add(node)
		}
	}
}
