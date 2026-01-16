package crypto

import (
	"io"
	"os"
)

// NoOpEncryption implements Encryption with no actual encryption
// Useful for testing and when encryption is disabled
type NoOpEncryption struct{}

// NewNoOpEncryption creates a new NoOpEncryption instance
func NewNoOpEncryption() *NoOpEncryption {
	return &NoOpEncryption{}
}

// Encrypt returns plaintext unchanged
func (n *NoOpEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	return plaintext, nil
}

// Decrypt returns ciphertext unchanged
func (n *NoOpEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

// EncryptFile copies file without encryption
func (n *NoOpEncryption) EncryptFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}

// DecryptFile copies file without decryption
func (n *NoOpEncryption) DecryptFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}

// EncryptReader copies data without encryption
func (n *NoOpEncryption) EncryptReader(plaintext io.Reader, ciphertext io.Writer) error {
	_, err := io.Copy(ciphertext, plaintext)
	return err
}

// DecryptReader copies data without decryption
func (n *NoOpEncryption) DecryptReader(ciphertext io.Reader, plaintext io.Writer) error {
	_, err := io.Copy(plaintext, ciphertext)
	return err
}
