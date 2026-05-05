package confidentialrelay

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/ragep2p/peeridhelper"
)

func mustSecretsHash(t *testing.T, r SecretsResponseResult, p SecretsRequestParams) [32]byte {
	t.Helper()
	h, err := r.Hash(p)
	require.NoError(t, err)
	return h
}

func mustCapabilityHash(t *testing.T, r CapabilityResponseResult, p CapabilityRequestParams) [32]byte {
	t.Helper()
	h, err := r.Hash(p)
	require.NoError(t, err)
	return h
}

func validSecretsParams() SecretsRequestParams {
	return SecretsRequestParams{
		WorkflowID:       "wf-1",
		Owner:            "0x1234",
		ExecutionID:      "exec-1",
		OrgID:            "org-1",
		EnclavePublicKey: "pubkey-1",
		Attestation:      "att-a",
		Secrets: []SecretIdentifier{
			{Key: "alpha", Namespace: "ns-a"},
		},
	}
}

func validCapabilityParams() CapabilityRequestParams {
	return CapabilityRequestParams{
		WorkflowID:   "wf-1",
		Owner:        "0x1234",
		ExecutionID:  "exec-1",
		ReferenceID:  "42",
		CapabilityID: "write_ethereum-testnet-sepolia@1.0.0",
		Payload:      "request-payload",
		Attestation:  "att-a",
	}
}

func TestSecretsResponseResultHash_IgnoresAttestationAndBindsRequestAndResponse(t *testing.T) {
	params := validSecretsParams()
	params.Secrets = []SecretIdentifier{
		{Key: "alpha", Namespace: "ns-a"},
		{Key: "beta", Namespace: "ns-b"},
	}
	result := SecretsResponseResult{
		Secrets: []SecretEntry{
			{
				ID:         SecretIdentifier{Key: "alpha", Namespace: "ns-a"},
				Ciphertext: "cipher-a",
				EncryptedShares: []string{
					"share-a1",
					"share-a2",
				},
			},
		},
	}

	sameButDifferentAttestation := params
	sameButDifferentAttestation.Attestation = "att-b"
	require.Equal(t, mustSecretsHash(t, result, params), mustSecretsHash(t, result, sameButDifferentAttestation))

	differentRequest := params
	differentRequest.Owner = "0x9999"
	require.NotEqual(t, mustSecretsHash(t, result, params), mustSecretsHash(t, result, differentRequest))

	differentResponse := result
	differentResponse.Secrets = append([]SecretEntry(nil), result.Secrets...)
	differentResponse.Secrets[0].EncryptedShares = append([]string(nil), result.Secrets[0].EncryptedShares...)
	differentResponse.Secrets[0].EncryptedShares[1] = "share-a3"
	require.NotEqual(t, mustSecretsHash(t, result, params), mustSecretsHash(t, differentResponse, params))
}

func TestSecretsResponseResultHash_IsStableUnderSecretsAndSharesReordering(t *testing.T) {
	paramsA := validSecretsParams()
	paramsA.Secrets = []SecretIdentifier{
		{Key: "beta", Namespace: "ns-b"},
		{Key: "alpha", Namespace: "ns-a"},
	}
	paramsB := validSecretsParams()
	paramsB.Secrets = []SecretIdentifier{
		{Key: "alpha", Namespace: "ns-a"},
		{Key: "beta", Namespace: "ns-b"},
	}
	resultA := SecretsResponseResult{
		Secrets: []SecretEntry{
			{
				ID:         SecretIdentifier{Key: "beta", Namespace: "ns-b"},
				Ciphertext: "cipher-b",
				EncryptedShares: []string{
					"share-b2",
					"share-b1",
				},
			},
			{
				ID:         SecretIdentifier{Key: "alpha", Namespace: "ns-a"},
				Ciphertext: "cipher-a",
				EncryptedShares: []string{
					"share-a2",
					"share-a1",
				},
			},
		},
	}
	resultB := SecretsResponseResult{
		Secrets: []SecretEntry{
			{
				ID:         SecretIdentifier{Key: "alpha", Namespace: "ns-a"},
				Ciphertext: "cipher-a",
				EncryptedShares: []string{
					"share-a1",
					"share-a2",
				},
			},
			{
				ID:         SecretIdentifier{Key: "beta", Namespace: "ns-b"},
				Ciphertext: "cipher-b",
				EncryptedShares: []string{
					"share-b1",
					"share-b2",
				},
			},
		},
	}

	require.Equal(t, mustSecretsHash(t, resultA, paramsA), mustSecretsHash(t, resultB, paramsB))
}

