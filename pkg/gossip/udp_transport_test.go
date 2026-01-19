package gossip

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestUDPTransport_WriteRead(t *testing.T) {
	// Node 1 transport
	tr1, err := NewUDPTransport("127.0.0.1:9001")
	if err != nil {
		t.Fatalf("failed to create transport 1: %v", err)
	}
	defer tr1.Stop()

	// Node 2 transport
	tr2, err := NewUDPTransport("127.0.0.1:9002")
	if err != nil {
		t.Fatalf("failed to create transport 2: %v", err)
	}
	defer tr2.Stop()

	// Message to send
	testMsg := []byte("hello from node1")

	// Send from tr1 to tr2
	err = tr1.Write(testMsg, "127.0.0.1:9002")
	if err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Read from tr2
	select {
	case receivedMsg := <-tr2.Read():
		if string(receivedMsg) != string(testMsg) {
			t.Errorf("Expected message %s, got %s", string(testMsg), string(receivedMsg))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	// Send from tr2 to tr1
	testMsg2 := []byte("hello from node2")
	err = tr2.Write(testMsg2, "127.0.0.1:9001")
	if err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Read from tr1
	select {
	case receivedMsg := <-tr1.Read():
		if string(receivedMsg) != string(testMsg2) {
			t.Errorf("Expected message %s, got %s", string(testMsg2), string(receivedMsg))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestUDPTransport_Stop(t *testing.T) {
	tr, err := NewUDPTransport("127.0.0.1:9003")
	if err != nil {
		t.Fatalf("failed to create transport: %v", err)
	}

	tr.Stop()

	// Try writing after stop (should fail)
	err = tr.Write([]byte("data"), "127.0.0.1:9004")
	if err == nil {
		t.Error("Expected error when writing after stop, got nil")
	}

	// Ensure read channel is closed
	select {
	case _, ok := <-tr.Read():
		if ok {
			t.Error("Read channel is still open after stop")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for read channel to close")
	}
}

func TestUDPTransport_ConcurrentRead(t *testing.T) {
	tr, err := NewUDPTransport("127.0.0.1:9005")
	if err != nil {
		t.Fatalf("failed to create transport: %v", err)
	}
	defer tr.Stop()

	numMsgs := 100
	expectedMsgs := make(map[string]bool)

	// Send messages concurrently
	for i := 0; i < numMsgs; i++ {
		msg := []byte(fmt.Sprintf("message-%d", i))
		expectedMsgs[string(msg)] = false
		err := tr.Write(msg, "127.0.0.1:9005") // Send to self
		if err != nil {
			t.Errorf("failed to write message %d: %v", i, err)
		}
	}

	// Read messages
	receivedCount := 0
	for {
		select {
		case receivedMsg := <-tr.Read():
			receivedCount++
			msgStr := string(receivedMsg)
			if _, ok := expectedMsgs[msgStr]; !ok {
				t.Errorf("Received unexpected message: %s", msgStr)
			}
			expectedMsgs[msgStr] = true
			if receivedCount == numMsgs {
				goto endRead
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Timeout waiting for all messages. Received %d of %d", receivedCount, numMsgs)
		}
	}

endRead:
	for msg, received := range expectedMsgs {
		if !received {
			t.Errorf("Message %s was not received", msg)
		}
	}
}
