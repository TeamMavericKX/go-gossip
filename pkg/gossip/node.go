package gossip

import (
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
