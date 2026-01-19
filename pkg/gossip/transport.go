package gossip

// Transport is an interface for sending and receiving messages.
type Transport interface {
	// Write sends a message to the given address.
	Write([]byte, string) error
	// Read returns a channel that can be used to receive messages.
	Read() <-chan []byte
}
