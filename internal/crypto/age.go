package crypto

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
)

// AgeEncryption implements Encryption using age
type AgeEncryption struct {
	identity  *age.X25519Identity
	recipient *age.X25519Recipient
}

// NewAgeEncryption creates a new AgeEncryption instance
func NewAgeEncryption(privateKey string) (*AgeEncryption, error) {
	identity, err := age.ParseX25519Identity(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	recipient := identity.Recipient()

	return &AgeEncryption{
		identity:  identity,
		recipient: recipient,
	}, nil
}

// NewAgeEncryptionWithPublicKey creates encryption instance with only public key
// (for encrypt-only operations)
func NewAgeEncryptionWithPublicKey(publicKey string) (*AgeEncryption, error) {
	recipient, err := age.ParseX25519Recipient(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &AgeEncryption{
		recipient: recipient,
	}, nil
}

// GenerateKey generates a new age key pair
func GenerateKey() (*KeyPair, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generate identity: %w", err)
	}

	return &KeyPair{
		PrivateKey: identity.String(),
		PublicKey:  identity.Recipient().String(),
	}, nil
}

// Encrypt encrypts plaintext
func (a *AgeEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	if a.recipient == nil {
		return nil, fmt.Errorf("no recipient configured")
	}

	out := &bytes.Buffer{}
	w, err := age.Encrypt(out, a.recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypter: %w", err)
	}

	if _, err := w.Write(plaintext); err != nil {
		return nil, fmt.Errorf("failed to write plaintext: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encrypter: %w", err)
	}

	return out.Bytes(), nil
}

// Decrypt decrypts ciphertext
func (a *AgeEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	if a.identity == nil {
		return nil, fmt.Errorf("no identity configured")
	}

	r, err := age.Decrypt(bytes.NewReader(ciphertext), a.identity)
	if err != nil {
		return nil, fmt.Errorf("failed to create decrypter: %w", err)
	}

	plaintext, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read plaintext: %w", err)
	}

	return plaintext, nil
}

// EncryptFile encrypts a file
func (a *AgeEncryption) EncryptFile(src, dst string) error {
	// Read source file
	plaintext, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Encrypt
	ciphertext, err := a.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}

	// Write destination file
	if err := os.WriteFile(dst, ciphertext, 0600); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// DecryptFile decrypts a file
func (a *AgeEncryption) DecryptFile(src, dst string) error {
	// Read source file
	ciphertext, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Decrypt
	plaintext, err := a.Decrypt(ciphertext)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	// Write destination file
	if err := os.WriteFile(dst, plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// EncryptReader encrypts from reader to writer
func (a *AgeEncryption) EncryptReader(plaintext io.Reader, ciphertext io.Writer) error {
	if a.recipient == nil {
		return fmt.Errorf("no recipient configured")
	}

	w, err := age.Encrypt(ciphertext, a.recipient)
	if err != nil {
		return fmt.Errorf("failed to create encrypter: %w", err)
	}

	if _, err := io.Copy(w, plaintext); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close encrypter: %w", err)
	}

	return nil
}

// DecryptReader decrypts from reader to writer
func (a *AgeEncryption) DecryptReader(ciphertext io.Reader, plaintext io.Writer) error {
	if a.identity == nil {
		return fmt.Errorf("no identity configured")
	}

	r, err := age.Decrypt(ciphertext, a.identity)
	if err != nil {
		return fmt.Errorf("failed to create decrypter: %w", err)
	}

	if _, err := io.Copy(plaintext, r); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	return nil
}

// SaveKeyToFile saves a private key to a file with secure permissions
func SaveKeyToFile(privateKey, path string) error {
	if err := os.WriteFile(path, []byte(privateKey), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}
	return nil
}

// LoadKeyFromFile loads a private key from a file
func LoadKeyFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}
	return string(data), nil
}

// GetPublicKey extracts the public key from a private key
func GetPublicKey(privateKey string) (string, error) {
	identity, err := age.ParseX25519Identity(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}
	return identity.Recipient().String(), nil
}
