package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/princetheprogrammer/go-gossip/pkg/gossip"
)

func main() {
	// Create two nodes.
	node1, err := createNode("127.0.0.1:8080", "127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}
	defer node1.Stop()

	node2, err := createNode("127.0.0.1:8081", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer node2.Stop()

	// Start the gossipers.
	node1.Start()
	node2.Start()

	// Wait for the nodes to gossip.
	time.Sleep(5 * time.Second)

	// Print the membership lists.
	fmt.Println("Node 1 members:")
	for _, member := range node1.Members() {
		fmt.Printf("- %s\n", member.Addr)
	}

	fmt.Println("Node 2 members:")
	for _, member := range node2.Members() {
		fmt.Printf("- %s\n", member.Addr)
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

// Members returns the members of the cluster.
func (g *Gossiper) Members() []*gossip.Node {
	return g.All()
}

func createNode(listenAddr, remoteAddr string) (*Gossiper, error) {
	transport, err := gossip.NewUDPTransport(listenAddr)
	if err != nil {
		return nil, err
	}

	g := gossip.NewGossiper(transport)

	// Add the remote node to the membership list.
	addr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if err != nil {
		return nil, err
	}
	g.Add(&gossip.Node{
		Addr:  addr,
		State: gossip.Alive,
	})

	return &Gossiper{
		Gossiper:  g,
		transport: transport,
	},
	nil
}
