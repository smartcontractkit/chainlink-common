// Package teeattestation provides platform-agnostic primitives for TEE
// attestation validation. Platform-specific validators (e.g. AWS Nitro)
// live in subpackages.
package teeattestation

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
)

// DomainSeparator is prepended to attestation payloads before hashing.
const DomainSeparator = "CONFIDENTIAL_COMPUTE_PAYLOAD"

// DomainHash computes SHA-256 over the DomainSeparator and the length-prefixed
// tag and data, so distinct (tag, data) pairs can never share a pre-image. [CL112-14]
func DomainHash(tag string, data []byte) ([]byte, error) {
	if tag == "" || !isAlphanumeric(tag) {
		return nil, fmt.Errorf("invalid tag: must be non-empty and contain only alphanumeric characters")
	}
	h := sha256.New()
	h.Write([]byte(DomainSeparator))
	writeWithLength(h, []byte(tag))
	writeWithLength(h, data)
	return h.Sum(nil), nil
}

// isAlphanumeric reports whether s consists only of ASCII letters and digits.
func isAlphanumeric(s string) bool {
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

// writeWithLength writes an 8-byte big-endian length prefix followed by data.
func writeWithLength(h hash.Hash, data []byte) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(len(data)))
	h.Write(buf[:])
	h.Write(data)
}
