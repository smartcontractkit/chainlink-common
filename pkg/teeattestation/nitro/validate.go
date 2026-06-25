// Package nitro provides AWS Nitro Enclave attestation validation.
package nitro

import (
	"bytes"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// HexBytes is a custom type that unmarshals hex strings into a byte slice
// and marshals byte slices back to hex strings. This allows parsing AWS Nitro
// measurements, which use hex byte strings in JSON.
type HexBytes []byte

func (h *HexBytes) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("HexBytes: cannot unmarshal JSON into string: %w", err)
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("HexBytes: failed to decode hex string '%s': %w", s, err)
	}
	*h = decoded
	return nil
}

func (h HexBytes) MarshalJSON() ([]byte, error) {
	s := hex.EncodeToString(h)
	return json.Marshal(s)
}

// PCRs holds Platform Configuration Register values for attestation validation.
type PCRs struct {
	PCR0 HexBytes `json:"pcr0"`
	PCR1 HexBytes `json:"pcr1"`
	PCR2 HexBytes `json:"pcr2"`
}

// DefaultCARoots is the AWS Nitro Enclaves root certificate.
// Downloaded from: https://aws-nitro-enclaves.amazonaws.com/AWS_NitroEnclaves_Root-G1.zip
const DefaultCARoots = "-----BEGIN CERTIFICATE-----\nMIICETCCAZagAwIBAgIRAPkxdWgbkK/hHUbMtOTn+FYwCgYIKoZIzj0EAwMwSTEL\nMAkGA1UEBhMCVVMxDzANBgNVBAoMBkFtYXpvbjEMMAoGA1UECwwDQVdTMRswGQYD\nVQQDDBJhd3Mubml0cm8tZW5jbGF2ZXMwHhcNMTkxMDI4MTMyODA1WhcNNDkxMDI4\nMTQyODA1WjBJMQswCQYDVQQGEwJVUzEPMA0GA1UECgwGQW1hem9uMQwwCgYDVQQL\nDANBV1MxGzAZBgNVBAMMEmF3cy5uaXRyby1lbmNsYXZlczB2MBAGByqGSM49AgEG\nBSuBBAAiA2IABPwCVOumCMHzaHDimtqQvkY4MpJzbolL//Zy2YlES1BR5TSksfbb\n48C8WBoyt7F2Bw7eEtaaP+ohG2bnUs990d0JX28TcPQXCEPZ3BABIeTPYwEoCWZE\nh8l5YoQwTcU/9KNCMEAwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUkCW1DdkF\nR+eWw5b6cp3PmanfS5YwDgYDVR0PAQH/BAQDAgGGMAoGCCqGSM49BAMDA2kAMGYC\nMQCjfy+Rocm9Xue4YnwWmNJVA44fA0P5W2OpYow9OYCVRaEevL8uO1XYru5xtMPW\nrfMCMQCi85sWBbJwKKXdS6BptQFuZbT73o/gBh1qUxl/nNr12UO8Yfwr6wPLb+6N\nIwLz3/Y=\n-----END CERTIFICATE-----\n"

// Document holds the validated, parsed fields of a Nitro attestation that a
// caller needs in order to bind an attestation to a specific enclave identity
// and to check freshness. A Document is returned only after the full
// validation chain (certificate chain, COSE signature, expected user data,
// trusted PCR0/1/2) has passed.
type Document struct {
	// PCRs holds every platform configuration register present in the
	// attestation, keyed by index. PCR0/1/2 have already been checked against
	// the trusted measurements; higher PCRs (e.g. 3/4/8) are exposed unchecked.
	PCRs map[uint][]byte
	// LeafPublicKey is the SPKI (DER) of the attestation's end-entity
	// certificate public key. AWS rotates the end-entity certificate per
	// enclave instance (roughly every few hours), so this is not a stable
	// identity anchor on its own.
	LeafPublicKey []byte
	// PublicKey is the enclave-supplied public_key field of the attestation
	// (nil if the enclave did not set one). This is the field an enclave can
	// use to publish a long-lived identity key it generated inside the TEE.
	PublicKey []byte
	// UserData is the attestation user_data field. It has already been checked
	// to equal the expectedUserData argument.
	UserData []byte
	// Nonce is the attestation nonce field (nil if the enclave did not set
	// one). ValidateAndParse does NOT check it: callers that want replay
	// protection must compare it against a freshly generated challenge.
	Nonce []byte
	// ModuleID is the attestation module_id field.
	ModuleID string
}

