package gossip

import (
	"log"
	"sync"
)

// MembershipList stores the state of all nodes in the cluster.
type MembershipList struct {
	mu    sync.RWMutex
	nodes map[string]*Node
}

// NewMembershipList creates a new membership list.
func NewMembershipList() *MembershipList {
	return &MembershipList{
		nodes: make(map[string]*Node),
	}
}

// Add adds a new node to the list.
func (m *MembershipList) Add(node *Node) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addOrUpdate(node)
}

// Merge merges a list of nodes with the local list.
func (m *MembershipList) Merge(nodes []*Node) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, node := range nodes {
		m.addOrUpdate(node)
	}
}

func (m *MembershipList) addOrUpdate(node *Node) {
	existing, ok := m.nodes[node.Addr.String()]
	if !ok {
		log.Printf("Adding new node: %s", node.Addr.String())
		m.nodes[node.Addr.String()] = node
		return
	}

	log.Printf("Updating node: %s", node.Addr.String())
	log.Printf("Existing: LastUpdated=%v, Payload=%s", existing.LastUpdated, existing.Payload)
	log.Printf("Incoming: LastUpdated=%v, Payload=%s", node.LastUpdated, node.Payload)

	// If the incoming node is older or equally old, ignore it.
	if !node.LastUpdated.After(existing.LastUpdated) {
		log.Printf("Incoming node is not newer. Ignoring.")
		return
	}

	log.Printf("Incoming node is newer. Updating.")
	existing.State = node.State
	existing.LastUpdated = node.LastUpdated
	if node.Payload != "" {
		log.Printf("Incoming payload is not empty. Updating payload.")
		existing.Payload = node.Payload
	} else {
		log.Printf("Incoming payload is empty. Preserving existing payload.")
	}
}
// Get returns a node from the list.
func (m *MembershipList) Get(addr string) (*Node, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	node, ok := m.nodes[addr]
	return node, ok
}

// All returns all nodes in the list.
func (m *MembershipList) All() []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	nodes := make([]*Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}
