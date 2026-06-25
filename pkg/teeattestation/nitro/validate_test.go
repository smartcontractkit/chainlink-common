package nitro

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/teeattestation"
	nitrofake "github.com/smartcontractkit/chainlink-common/pkg/teeattestation/nitro/fake"
)

func TestValidateAttestation_Attestor(t *testing.T) {
	fa, err := nitrofake.NewAttestor()
	require.NoError(t, err)

	userData := teeattestation.DomainHash("test-tag", []byte(`{"key":"value"}`))
	doc, err := fa.CreateAttestation(userData)
	require.NoError(t, err)

	err = ValidateAttestationWithRoots(doc, userData, fa.TrustedPCRsJSON(), fa.CARootsPEM())
	require.NoError(t, err)
}

// TestVerifyAttestationDocument_MaxAge covers PRIV-438 / CL112-10: fresh
// attestations verify; stale or far-future ones are rejected even while the
// leaf cert is still valid.
func TestVerifyAttestationDocument_MaxAge(t *testing.T) {
	fa, err := nitrofake.NewAttestor()
	require.NoError(t, err)
	doc, err := fa.CreateAttestation([]byte("user-data"))
	require.NoError(t, err)
	pool := fa.CARoots()

	// Fresh: accepted.
	_, err = verifyAttestationDocument(doc, pool, time.Now())
	require.NoError(t, err)

	// Too old: rejected (cert is still valid, only age fails).
	_, err = verifyAttestationDocument(doc, pool, time.Now().Add(maxAttestationAge+time.Minute))
	require.ErrorIs(t, err, errStaleAttestation)

	// Too far in the future: also rejected.
	_, err = verifyAttestationDocument(doc, pool, time.Now().Add(-(maxAttestationAge + time.Minute)))
	require.ErrorIs(t, err, errStaleAttestation)
}

func TestValidateAttestation_WrongUserData(t *testing.T) {
	fa, err := nitrofake.NewAttestor()
	require.NoError(t, err)

	userData := teeattestation.DomainHash("test-tag", []byte(`{"key":"value"}`))
	doc, err := fa.CreateAttestation(userData)
	require.NoError(t, err)

	wrongData := teeattestation.DomainHash("wrong-tag", []byte(`{"key":"value"}`))
	err = ValidateAttestationWithRoots(doc, wrongData, fa.TrustedPCRsJSON(), fa.CARootsPEM())
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected user data")
}

func TestValidateAttestation_WrongPCRs(t *testing.T) {
	fa, err := nitrofake.NewAttestor()
	require.NoError(t, err)

	userData := []byte("test-data")
	doc, err := fa.CreateAttestation(userData)
	require.NoError(t, err)

	wrongPCRs := []byte(`{"pcr0":"aa","pcr1":"bb","pcr2":"cc"}`)
	err = ValidateAttestationWithRoots(doc, userData, wrongPCRs, fa.CARootsPEM())
	require.Error(t, err)
	require.Contains(t, err.Error(), "PCR0 mismatch")
}

func TestValidateAndParse_SurfacesIdentityFields(t *testing.T) {
	fa, err := nitrofake.NewAttestor()
	require.NoError(t, err)

	userData := teeattestation.DomainHash("test-tag", []byte(`{"key":"value"}`))
	nonce := []byte("request-id-as-nonce")
	identityKey := []byte("enclave-long-lived-identity-pubkey")

	doc, err := fa.CreateAttestation(userData,
		nitrofake.WithNonce(nonce),
		nitrofake.WithPublicKey(identityKey),
	)
	require.NoError(t, err)

	parsed, err := ValidateAndParseWithRoots(doc, userData, fa.TrustedPCRsJSON(), fa.CARootsPEM())
	require.NoError(t, err)
	require.NotNil(t, parsed)

	assert.Equal(t, userData, parsed.UserData)
	assert.Equal(t, nonce, parsed.Nonce)
	assert.Equal(t, identityKey, parsed.PublicKey)
	assert.NotEmpty(t, parsed.PCRs)

	wantLeaf, err := fa.LeafPublicKeyDER()
	require.NoError(t, err)
	assert.Equal(t, wantLeaf, parsed.LeafPublicKey)
}

// ValidateAndParse does not check the nonce itself; freshness is the caller's
// responsibility. A mismatched user data still fails, and the same validation
// failures surface as for ValidateAttestation.
func TestValidateAndParse_FailsOnWrongUserData(t *testing.T) {
	fa, err := nitrofake.NewAttestor()
	require.NoError(t, err)

	userData := teeattestation.DomainHash("test-tag", []byte(`{"key":"value"}`))
	doc, err := fa.CreateAttestation(userData, nitrofake.WithNonce([]byte("n")))
	require.NoError(t, err)

	parsed, err := ValidateAndParseWithRoots(doc, []byte("wrong"), fa.TrustedPCRsJSON(), fa.CARootsPEM())
	require.Error(t, err)
	require.Nil(t, parsed)
	require.Contains(t, err.Error(), "expected user data")
}

func TestDocument_VerifyPCR(t *testing.T) {
	pcr4 := make([]byte, 48)
	for i := range pcr4 {
		pcr4[i] = byte(i + 1)
	}
	doc := &Document{PCRs: map[uint][]byte{4: pcr4}}

	t.Run("match", func(t *testing.T) {
		require.NoError(t, doc.VerifyPCR(4, pcr4))
	})

	t.Run("mismatch", func(t *testing.T) {
		other := make([]byte, 48)
		copy(other, pcr4)
		other[0] ^= 0xff
		err := doc.VerifyPCR(4, other)
		require.Error(t, err)
		require.Contains(t, err.Error(), "PCR4 mismatch")
	})

	t.Run("absent index", func(t *testing.T) {
		err := doc.VerifyPCR(3, pcr4)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no PCR3")
	})

	t.Run("length mismatch", func(t *testing.T) {
		err := doc.VerifyPCR(4, pcr4[:32])
		require.Error(t, err)
		require.Contains(t, err.Error(), "length mismatch")
	})

	t.Run("empty expected", func(t *testing.T) {
		err := doc.VerifyPCR(4, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty")
	})

	t.Run("all-zero PCR is rejected (debug mode)", func(t *testing.T) {
		zero := make([]byte, 48)
		debugDoc := &Document{PCRs: map[uint][]byte{4: zero}}
		err := debugDoc.VerifyPCR(4, zero)
		require.Error(t, err)
		require.Contains(t, err.Error(), "all zero")
	})
}

func TestDocument_VerifyExpectedPCRs(t *testing.T) {
	pcr4 := make([]byte, 48)
	pcr8 := make([]byte, 48)
	for i := range pcr4 {
		pcr4[i] = byte(i + 1)
		pcr8[i] = byte(i + 100)
	}
	doc := &Document{PCRs: map[uint][]byte{4: pcr4, 8: pcr8}}

	require.NoError(t, doc.VerifyExpectedPCRs(map[uint][]byte{4: pcr4, 8: pcr8}))

	wrong := make([]byte, 48)
	copy(wrong, pcr8)
	wrong[0] ^= 0xff
	err := doc.VerifyExpectedPCRs(map[uint][]byte{4: pcr4, 8: wrong})
	require.Error(t, err)
	require.Contains(t, err.Error(), "PCR8 mismatch")
}
