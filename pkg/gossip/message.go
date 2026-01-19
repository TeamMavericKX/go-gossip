package gossip

import "encoding/json"

// MessageType is the type of a message.
type MessageType int

const (
	// Ping is a message sent to a node to check if it is alive.
	Ping MessageType = iota
)

// Message is the message that is sent between nodes.
type Message struct {
	Type    MessageType `json:"type"`
	Payload []byte      `json:"payload"`
}

// Encode encodes a message to JSON.
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode decodes a message from JSON.
func Decode(data []byte) (*Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	return &m, err
}
