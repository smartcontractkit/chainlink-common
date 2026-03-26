package fake

import (
	"testing"
	"time"

	"github.com/hf/nitrite"
	"github.com/stretchr/testify/require"
)

func TestAttestor_RoundTrip(t *testing.T) {
	fa, err := NewAttestor()
	require.NoError(t, err)

	userData := []byte("test-user-data-12345")
	attestation, err := fa.CreateAttestation(userData)
	require.NoError(t, err)
	require.NotEmpty(t, attestation)

	result, err := nitrite.Verify(attestation, nitrite.VerifyOptions{
		CurrentTime: time.Now(),
		Roots:       fa.CARoots(),
	})
	require.NoError(t, err)
	require.True(t, result.SignatureOK, "ECDSA signature should be valid")
	require.Equal(t, userData, result.Document.UserData)
	require.Equal(t, "SHA384", result.Document.Digest)
	require.Equal(t, "fake-enclave-module", result.Document.ModuleID)
	require.Len(t, result.Document.PCRs, 3)
	require.Len(t, result.Document.PCRs[0], 48)
	require.Len(t, result.Document.PCRs[1], 48)
	require.Len(t, result.Document.PCRs[2], 48)
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
