package confidentialrelay

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func sampleComputeRequest() ComputeRequest {
	var rid [32]byte
	for i := range rid {
		rid[i] = byte(i)
	}
	return ComputeRequest{
		RequestID:                 rid,
		PublicData:                []byte("public-data"),
		Ciphertexts:               [][]byte{[]byte("ct-a"), []byte("ct-b")},
		CiphertextNames:           []string{"name-a", "name-b"},
		EnclaveEphemeralPublicKey: []byte("ephemeral-pub-key"),
		MasterPublicKey:           []byte("master-pub-key"),
		AppID:                     "test-app",
		Version:                   computeRequestLegacyVersion,
	}
}

func TestComputeRequestHash_Deterministic(t *testing.T) {
	require.Equal(t, sampleComputeRequest().Hash(), sampleComputeRequest().Hash())
}

// Every field the source binds must change the hash. (Conformance with
// confidential-compute's source Hash is enforced by a test in that repo, which can
// import this package; chainlink-common cannot import confidential-compute.)
func TestComputeRequestHash_BindsFields(t *testing.T) {
	base := sampleComputeRequest().Hash()

	mutations := map[string]func(*ComputeRequest){
		"requestID":       func(c *ComputeRequest) { c.RequestID = [32]byte{0xff} },
		"publicData":      func(c *ComputeRequest) { c.PublicData = []byte("other") },
		"ciphertextNames": func(c *ComputeRequest) { c.CiphertextNames = []string{"x"} },
		"ciphertexts":     func(c *ComputeRequest) { c.Ciphertexts = [][]byte{[]byte("x")} },
		"ephemeralKey":    func(c *ComputeRequest) { c.EnclaveEphemeralPublicKey = []byte("x") },
		"masterKey":       func(c *ComputeRequest) { c.MasterPublicKey = []byte("x") },
		"appID":           func(c *ComputeRequest) { c.AppID = "other" },
		"version":         func(c *ComputeRequest) { c.Version = "other" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			c := sampleComputeRequest()
			mutate(&c)
			require.NotEqual(t, base, c.Hash(), "hash must change when %s changes", name)
		})
	}
}

// EncryptedDecryptionKeyShares is intentionally excluded from the hash, matching the
// source; this pins that so a future copy edit can't silently start binding it.
func TestComputeRequestHash_IgnoresEncryptedShares(t *testing.T) {
	withShares := sampleComputeRequest()
	withShares.EncryptedDecryptionKeyShares = [][][]byte{{[]byte("share")}}
	require.Equal(t, sampleComputeRequest().Hash(), withShares.Hash())
}

// Version is hashed only for the legacy version, matching confidential-compute (which is
// migrating Version out of the hash). Non-legacy versions are excluded, so different
// non-legacy versions hash identically, while the legacy version is bound.
func TestComputeRequestHash_VersionOnlyHashedForLegacy(t *testing.T) {
	nonLegacyA := sampleComputeRequest()
	nonLegacyA.Version = "0.0.7"
	nonLegacyB := sampleComputeRequest()
	nonLegacyB.Version = "1.2.3"
	require.Equal(t, nonLegacyA.Hash(), nonLegacyB.Hash(), "non-legacy Version must not affect the hash")

	legacy := sampleComputeRequest()
	legacy.Version = computeRequestLegacyVersion
	require.NotEqual(t, legacy.Hash(), nonLegacyA.Hash(), "legacy Version must be bound into the hash")
}
