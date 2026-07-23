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

	userData, err := teeattestation.DomainHash("testtag", []byte(`{"key":"value"}`))
	require.NoError(t, err)
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

	userData, err := teeattestation.DomainHash("testtag", []byte(`{"key":"value"}`))
	require.NoError(t, err)
	doc, err := fa.CreateAttestation(userData)
	require.NoError(t, err)

	wrongData, err := teeattestation.DomainHash("wrongtag", []byte(`{"key":"value"}`))
	require.NoError(t, err)
	err = ValidateAttestationWithRoots(doc, wrongData, fa.TrustedPCRsJSON(), fa.CARootsPEM())
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected user data")
}

func TestValidateAttestation_WrongPCRs(t *testing.T) {
	fa, err := nitrofake.NewAttestor()
	require.NoError(t, err)

	userData := []byte("testdata")
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

	userData, err := teeattestation.DomainHash("testtag", []byte(`{"key":"value"}`))
	require.NoError(t, err)
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

	userData, err := teeattestation.DomainHash("testtag", []byte(`{"key":"value"}`))
	require.NoError(t, err)
	doc, err := fa.CreateAttestation(userData, nitrofake.WithNonce([]byte("n")))
	require.NoError(t, err)

	parsed, err := ValidateAndParseWithRoots(doc, []byte("wrong"), fa.TrustedPCRsJSON(), fa.CARootsPEM())
	require.Error(t, err)
	require.Nil(t, parsed)
	require.Contains(t, err.Error(), "expected user data")
}
