package services

import (
	"crypto/sha256"
)

// DeriveEncryptionKey derives a consistent 32-byte encryption key from a salt string
// This ensures the same salt always produces the same key
func DeriveEncryptionKey(salt string) []byte {
	// Use SHA256 to derive a consistent 32-byte key from any length salt
	hash := sha256.Sum256([]byte(salt))
	return hash[:]
}
