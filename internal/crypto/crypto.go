package crypto

import (
	"io"
)

// Encryption interface defines methods for encrypting and decrypting data
type Encryption interface {
	// Encrypt encrypts plaintext data
	Encrypt(plaintext []byte) ([]byte, error)

	// Decrypt decrypts ciphertext data
	Decrypt(ciphertext []byte) ([]byte, error)

	// EncryptFile encrypts a file and writes to destination
	EncryptFile(src, dst string) error

	// DecryptFile decrypts a file and writes to destination
	DecryptFile(src, dst string) error

	// EncryptReader encrypts data from reader and writes to writer
	EncryptReader(plaintext io.Reader, ciphertext io.Writer) error

	// DecryptReader decrypts data from reader and writes to writer
	DecryptReader(ciphertext io.Reader, plaintext io.Writer) error
}

// KeyPair represents a public/private key pair
type KeyPair struct {
	PublicKey  string
	PrivateKey string
}

// KeyGenerator generates encryption keys
type KeyGenerator interface {
	// Generate generates a new key pair
	Generate() (*KeyPair, error)

	// ParsePublicKey parses a public key from string
	ParsePublicKey(key string) error

	// ParsePrivateKey parses a private key from string
	ParsePrivateKey(key string) error
}
