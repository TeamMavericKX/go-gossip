package gossip

import (
	"encoding/json"
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

// Node represents a single member in the cluster.
type Node struct {
	// Addr is the network address of the node.
	Addr net.Addr
	// State is the current state of the node.
	State State
	// LastUpdated is the time when the node's state was last updated.
	LastUpdated time.Time
}

type nodeJSON struct {
	Addr        string    `json:"addr"`
	State       State     `json:"state"`
	LastUpdated time.Time `json:"last_updated"`
}

// MarshalJSON implements the json.Marshaler interface.
func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&nodeJSON{
		Addr:        n.Addr.String(),
		State:       n.State,
		LastUpdated: n.LastUpdated,
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

	return nil
}
