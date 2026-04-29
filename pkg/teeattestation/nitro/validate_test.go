package nitro

import (
	"testing"

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
