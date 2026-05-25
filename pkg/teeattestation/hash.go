// Package teeattestation provides platform-agnostic primitives for TEE
// attestation validation. Platform-specific validators (e.g. AWS Nitro)
// live in subpackages.
package teeattestation

import "crypto/sha256"

// DomainSeparator is prepended to attestation payloads before hashing.
const DomainSeparator = "CONFIDENTIAL_COMPUTE_PAYLOAD"

// DomainHash computes SHA-256 over DomainSeparator + "\n" + tag + "\n" + data.
// This is the standard domain-separated hash used for attestation UserData
// throughout the system.
func DomainHash(tag string, data []byte) []byte {
	h := sha256.New()
	h.Write([]byte(DomainSeparator))
	h.Write([]byte("\n" + tag + "\n"))
	h.Write(data)
	return h.Sum(nil)
}
