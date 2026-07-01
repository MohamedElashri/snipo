package services

import (
	"crypto/hmac"
	"crypto/sha256"
)

// DeriveEncryptionKeyWithSecret binds the persisted encryption salt to the
// session secret so a copied database plus salt file is not enough to decrypt
// stored third-party tokens.
func DeriveEncryptionKeyWithSecret(salt, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("snipo-encryption-key-v2"))
	mac.Write([]byte{0})
	mac.Write([]byte(salt))
	return mac.Sum(nil)
}
