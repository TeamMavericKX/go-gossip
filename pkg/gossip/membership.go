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
	log.Printf("Existing: LastUpdated=%v, Payload=%s", existing.LastUpdated, string(existing.Payload))
	log.Printf("Incoming: LastUpdated=%v, Payload=%s", node.LastUpdated, string(node.Payload))

	if node.LastUpdated.After(existing.LastUpdated) {
		log.Printf("Incoming node is newer. Updating.")
		if len(node.Payload) == 0 {
			log.Printf("Incoming payload is empty. Preserving existing payload.")
			node.Payload = existing.Payload
		}
		m.nodes[node.Addr.String()] = node
	} else {
		log.Printf("Incoming node is not newer. Ignoring.")
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
