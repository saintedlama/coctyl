package coctyl

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashBytes computes the SHA-256 dactyl string for a byte slice.
func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// HashString computes the SHA-256 dactyl string for a string.
func HashString(data string) string {
	return HashBytes([]byte(data))
}
