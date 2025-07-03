package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// CalculateRequestDigest creates a SHA256 digest of the request for integrity verification
// This function is shared between client (JWT generation) and server (JWT validation)
func CalculateRequestDigest(req any) string {
	// Create canonical string representation
	var canonical string
	if s, ok := req.(fmt.Stringer); ok {
		canonical = s.String()
	} else {
		canonical = fmt.Sprintf("%v", req)
	}

	// Hash and encode as hex
	hash := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(hash[:])
}
