package gossip

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// State represents the state of a node.
type State int

const (
	// Alive is the state of a node that is alive and well.
	Alive State = iota
	// Suspected is the state of a node that is suspected to be dead.
	Suspected
	// Dead is the state of a node that is confirmed to be dead.
	Dead
)

func (s State) String() string {
	switch s {
	case Alive:
		return "alive"
	case Suspected:
		return "suspected"
	case Dead:
		return "dead"
	default:
		return "unknown"
	}
}

// Node represents a single member in the cluster.
type Node struct {
	// Addr is the network address of the node.
	Addr net.Addr
	// State is the current state of the node.
	State State
	// LastUpdated is the time when the node's state was last updated.
	LastUpdated time.Time
	// Payload is a custom payload associated with the node.
	Payload []byte
}

func (n *Node) String() string {
	return fmt.Sprintf("Node{Addr: %s, State: %s, Payload: %s, LastUpdated: %v}", n.Addr, n.State, string(n.Payload), n.LastUpdated)
}

type nodeJSON struct {
	Addr        string    `json:"addr"`
	State       State     `json:"state"`
	LastUpdated time.Time `json:"last_updated"`
	Payload     []byte    `json:"payload"`
}

// MarshalJSON implements the json.Marshaler interface.
func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&nodeJSON{
		Addr:        n.Addr.String(),
		State:       n.State,
		LastUpdated: n.LastUpdated,
		Payload:     n.Payload,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *Node) UnmarshalJSON(data []byte) error {
	var obj nodeJSON
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", obj.Addr)
	if err != nil {
		return err
	}

	n.Addr = addr
	n.State = obj.State
	n.LastUpdated = obj.LastUpdated
	n.Payload = obj.Payload

	return nil
}
