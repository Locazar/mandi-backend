package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashRefreshToken returns the hex-encoded SHA-256 of a high-entropy refresh
// token for storage. Compare hashes, never plaintext.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
