package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GareArc/opencode-sync/internal/crypto"
)

func main() {
	testDir := "/tmp/test-opencode-sync"

	// Test 1: Generate keypair
	fmt.Println("=== Test 1: Generate Keypair ===")
	keyPair, err := crypto.GenerateKey()
	if err != nil {
		fmt.Printf("❌ Failed to generate key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Generated keypair\n")
	fmt.Printf("  Public key: %s\n", keyPair.PublicKey)

	// Test 2: Save key to file
	fmt.Println("\n=== Test 2: Save Key ===")
	keyFile := filepath.Join(testDir, "test.key")
	if err := crypto.SaveKeyToFile(keyPair.PrivateKey, keyFile); err != nil {
		fmt.Printf("❌ Failed to save key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Saved key to: %s\n", keyFile)

	// Test 3: Load key from file
	fmt.Println("\n=== Test 3: Load Key ===")
	loadedKey, err := crypto.LoadKeyFromFile(keyFile)
	if err != nil {
		fmt.Printf("❌ Failed to load key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Loaded key from file\n")

	// Test 4: Initialize encryption
	fmt.Println("\n=== Test 4: Initialize Encryption ===")
	enc, err := crypto.NewAgeEncryption(loadedKey)
	if err != nil {
		fmt.Printf("❌ Failed to initialize encryption: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Initialized encryption\n")

	// Test 5: Encrypt auth.json
	fmt.Println("\n=== Test 5: Encrypt File ===")
	authSrc := filepath.Join(testDir, "opencode-data", "auth.json")
	authEnc := filepath.Join(testDir, "auth.json.age")

	if err := enc.EncryptFile(authSrc, authEnc); err != nil {
		fmt.Printf("❌ Failed to encrypt: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Encrypted %s -> %s\n", authSrc, authEnc)

	// Verify encrypted file is not plaintext
	encData, _ := os.ReadFile(authEnc)
	srcData, _ := os.ReadFile(authSrc)
	if string(encData) == string(srcData) {
		fmt.Printf("❌ File is not encrypted! Content is plaintext\n")
		os.Exit(1)
	}
	fmt.Printf("✓ Verified file is encrypted (not plaintext)\n")

	// Test 6: Decrypt auth.json
	fmt.Println("\n=== Test 6: Decrypt File ===")
	authDec := filepath.Join(testDir, "auth-decrypted.json")

	if err := enc.DecryptFile(authEnc, authDec); err != nil {
		fmt.Printf("❌ Failed to decrypt: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Decrypted %s -> %s\n", authEnc, authDec)

	// Test 7: Verify decrypted content matches original
	fmt.Println("\n=== Test 7: Verify Content ===")
	decData, _ := os.ReadFile(authDec)

	if string(srcData) != string(decData) {
		fmt.Printf("❌ Decrypted content doesn't match original!\n")
		fmt.Printf("Original: %s\n", string(srcData))
		fmt.Printf("Decrypted: %s\n", string(decData))
		os.Exit(1)
	}
	fmt.Printf("✓ Decrypted content matches original\n")

	fmt.Println("\n=== All Tests Passed! ===")
}
