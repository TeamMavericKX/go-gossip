package gossip

import (
	"net"
	"testing"
	"time"
)

func TestMembershipList_Add(t *testing.T) {
	ml := NewMembershipList()
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
	node := &Node{
		Addr:        addr,
		State:       Alive,
		LastUpdated: time.Now(),
		Payload:     "test-payload",
	}

	ml.Add(node)

	retrievedNode, ok := ml.Get(addr.String())
	if !ok {
		t.Fatal("Node not found after adding")
	}
	if retrievedNode.Addr.String() != node.Addr.String() {
		t.Errorf("Expected address %s, got %s", node.Addr.String(), retrievedNode.Addr.String())
	}
	if retrievedNode.Payload != node.Payload {
		t.Errorf("Expected payload %s, got %s", node.Payload, retrievedNode.Payload)
	}
}

func TestMembershipList_Merge(t *testing.T) {
	ml := NewMembershipList()

	// Add an initial node
	addr1 := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
	node1 := &Node{
		Addr:        addr1,
		State:       Alive,
		LastUpdated: time.Now(),
		Payload:     "payload1",
	}
	ml.Add(node1)

	// Create a newer version of node1 with updated payload
	node1Newer := &Node{
		Addr:        addr1,
		State:       Alive,
		LastUpdated: time.Now().Add(1 * time.Second),
		Payload:     "payload1-updated",
	}

	// Create an older version of node1 (should be ignored)
	node1Older := &Node{
		Addr:        addr1,
		State:       Alive,
		LastUpdated: time.Now().Add(-1 * time.Second),
		Payload:     "payload1-old",
	}

	// Create a new node2
	addr2 := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8081}
	node2 := &Node{
		Addr:        addr2,
		State:       Alive,
		LastUpdated: time.Now(),
		Payload:     "payload2",
	}

	// Test merging: node1Newer should update, node1Older should be ignored, node2 should be added
	nodesToMerge := []*Node{node1Newer, node1Older, node2}
	ml.Merge(nodesToMerge)

	// Verify node1 updated
	retrievedNode1, ok := ml.Get(addr1.String())
	if !ok {
		t.Fatal("Node1 not found after merge")
	}
	if !retrievedNode1.LastUpdated.Equal(node1Newer.LastUpdated) {
		t.Errorf("Node1 LastUpdated not updated. Expected %v, got %v", node1Newer.LastUpdated, retrievedNode1.LastUpdated)
	}
	if retrievedNode1.Payload != node1Newer.Payload {
		t.Errorf("Node1 Payload not updated. Expected %s, got %s", node1Newer.Payload, retrievedNode1.Payload)
	}

	// Verify node2 added
	retrievedNode2, ok := ml.Get(addr2.String())
	if !ok {
		t.Fatal("Node2 not found after merge")
	}
	if retrievedNode2.Payload != node2.Payload {
		t.Errorf("Expected node2 payload %s, got %s", node2.Payload, retrievedNode2.Payload)
	}

	// Test merging with empty payload on newer node
	node1NewerEmptyPayload := &Node{
		Addr:        addr1,
		State:       Alive,
		LastUpdated: node1Newer.LastUpdated.Add(1 * time.Second),
		Payload:     "", // Empty payload
	}
	ml.Merge([]*Node{node1NewerEmptyPayload})
	retrievedNode1, _ = ml.Get(addr1.String())
	if retrievedNode1.Payload != node1Newer.Payload {
		t.Errorf("Expected node1 payload to be preserved as %s, got %s", node1Newer.Payload, retrievedNode1.Payload)
	}
}
