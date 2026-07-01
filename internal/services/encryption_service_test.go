package services

import (
	"testing"
)

func TestEncryptionService(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	service, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("failed to create encryption service: %v", err)
	}

	t.Run("encrypt and decrypt", func(t *testing.T) {
		plaintext := "test-github-token-12345"

		encrypted, err := service.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("failed to encrypt: %v", err)
		}

		if encrypted == plaintext {
			t.Error("encrypted text should not equal plaintext")
		}

		decrypted, err := service.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("failed to decrypt: %v", err)
		}

		if decrypted != plaintext {
			t.Errorf("decrypted text does not match plaintext: got %s, want %s", decrypted, plaintext)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		encrypted, err := service.Encrypt("")
		if err != nil {
			t.Fatalf("failed to encrypt empty string: %v", err)
		}

		if encrypted != "" {
			t.Error("encrypted empty string should be empty")
		}

		decrypted, err := service.Decrypt("")
		if err != nil {
			t.Fatalf("failed to decrypt empty string: %v", err)
		}

		if decrypted != "" {
			t.Error("decrypted empty string should be empty")
		}
	})

	t.Run("invalid key length", func(t *testing.T) {
		invalidKey := []byte("short")
		_, err := NewEncryptionService(invalidKey)
		if err == nil {
			t.Error("expected error for invalid key length")
		}
	})
}
