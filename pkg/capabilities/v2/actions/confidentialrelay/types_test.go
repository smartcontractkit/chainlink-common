package confidentialrelay

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/ragep2p/peeridhelper"
)

func TestSecretsResponseResultHash_IgnoresAttestationAndBindsRequestAndResponse(t *testing.T) {
	params := SecretsRequestParams{
		WorkflowID:       "wf-1",
		Owner:            "0x1234",
		ExecutionID:      "exec-1",
		OrgID:            "org-1",
		EnclavePublicKey: "pubkey-1",
		Attestation:      "att-a",
		Secrets: []SecretIdentifier{
			{Key: "alpha", Namespace: "ns-a"},
			{Key: "beta", Namespace: "ns-b"},
		},
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

	require.Equal(t, result.Hash(params), result.Hash(sameButDifferentAttestation))

	differentRequest := params
	differentRequest.Owner = "0x9999"
	require.NotEqual(t, result.Hash(params), result.Hash(differentRequest))

	differentResponse := result
	differentResponse.Secrets = append([]SecretEntry(nil), result.Secrets...)
	differentResponse.Secrets[0].EncryptedShares = append([]string(nil), result.Secrets[0].EncryptedShares...)
	differentResponse.Secrets[0].EncryptedShares[1] = "share-a3"
	require.NotEqual(t, result.Hash(params), differentResponse.Hash(params))
}

func TestCapabilityResponseResultHash_IgnoresAttestationAndBindsRequestAndResponse(t *testing.T) {
	params := CapabilityRequestParams{
		WorkflowID:   "wf-1",
		Owner:        "0x1234",
		ExecutionID:  "exec-1",
		ReferenceID:  "42",
		CapabilityID: "write_ethereum-testnet-sepolia@1.0.0",
		Payload:      "request-payload",
		Attestation:  "att-a",
	}
	result := CapabilityResponseResult{
		Payload: "response-payload",
	}

	sameButDifferentAttestation := params
	sameButDifferentAttestation.Attestation = "att-b"
	require.Equal(t, result.Hash(params), result.Hash(sameButDifferentAttestation))

	differentRequest := params
	differentRequest.ReferenceID = "43"
	require.NotEqual(t, result.Hash(params), result.Hash(differentRequest))

	differentResponse := result
	differentResponse.Error = "boom"
	require.NotEqual(t, result.Hash(params), differentResponse.Hash(params))
}

func TestRelayResponseSignaturePayload_UsesExpectedPrefix(t *testing.T) {
	hash := [32]byte{1, 2, 3, 4}

	expected := peeridhelper.MakePeerIDSignatureDomainSeparatedPayload(RelayResponseSignaturePrefix, hash[:])
	require.Equal(t, expected, RelayResponseSignaturePayload(hash))
	require.NotEqual(t, hash[:], RelayResponseSignaturePayload(hash))
}
