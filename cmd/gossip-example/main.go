package main

import (
	"fmt"
	"log"
	"time"

	"github.com/princetheprogrammer/go-gossip/pkg/gossip"
)

func main() {
	// Secret key for encryption. Must be 16, 24, or 32 bytes long.
	key := []byte("a very secret key 12345678901234") // 32 bytes

	// Create transport for node 1
	udpTransport1, err := gossip.NewUDPTransport("127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	transport1, err := gossip.NewSecureTransport(udpTransport1, key)
	if err != nil {
		log.Fatal(err)
	}
	defer transport1.Stop()

	// Create node 1
	node1, err := gossip.NewGossiper("node1", "127.0.0.1:8080", []string{"127.0.0.1:8081"}, transport1)
	if err != nil {
		log.Fatal(err)
	}
	defer node1.Stop()

	// Create transport for node 2
	udpTransport2, err := gossip.NewUDPTransport("127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}
	transport2, err := gossip.NewSecureTransport(udpTransport2, key)
	if err != nil {
		log.Fatal(err)
	}
	defer transport2.Stop()

	// Create node 2
	node2, err := gossip.NewGossiper("node2", "127.0.0.1:8081", []string{"127.0.0.1:8080"}, transport2)
	if err != nil {
		log.Fatal(err)
	}
	defer node2.Stop()

	// Set payloads.
	node1.SetPayload("node1")
	node2.SetPayload("node2")

	// Start the gossipers.
	node1.Start()
	node2.Start()

	// Wait for the nodes to gossip.
	time.Sleep(30 * time.Second)

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