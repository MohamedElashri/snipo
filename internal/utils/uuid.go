package utils

import (
	"crypto/rand"
	"fmt"
)

// GenerateUUID generates a v4 UUID.
func GenerateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	// Set version to 4
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant to 10
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
