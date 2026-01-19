package main

import (
	"fmt"
	"log"
	"time"

	"github.com/princetheprogrammer/go-gossip/pkg/gossip"
)

func main() {
	// Create transport for node 1
	transport1, err := gossip.NewUDPTransport("127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer transport1.Stop()

	// Create node 1
	node1, err := gossip.NewGossiper("127.0.0.1:8080", transport1)
	if err != nil {
		log.Fatal(err)
	}
	defer node1.Stop()

	// Create transport for node 2
	transport2, err := gossip.NewUDPTransport("127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}
	defer transport2.Stop()

	// Create node 2
	node2, err := gossip.NewGossiper("127.0.0.1:8081", transport2)
	if err != nil {
		log.Fatal(err)
	}
	defer node2.Stop()

	// Cross-seed the nodes.
	if err := node1.AddNode("127.0.0.1:8081"); err != nil {
		log.Printf("failed to add node: %v", err)
	}
	if err := node2.AddNode("127.0.0.1:8080"); err != nil {
		log.Printf("failed to add node: %v", err)
	}

	// Set payloads.
	node1.SetPayload([]byte("node1"))
	node2.SetPayload([]byte("node2"))

	// Start the gossipers.
	node1.Start()
	node2.Start()

	// Wait for the nodes to gossip.
	time.Sleep(10 * time.Second)

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
