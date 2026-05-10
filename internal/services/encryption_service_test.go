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

func TestEncryptionServiceFallbackKey(t *testing.T) {
	legacyKey, err := GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate legacy key: %v", err)
	}
	primaryKey, err := GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate primary key: %v", err)
	}

	legacyService, err := NewEncryptionService(legacyKey)
	if err != nil {
		t.Fatalf("failed to create legacy service: %v", err)
	}

	encrypted, err := legacyService.Encrypt("legacy-token")
	if err != nil {
		t.Fatalf("failed to encrypt with legacy key: %v", err)
	}

	service, err := NewEncryptionServiceWithFallback(primaryKey, legacyKey)
	if err != nil {
		t.Fatalf("failed to create fallback service: %v", err)
	}

	decrypted, err := service.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("failed to decrypt with fallback key: %v", err)
	}
	if decrypted != "legacy-token" {
		t.Fatalf("expected legacy-token, got %q", decrypted)
	}

	newEncrypted, err := service.Encrypt("new-token")
	if err != nil {
		t.Fatalf("failed to encrypt with primary key: %v", err)
	}
	if _, err := legacyService.Decrypt(newEncrypted); err == nil {
		t.Fatal("legacy key should not decrypt newly encrypted content")
	}
}