func TestCapabilityResponseResultHash_IgnoresAttestationAndBindsRequestAndResponse(t *testing.T) {
	params := validCapabilityParams()
	result := CapabilityResponseResult{
		Payload: "response-payload",
	}

	sameButDifferentAttestation := params
	sameButDifferentAttestation.Attestation = "att-b"
	require.Equal(t, mustCapabilityHash(t, result, params), mustCapabilityHash(t, result, sameButDifferentAttestation))

	differentRequest := params
	differentRequest.ReferenceID = "43"
	require.NotEqual(t, mustCapabilityHash(t, result, params), mustCapabilityHash(t, result, differentRequest))

	differentResponse := result
	differentResponse.Error = "boom"
	require.NotEqual(t, mustCapabilityHash(t, result, params), mustCapabilityHash(t, differentResponse, params))
}

func TestSecretsRequestParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*SecretsRequestParams)
		wantErr string
	}{
		{"missing workflow_id", func(p *SecretsRequestParams) { p.WorkflowID = "" }, "workflow_id is required"},
		{"missing owner", func(p *SecretsRequestParams) { p.Owner = "" }, "owner is required"},
		{"missing execution_id", func(p *SecretsRequestParams) { p.ExecutionID = "" }, "execution_id is required"},
		{"missing enclave_public_key", func(p *SecretsRequestParams) { p.EnclavePublicKey = "" }, "enclave_public_key is required"},
		{"empty secrets slice", func(p *SecretsRequestParams) { p.Secrets = nil }, "secrets must be non-empty"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := validSecretsParams()
			tc.mutate(&p)
			err := p.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}

	t.Run("valid params accepted", func(t *testing.T) {
		require.NoError(t, validSecretsParams().Validate())
	})

	t.Run("optional fields can be empty", func(t *testing.T) {
		p := validSecretsParams()
		p.OrgID = ""
		p.Attestation = ""
		require.NoError(t, p.Validate())
	})
}

func TestCapabilityRequestParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*CapabilityRequestParams)
		wantErr string
	}{
		{"missing workflow_id", func(p *CapabilityRequestParams) { p.WorkflowID = "" }, "workflow_id is required"},
		{"missing owner", func(p *CapabilityRequestParams) { p.Owner = "" }, "owner is required"},
		{"missing execution_id", func(p *CapabilityRequestParams) { p.ExecutionID = "" }, "execution_id is required"},
		{"missing reference_id", func(p *CapabilityRequestParams) { p.ReferenceID = "" }, "reference_id is required"},
		{"missing capability_id", func(p *CapabilityRequestParams) { p.CapabilityID = "" }, "capability_id is required"},
		{"missing payload", func(p *CapabilityRequestParams) { p.Payload = "" }, "payload is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := validCapabilityParams()
			tc.mutate(&p)
			err := p.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}

	t.Run("valid params accepted", func(t *testing.T) {
		require.NoError(t, validCapabilityParams().Validate())
	})

	t.Run("attestation can be empty", func(t *testing.T) {
		p := validCapabilityParams()
		p.Attestation = ""
		require.NoError(t, p.Validate())
	})
}

func TestSecretsResponseResultHash_RejectsInvalidParams(t *testing.T) {
	result := SecretsResponseResult{
		Secrets: []SecretEntry{{ID: SecretIdentifier{Key: "k", Namespace: "n"}, Ciphertext: "ct"}},
	}
	params := validSecretsParams()
	params.ExecutionID = ""

	_, err := result.Hash(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "execution_id is required")
}

func TestCapabilityResponseResultHash_RejectsInvalidParams(t *testing.T) {
	result := CapabilityResponseResult{Payload: "out"}
	params := validCapabilityParams()
	params.ReferenceID = ""

	_, err := result.Hash(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reference_id is required")
}

func TestRelayResponseSignaturePayload_UsesExpectedPrefix(t *testing.T) {
	hash := [32]byte{1, 2, 3, 4}

	expected := peeridhelper.MakePeerIDSignatureDomainSeparatedPayload(RelayResponseSignaturePrefix, hash[:])
	require.Equal(t, expected, RelayResponseSignaturePayload(hash))
	require.NotEqual(t, hash[:], RelayResponseSignaturePayload(hash))
}
