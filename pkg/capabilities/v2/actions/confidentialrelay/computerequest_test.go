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
		ApplicationRequestID:      "application-request-id",
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

func TestUsesLegacyComputeRequestHash(t *testing.T) {
	for _, v := range []string{computeRequestLegacyVersion, "0.0.1", "0.0.5"} {
		require.True(t, usesLegacyComputeRequestHash(v), "version %q should use the legacy hashing scheme", v)
	}

	for _, v := range []string{"0.0.7", "0.1.0", "1.2.3"} {
		require.False(t, usesLegacyComputeRequestHash(v), "version %q should use the current hashing scheme", v)
	}

	for _, v := range []string{"", "not-a-version", "1.x"} {
		require.False(t, usesLegacyComputeRequestHash(v), "unparseable version %q degrades to the current scheme", v)
	}
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

// Version is hashed for the legacy scheme, matching confidential-compute (which is
// migrating Version out of the hash). Non-legacy versions are excluded, so different
// non-legacy versions hash identically, while legacy-scheme versions are bound.
func TestComputeRequestHash_VersionOnlyHashedForLegacy(t *testing.T) {
	nonLegacyA := sampleComputeRequest()
	nonLegacyA.Version = "0.0.7"
	nonLegacyB := sampleComputeRequest()
	nonLegacyB.Version = "1.2.3"
	require.Equal(t, nonLegacyA.Hash(), nonLegacyB.Hash(), "non-legacy Version must not affect the hash")

	legacy := sampleComputeRequest()
	legacy.Version = computeRequestLegacyVersion
	require.NotEqual(t, legacy.Hash(), nonLegacyA.Hash(), "legacy Version must be bound into the hash")

	olderLegacyA := sampleComputeRequest()
	olderLegacyA.Version = "0.0.5"
	olderLegacyB := sampleComputeRequest()
	olderLegacyB.Version = "0.0.4"
	require.NotEqual(t, olderLegacyA.Hash(), olderLegacyB.Hash(), "versions at or below legacy must be bound into the hash")
}

// ApplicationRequestID is the post-legacy replacement for binding application-level
// request identity without constraining RequestID's 32-byte protocol shape.
func TestComputeRequestHash_ApplicationRequestIDOnlyHashedForNonLegacy(t *testing.T) {
	legacyA := sampleComputeRequest()
	legacyA.ApplicationRequestID = "exec-a"
	legacyB := sampleComputeRequest()
	legacyB.ApplicationRequestID = "exec-b"
	require.Equal(t, legacyA.Hash(), legacyB.Hash(), "legacy ApplicationRequestID must not affect the hash")

	olderLegacyA := sampleComputeRequest()
	olderLegacyA.Version = "0.0.5"
	olderLegacyA.ApplicationRequestID = "exec-a"
	olderLegacyB := sampleComputeRequest()
	olderLegacyB.Version = "0.0.5"
	olderLegacyB.ApplicationRequestID = "exec-b"
	require.Equal(t, olderLegacyA.Hash(), olderLegacyB.Hash(), "legacy-scheme ApplicationRequestID must not affect the hash")

	nonLegacyA := sampleComputeRequest()
	nonLegacyA.Version = "0.0.7"
	nonLegacyA.ApplicationRequestID = "exec-a"
	nonLegacyB := sampleComputeRequest()
	nonLegacyB.Version = "0.0.7"
	nonLegacyB.ApplicationRequestID = "exec-b"
	require.NotEqual(t, nonLegacyA.Hash(), nonLegacyB.Hash(), "non-legacy ApplicationRequestID must be bound into the hash")
}
