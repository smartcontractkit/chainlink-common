package nitro

import (
	"testing"

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
