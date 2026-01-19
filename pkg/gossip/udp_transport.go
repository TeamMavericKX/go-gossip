package gossip

import (
	"net"
)

// UDPTransport is a transport that uses UDP for communication.
type UDPTransport struct {
	conn    *net.UDPConn
	readCh  chan []byte
	stop    chan struct{}
}

// NewUDPTransport creates a new UDP transport.
func NewUDPTransport(addr string) (*UDPTransport, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	t := &UDPTransport{
		conn:   conn,
		readCh: make(chan []byte),
		stop:   make(chan struct{}),
	}

	go t.readLoop()

	return t, nil
}

// Write sends a message to the given address.
func (t *UDPTransport) Write(data []byte, addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	_, err = t.conn.WriteToUDP(data, udpAddr)
	return err
}

// Read returns a channel that can be used to receive messages.
func (t *UDPTransport) Read() <-chan []byte {
	return t.readCh
}

// Stop stops the transport.
func (t *UDPTransport) Stop() {
	close(t.stop)
	t.conn.Close()
}

func (t *UDPTransport) readLoop() {
	buf := make([]byte, 1024)
	for {
		select {
		case <-t.stop:
			return
		default:
			n, _, err := t.conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			t.readCh <- buf[:n]
		}
	}
}
