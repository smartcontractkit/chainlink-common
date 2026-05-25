package nitrofake

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/teeattestation/nitro"
)

func TestAttestor_RoundTrip(t *testing.T) {
	fa, err := NewAttestor()
	require.NoError(t, err)

	userData := []byte("test-user-data-12345")
	attestation, err := fa.CreateAttestation(userData)
	require.NoError(t, err)
	require.NotEmpty(t, attestation)

	err = nitro.ValidateAttestationWithRoots(attestation, userData, fa.TrustedPCRsJSON(), fa.CARootsPEM())
	require.NoError(t, err)
}

func TestAttestor_TrustedPCRsJSON(t *testing.T) {
	fa, err := NewAttestor()
	require.NoError(t, err)

	pcrsJSON := fa.TrustedPCRsJSON()
	require.NotEmpty(t, pcrsJSON)
	require.Contains(t, string(pcrsJSON), `"pcr0"`)
	require.Contains(t, string(pcrsJSON), `"pcr1"`)
	require.Contains(t, string(pcrsJSON), `"pcr2"`)
}

func TestAttestor_CARootsPEM(t *testing.T) {
	fa, err := NewAttestor()
	require.NoError(t, err)

	pemStr := fa.CARootsPEM()
	require.Contains(t, pemStr, "BEGIN CERTIFICATE")
	require.Contains(t, pemStr, "END CERTIFICATE")
}
