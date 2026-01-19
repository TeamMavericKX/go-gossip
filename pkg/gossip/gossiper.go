package gossip

import (
	"encoding/json"
	"log"
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
func NewGossiper(name, listenAddr string, transport Transport) (*Gossiper, error) {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	self := &Node{
		Addr:  addr,
		State: Alive,
	}

	g := &Gossiper{
		name:      name,
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

			log.Printf("[%s] failed to encode message: %v", g.name, err)

			return

		}

	

		// Send the message.

		if err := g.transport.Write(data, node.Addr.String()); err != nil {

			log.Printf("[%s] failed to send message to %s: %v", g.name, node.Addr.String(), err)

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

			log.Printf("[%s] failed to marshal membership list: %v", g.name, err)

			return

		}

	

		msg := &Message{

			Type:    Sync,

			Payload: payload,

		}

	

		// Encode the message.

		data, err := msg.Encode()

		if err != nil {

			log.Printf("[%s] failed to encode message: %v", g.name, err)

			return

		}

	

		// Send the message.

		if err := g.transport.Write(data, node.Addr.String()); err != nil {

			log.Printf("[%s] failed to send message to %s: %v", g.name, node.Addr.String(), err)

		}

	}

	

	func (g *Gossiper) handleMessage(data []byte) {

		msg, err := Decode(data)

		if err != nil {

			log.Printf("[%s] failed to decode message: %v", g.name, err)

			return

		}

	

		switch msg.Type {

		case Ping:

			// Do nothing for now.

			// In the future, this could be used to request a sync.

		case Sync:

			var nodes []*Node

			if err := json.Unmarshal(msg.Payload, &nodes); err != nil {

				log.Printf("[%s] failed to unmarshal sync payload: %v", g.name, err)

				return

			}

			g.members.Merge(nodes)

		}

	}

	