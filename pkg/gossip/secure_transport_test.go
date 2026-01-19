package gossip

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

// MockTransport implements the Transport interface for testing purposes.
type MockTransport struct {
	writeCh chan []byte
	readCh  chan []byte
	stopCh  chan struct{}
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		writeCh: make(chan []byte, 100),
		readCh:  make(chan []byte, 100),
		stopCh:  make(chan struct{}),
	}
}

func (m *MockTransport) Write(data []byte, addr string) error {
	select {
	case m.writeCh <- data:
		return nil
	case <-m.stopCh:
		return fmt.Errorf("mock transport stopped")
	}
}

func (m *MockTransport) Read() <-chan []byte {
	return m.readCh
}

func (m *MockTransport) Stop() {
	close(m.stopCh)
}

// Simulate message flow between two mock transports
func (m *MockTransport) Connect(other *MockTransport) {
	go func() {
		for {
			select {
			case msg := <-m.writeCh:
				other.readCh <- msg
			case <-m.stopCh:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case msg := <-other.writeCh:
				m.readCh <- msg
			case <-other.stopCh:
				return
			}
		}
	}()
}

func TestSecureTransport_WriteRead(t *testing.T) {
	// Generate a random AES key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	mockTr1 := NewMockTransport()
	defer mockTr1.Stop()
	mockTr2 := NewMockTransport()
	defer mockTr2.Stop()

	mockTr1.Connect(mockTr2)

	secureTr1, err := NewSecureTransport(mockTr1, key)
	if err != nil {
		t.Fatalf("failed to create secure transport 1: %v", err)
	}
	defer secureTr1.Stop()

	secureTr2, err := NewSecureTransport(mockTr2, key)
	if err != nil {
		t.Fatalf("failed to create secure transport 2: %v", err)
	}
	defer secureTr2.Stop()

	testMsg := []byte("secret message from sender")
	targetAddr := "127.0.0.1:8080" // Address doesn't matter for MockTransport

	err = secureTr1.Write(testMsg, targetAddr)
	if err != nil {
		t.Fatalf("secureTr1 failed to write: %v", err)
	}

	select {
	case receivedMsg := <-secureTr2.Read():
		if !bytes.Equal(receivedMsg, testMsg) {
			t.Errorf("Expected %s, got %s", hex.EncodeToString(testMsg), hex.EncodeToString(receivedMsg))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for message in secureTr2")
	}

	// Test reverse direction
	testMsg2 := []byte("another secret message")
	err = secureTr2.Write(testMsg2, targetAddr)
	if err != nil {
		t.Fatalf("secureTr2 failed to write: %v", err)
	}

	select {
	case receivedMsg := <-secureTr1.Read():
		if !bytes.Equal(receivedMsg, testMsg2) {
			t.Errorf("Expected %s, got %s", hex.EncodeToString(testMsg2), hex.EncodeToString(receivedMsg))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for message in secureTr1")
	}
}

func TestSecureTransport_MismatchedKeys(t *testing.T) {
	key1 := make([]byte, 32)
	rand.Read(key1)
	key2 := make([]byte, 32)
	rand.Read(key2) // Different key

	mockTr1 := NewMockTransport()
	defer mockTr1.Stop()
	mockTr2 := NewMockTransport()
	defer mockTr2.Stop()

	mockTr1.Connect(mockTr2)

	secureTr1, err := NewSecureTransport(mockTr1, key1)
	if err != nil {
		t.Fatalf("failed to create secure transport 1: %v", err)
	}
	defer secureTr1.Stop()

	secureTr2, err := NewSecureTransport(mockTr2, key2) // Use different key
	if err != nil {
		t.Fatalf("failed to create secure transport 2: %v", err)
	}
	defer secureTr2.Stop()

	testMsg := []byte("this message should not decrypt")
	targetAddr := "127.0.0.1:8080"

	err = secureTr1.Write(testMsg, targetAddr)
	if err != nil {
		t.Fatalf("secureTr1 failed to write: %v", err)
	}

	select {
	case <-secureTr2.Read():
		t.Error("Received a message with mismatched keys, expected decryption failure")
	case <-time.After(500 * time.Millisecond):
		// Expected timeout or decryption error logged, not a successful read
	}
}
