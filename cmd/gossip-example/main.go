package main

import (
	"fmt"
	"log"
	"time"

	"github.com/princetheprogrammer/go-gossip/pkg/gossip"
)

func main() {
	// Create two nodes.
	node1, err := createNode("127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer node1.Stop()

	node2, err := createNode("127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}
	defer node2.Stop()

	// Cross-seed the nodes.
	if err := node1.AddNode("127.0.0.1:8081"); err != nil {
		log.Fatal(err)
	}
	if err := node2.AddNode("127.0.0.1:8080"); err != nil {
		log.Fatal(err)
	}

	// Set payloads.
	node1.SetPayload([]byte("node1"))
	node2.SetPayload([]byte("node2"))

	// Add a small delay to ensure payload is set before first sync.
	time.Sleep(1 * time.Second)

	// Start the gossipers.
	node1.Start()
	node2.Start()

	// Wait for the nodes to gossip.
	time.Sleep(5 * time.Second)

	// Print the membership lists.
	fmt.Println("Node 1 members:")
	for _, member := range node1.Members() {
		fmt.Printf("- %s, payload: %s\n", member.Addr, string(member.Payload))
	}

	fmt.Println("Node 2 members:")
	for _, member := range node2.Members() {
		fmt.Printf("- %s, payload: %s\n", member.Addr, string(member.Payload))
	}
}

// Gossiper is a wrapper around the gossip.Gossiper struct.
type Gossiper struct {
	*gossip.Gossiper
	transport *gossip.UDPTransport
}

// Stop stops the gossiper and the transport.
func (g *Gossiper) Stop() {
	g.Gossiper.Stop()
	g.transport.Stop()
}

func createNode(listenAddr string) (*Gossiper, error) {
	transport, err := gossip.NewUDPTransport(listenAddr)
	if err != nil {
		return nil, err
	}

	g, err := gossip.NewGossiper(listenAddr, transport)
	if err != nil {
		return nil, err
	}

	return &Gossiper{
		Gossiper:  g,
		transport: transport,
	}, nil
}
