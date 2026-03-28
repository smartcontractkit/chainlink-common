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

	"github.com/hf/nitrite"
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

// ValidateAttestation verifies an AWS Nitro attestation document against
// expected user data and trusted PCR measurements. Always validates against
// the AWS Nitro Enclaves root certificate.
//
// For testing with fake enclaves, use ValidateAttestationWithRoots or inject
// a custom validator function.
func ValidateAttestation(attestation, expectedUserData, trustedMeasurements []byte) error {
	return ValidateAttestationWithRoots(attestation, expectedUserData, trustedMeasurements, DefaultCARoots)
}

// ValidateAttestationWithRoots verifies an AWS Nitro attestation document
// using a custom CA root certificate. This is primarily for testing with
// fake enclaves that use self-signed CA roots.
func ValidateAttestationWithRoots(attestation, expectedUserData, trustedMeasurements []byte, caRootsPEM string) error {
	if attestation == nil {
		return errors.New("attestation is nil")
	}

	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM([]byte(caRootsPEM))
	if !ok {
		return errors.New("failed to parse CA roots")
	}
	result, err := nitrite.Verify(attestation, nitrite.VerifyOptions{
		CurrentTime: time.Now(),
		Roots:       pool,
	})
	if err != nil {
		return fmt.Errorf("failed to verify nitro attestation: %w", err)
	}
	if !result.SignatureOK {
		return errors.New("signature verification failed")
	}

	if !bytes.Equal(expectedUserData, result.Document.UserData) {
		return fmt.Errorf("expected user data %x, got %x", expectedUserData, result.Document.UserData)
	}

	var trustedPCRs PCRs
	if err := json.Unmarshal(trustedMeasurements, &trustedPCRs); err != nil {
		return fmt.Errorf("failed to unmarshal trusted PCRs: %w", err)
	}
	if len(result.Document.PCRs) < 3 {
		return fmt.Errorf("attestation document has %d PCRs, need at least 3", len(result.Document.PCRs))
	}
	if !bytes.Equal(result.Document.PCRs[0], trustedPCRs.PCR0) {
		return fmt.Errorf("PCR0 mismatch: expected %x", trustedPCRs.PCR0)
	}
	if !bytes.Equal(result.Document.PCRs[1], trustedPCRs.PCR1) {
		return fmt.Errorf("PCR1 mismatch: expected %x", trustedPCRs.PCR1)
	}
	if !bytes.Equal(result.Document.PCRs[2], trustedPCRs.PCR2) {
		return fmt.Errorf("PCR2 mismatch: expected %x", trustedPCRs.PCR2)
	}
	return nil
}
