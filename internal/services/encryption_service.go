package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptionService handles encryption and decryption of sensitive data
type EncryptionService struct {
	key          []byte
	fallbackKeys [][]byte
}

// NewEncryptionService creates a new encryption service
// key should be 32 bytes for AES-256
func NewEncryptionService(key []byte) (*EncryptionService, error) {
	return NewEncryptionServiceWithFallback(key)
}

// NewEncryptionServiceWithFallback creates a service that encrypts with the
// primary key and can decrypt legacy ciphertext with fallback keys.
func NewEncryptionServiceWithFallback(key []byte, fallbackKeys ...[]byte) (*EncryptionService, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes for AES-256")
	}
	for _, fallbackKey := range fallbackKeys {
		if len(fallbackKey) != 32 {
			return nil, fmt.Errorf("fallback encryption key must be 32 bytes for AES-256")
		}
	}

	return &EncryptionService{
		key:          key,
		fallbackKeys: fallbackKeys,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	plaintext, err := openGCM(s.key, data)
	if err == nil {
		return string(plaintext), nil
	}

	for _, fallbackKey := range s.fallbackKeys {
		plaintext, fallbackErr := openGCM(fallbackKey, data)
		if fallbackErr == nil {
			return string(plaintext), nil
		}
	}

	return "", fmt.Errorf("failed to decrypt: %w", err)
}

func openGCM(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, encryptedData, nil)
}

// GenerateKey generates a random 32-byte key for AES-256
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}
