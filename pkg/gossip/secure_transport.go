package gossip

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// SecureTransport is a transport that encrypts and decrypts messages.
type SecureTransport struct {
	transport Transport
	aead      cipher.AEAD
}

// NewSecureTransport creates a new secure transport.
func NewSecureTransport(transport Transport, key []byte) (*SecureTransport, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &SecureTransport{
		transport: transport,
		aead:      aead,
	},
	nil
}

// Write encrypts and sends a message.
func (t *SecureTransport) Write(data []byte, addr string) error {
	nonce := make([]byte, t.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := t.aead.Seal(nonce, nonce, data, nil)
	return t.transport.Write(ciphertext, addr)
}

// Read receives and decrypts a message.
func (t *SecureTransport) Read() <-chan []byte {
	out := make(chan []byte)
	go func() {
		for data := range t.transport.Read() {
			nonceSize := t.aead.NonceSize()
			if len(data) < nonceSize {
				fmt.Printf("ciphertext too short: %d\n", len(data))
				continue
			}

		nonce, ciphertext := data[:nonceSize], data[nonceSize:]
		plaintext, err := t.aead.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				fmt.Printf("failed to decrypt message: %v\n", err)
				continue
			}
			out <- plaintext
		}
		close(out)
	}()
	return out
}

// Stop stops the underlying transport.
func (t *SecureTransport) Stop() {
	t.transport.Stop()
}