// VerifyPCR checks that the PCR at index equals expected. Use it for
// per-instance PCRs such as PCR4 (the SHA384 of the parent instance ID), which
// differ per enclave and therefore cannot be part of the shared trusted
// measurements checked by ValidateAndParse.
//
// It returns an error if expected is empty, the PCR is absent or
// length-mismatched, the PCR is all zero (debug-mode enclaves emit all-zero
// PCRs, which must never be accepted as a real measurement), or the values
// differ.
func (d *Document) VerifyPCR(index uint, expected []byte) error {
	if len(expected) == 0 {
		return fmt.Errorf("expected PCR%d value is empty", index)
	}
	actual, ok := d.PCRs[index]
	if !ok {
		return fmt.Errorf("attestation has no PCR%d", index)
	}
	if len(actual) != len(expected) {
		return fmt.Errorf("PCR%d length mismatch: expected %d bytes, got %d", index, len(expected), len(actual))
	}
	if allZero(actual) {
		return fmt.Errorf("PCR%d is all zero (debug-mode enclave), refusing", index)
	}
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("PCR%d mismatch", index)
	}
	return nil
}

// VerifyExpectedPCRs checks each index/value in expected against the document
// via VerifyPCR. It is a convenience for callers asserting several per-instance
// PCRs at once and returns the first failure.
func (d *Document) VerifyExpectedPCRs(expected map[uint][]byte) error {
	for index, value := range expected {
		if err := d.VerifyPCR(index, value); err != nil {
			return err
		}
	}
	return nil
}

func allZero(b []byte) bool {
	for _, x := range b {
		if x != 0 {
			return false
		}
	}
	return len(b) != 0
}

// ValidateAttestation verifies an AWS Nitro attestation document against
// expected user data and trusted PCR measurements. Always validates against
// the AWS Nitro Enclaves root certificate.
//
// For testing with fake enclaves, use ValidateAttestationWithRoots or inject
// a custom validator function.
func ValidateAttestation(attestation, expectedUserData, trustedMeasurements []byte) error {
	_, err := ValidateAndParse(attestation, expectedUserData, trustedMeasurements)
	return err
}

// ValidateAttestationWithRoots verifies an AWS Nitro attestation document
// using a custom CA root certificate. This is primarily for testing with
// fake enclaves that use self-signed CA roots.
func ValidateAttestationWithRoots(attestation, expectedUserData, trustedMeasurements []byte, caRootsPEM string) error {
	_, err := ValidateAndParseWithRoots(attestation, expectedUserData, trustedMeasurements, caRootsPEM)
	return err
}

// ValidateAndParse runs the same validation as ValidateAttestation and, on
// success, returns the parsed attestation fields. Callers that need to bind
// the attestation to a specific enclave identity (via Document.PublicKey or
// Document.LeafPublicKey) or to check freshness (via Document.Nonce) should
// use this instead of ValidateAttestation.
func ValidateAndParse(attestation, expectedUserData, trustedMeasurements []byte) (*Document, error) {
	return ValidateAndParseWithRoots(attestation, expectedUserData, trustedMeasurements, DefaultCARoots)
}

// ValidateAndParseWithRoots is ValidateAndParse against a custom CA root
// certificate. This is primarily for testing with fake enclaves that use
// self-signed CA roots.
func ValidateAndParseWithRoots(attestation, expectedUserData, trustedMeasurements []byte, caRootsPEM string) (*Document, error) {
	if attestation == nil {
		return nil, errors.New("attestation is nil")
	}

	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM([]byte(caRootsPEM))
	if !ok {
		return nil, errors.New("failed to parse CA roots")
	}
	result, err := verifyAttestationDocument(attestation, pool, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to verify nitro attestation: %w", err)
	}
	if !result.signatureOK {
		return nil, errors.New("signature verification failed")
	}

	if !bytes.Equal(expectedUserData, result.document.UserData) {
		return nil, fmt.Errorf("expected user data %x, got %x", expectedUserData, result.document.UserData)
	}

	var trustedPCRs PCRs
	if err := json.Unmarshal(trustedMeasurements, &trustedPCRs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trusted PCRs: %w", err)
	}
	if len(result.document.PCRs) < 3 {
		return nil, fmt.Errorf("attestation document has %d PCRs, need at least 3", len(result.document.PCRs))
	}
	if !bytes.Equal(result.document.PCRs[0], trustedPCRs.PCR0) {
		return nil, fmt.Errorf("PCR0 mismatch: expected %x", trustedPCRs.PCR0)
	}
	if !bytes.Equal(result.document.PCRs[1], trustedPCRs.PCR1) {
		return nil, fmt.Errorf("PCR1 mismatch: expected %x", trustedPCRs.PCR1)
	}
	if !bytes.Equal(result.document.PCRs[2], trustedPCRs.PCR2) {
		return nil, fmt.Errorf("PCR2 mismatch: expected %x", trustedPCRs.PCR2)
	}

	return &Document{
		PCRs:          result.document.PCRs,
		LeafPublicKey: result.leafPublicKey,
		PublicKey:     result.document.PublicKey,
		UserData:      result.document.UserData,
		Nonce:         result.document.Nonce,
		ModuleID:      result.document.ModuleID,
	}, nil
}
