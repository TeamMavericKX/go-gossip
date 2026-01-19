package gossip

import (
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
	m.nodes[node.Addr.String()] = node
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
