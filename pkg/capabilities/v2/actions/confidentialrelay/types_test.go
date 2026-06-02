package confidentialrelay

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/ragep2p/peeridhelper"
)

const (
	validOwnerA        = "0x1111111111111111111111111111111111111111"
	validOwnerB        = "0x2222222222222222222222222222222222222222"
	validExecutionID   = "1111111111111111111111111111111111111111111111111111111111111111"
	validEnclavePubKey = "deadbeef"
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

func validEnclaveConfig() EnclaveConfig {
	return EnclaveConfig{
		Signers: [][]byte{
			{0x01, 0x02, 0x03},
			{0x04, 0x05, 0x06},
			{0x07, 0x08, 0x09},
			{0x0a, 0x0b, 0x0c},
		},
		MasterPublicKey: []byte("master-public-key"),
		T:               2,
		F:               1,
	}
}

func validEnclaveConfigPtr() *EnclaveConfig {
	c := validEnclaveConfig()
	return &c
}

func validSecretsParams() SecretsRequestParams {
	return SecretsRequestParams{
		WorkflowID:       "wf-1",
		Owner:            validOwnerA,
		ExecutionID:      validExecutionID,
		OrgID:            "org-1",
		EnclavePublicKey: validEnclavePubKey,
		EnclaveConfig:    validEnclaveConfigPtr(),
		Attestation:      "att-a",
		Secrets: []SecretIdentifier{
			{Key: "alpha", Namespace: "ns-a"},
		},
	}
}

func validCapabilityParams() CapabilityRequestParams {
	return CapabilityRequestParams{
		WorkflowID:    "wf-1",
		Owner:         validOwnerA,
		ExecutionID:   validExecutionID,
		ReferenceID:   "42",
		CapabilityID:  "write_ethereum-testnet-sepolia@1.0.0",
		Payload:       "request-payload",
		EnclaveConfig: validEnclaveConfigPtr(),
		Attestation:   "att-a",
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
	differentRequest.Owner = validOwnerB
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

	differentOrg := params
	differentOrg.OrgID = "org-other"
	require.NotEqual(t, mustCapabilityHash(t, result, params), mustCapabilityHash(t, result, differentOrg))

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
		{"owner without 0x prefix", func(p *SecretsRequestParams) {
			p.Owner = "1111111111111111111111111111111111111111"
		}, "owner must be a 0x-prefixed 20-byte hex address"},
		{"owner wrong length", func(p *SecretsRequestParams) { p.Owner = "0x1234" }, "owner must be a 0x-prefixed 20-byte hex address"},
		{"owner non-hex digits", func(p *SecretsRequestParams) {
			p.Owner = "0xZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"
		}, "owner must be a 0x-prefixed 20-byte hex address"},
		{"missing execution_id", func(p *SecretsRequestParams) { p.ExecutionID = "" }, "execution_id is required"},
		{"execution_id wrong length", func(p *SecretsRequestParams) { p.ExecutionID = "abcd" }, "execution_id must be 32 bytes hex-encoded"},
		{"execution_id with 0x prefix", func(p *SecretsRequestParams) {
			p.ExecutionID = "0x" + validExecutionID[:62]
		}, "execution_id must be 32 bytes hex-encoded"},
		{"execution_id non-hex digits", func(p *SecretsRequestParams) {
			p.ExecutionID = "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"
		}, "execution_id must be 32 bytes hex-encoded"},
		{"missing enclave_public_key", func(p *SecretsRequestParams) { p.EnclavePublicKey = "" }, "enclave_public_key is required"},
		{"enclave_public_key non-hex digits", func(p *SecretsRequestParams) {
			p.EnclavePublicKey = "not-hex"
		}, "enclave_public_key must be hex-encoded"},
		{"empty secrets slice", func(p *SecretsRequestParams) { p.Secrets = nil }, "secrets must be non-empty"},
		{"secret with empty key", func(p *SecretsRequestParams) {
			p.Secrets = []SecretIdentifier{{Key: "", Namespace: "ns"}}
		}, "secrets[0].key is required"},
		{"secret with empty namespace", func(p *SecretsRequestParams) {
			p.Secrets = []SecretIdentifier{{Key: "k", Namespace: ""}}
		}, "secrets[0].namespace is required"},
		{"second secret with empty key reports its index", func(p *SecretsRequestParams) {
			p.Secrets = []SecretIdentifier{
				{Key: "k", Namespace: "ns"},
				{Key: "", Namespace: "ns"},
			}
		}, "secrets[1].key is required"},
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
		{"owner wrong length", func(p *CapabilityRequestParams) { p.Owner = "0x1234" }, "owner must be a 0x-prefixed 20-byte hex address"},
		{"owner without 0x prefix", func(p *CapabilityRequestParams) {
			p.Owner = "1111111111111111111111111111111111111111"
		}, "owner must be a 0x-prefixed 20-byte hex address"},
		{"missing execution_id", func(p *CapabilityRequestParams) { p.ExecutionID = "" }, "execution_id is required"},
		{"execution_id wrong length", func(p *CapabilityRequestParams) { p.ExecutionID = "abcd" }, "execution_id must be 32 bytes hex-encoded"},
		{"execution_id non-hex digits", func(p *CapabilityRequestParams) {
			p.ExecutionID = "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"
		}, "execution_id must be 32 bytes hex-encoded"},
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

// TestValidateEnclaveConfig covers the EnclaveConfig validation added for
// PRIV-458. The relay needs each request to carry a non-empty
// signers list, non-zero F, and a non-empty MasterPublicKey so it can
// meaningfully compare against onchain DON state.
func TestValidateEnclaveConfig(t *testing.T) {
	t.Run("valid config accepted", func(t *testing.T) {
		require.NoError(t, validateEnclaveConfig(validEnclaveConfig()))
	})
	t.Run("missing signers rejected", func(t *testing.T) {
		c := validEnclaveConfig()
		c.Signers = nil
		require.Error(t, validateEnclaveConfig(c))
	})
	t.Run("empty signer rejected", func(t *testing.T) {
		c := validEnclaveConfig()
		c.Signers = append(c.Signers, []byte{})
		err := validateEnclaveConfig(c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "signers[")
	})
	t.Run("F=0 rejected", func(t *testing.T) {
		c := validEnclaveConfig()
		c.F = 0
		require.Error(t, validateEnclaveConfig(c))
	})
	t.Run("empty master_public_key rejected", func(t *testing.T) {
		c := validEnclaveConfig()
		c.MasterPublicKey = nil
		require.Error(t, validateEnclaveConfig(c))
	})
}

// TestSecretsRequestParams_Validate_EnclaveConfigOptional covers that a nil
// EnclaveConfig is accepted (sender on an older protocol), while a present
// but invalid one is still rejected.
func TestSecretsRequestParams_Validate_EnclaveConfigOptional(t *testing.T) {
	p := validSecretsParams()
	p.EnclaveConfig = nil
	require.NoError(t, p.Validate())

	p.EnclaveConfig = &EnclaveConfig{} // present but empty
	require.Error(t, p.Validate())
}

// TestCapabilityRequestParams_Validate_EnclaveConfigOptional same as above
// for the capability execute path.
func TestCapabilityRequestParams_Validate_EnclaveConfigOptional(t *testing.T) {
	p := validCapabilityParams()
	p.EnclaveConfig = nil
	require.NoError(t, p.Validate())

	p.EnclaveConfig = &EnclaveConfig{} // present but empty
	require.Error(t, p.Validate())
}

// TestSecretsResponseHash_NilEnclaveConfig proves that a nil EnclaveConfig is
// accepted (Hash returns no error, exercised by mustSecretsHash) and hashes
// deterministically across independent constructions of the same params.
func TestSecretsResponseHash_NilEnclaveConfig(t *testing.T) {
	result := SecretsResponseResult{
		Secrets: []SecretEntry{
			{ID: SecretIdentifier{Key: "alpha", Namespace: "ns-a"}, Ciphertext: "c", EncryptedShares: []string{"s"}},
		},
	}

	p1 := validSecretsParams()
	p1.EnclaveConfig = nil
	p2 := validSecretsParams()
	p2.EnclaveConfig = nil
	require.Equal(t, mustSecretsHash(t, result, p1), mustSecretsHash(t, result, p2))
}

// TestCapabilityResponseHash_NilEnclaveConfig is the capability-path counterpart:
// a nil EnclaveConfig is accepted and hashes deterministically.
func TestCapabilityResponseHash_NilEnclaveConfig(t *testing.T) {
	result := CapabilityResponseResult{Payload: "out"}

	p1 := validCapabilityParams()
	p1.EnclaveConfig = nil
	p2 := validCapabilityParams()
	p2.EnclaveConfig = nil
	require.Equal(t, mustCapabilityHash(t, result, p1), mustCapabilityHash(t, result, p2))
}

// TestSecretsResponseHash_BindsEnclaveConfig proves the response signature
// hash differs when EnclaveConfig differs. If the hash did not bind
// EnclaveConfig, two responses signed over the same secrets but with
// different enclave configs would have indistinguishable signatures.
func TestSecretsResponseHash_BindsEnclaveConfig(t *testing.T) {
	params := validSecretsParams()
	result := SecretsResponseResult{
		Secrets: []SecretEntry{
			{
				ID:              SecretIdentifier{Key: "alpha", Namespace: "ns-a"},
				Ciphertext:      "cipher-a",
				EncryptedShares: []string{"share-a1"},
			},
		},
	}
	base := mustSecretsHash(t, result, params)

	changed := params
	c1 := *params.EnclaveConfig
	c1.F = params.EnclaveConfig.F + 1
	changed.EnclaveConfig = &c1
	require.NotEqual(t, base, mustSecretsHash(t, result, changed))

	changed2 := params
	c2 := *params.EnclaveConfig
	c2.T = params.EnclaveConfig.T + 1
	changed2.EnclaveConfig = &c2
	require.NotEqual(t, base, mustSecretsHash(t, result, changed2))

	changed3 := params
	c3 := *params.EnclaveConfig
	c3.MasterPublicKey = append([]byte(nil), params.EnclaveConfig.MasterPublicKey...)
	c3.MasterPublicKey[0] ^= 0xff
	changed3.EnclaveConfig = &c3
	require.NotEqual(t, base, mustSecretsHash(t, result, changed3))

	changed4 := params
	c4 := *params.EnclaveConfig
	c4.Signers = append([][]byte(nil), params.EnclaveConfig.Signers...)
	c4.Signers = append(c4.Signers, []byte{0xff})
	changed4.EnclaveConfig = &c4
	require.NotEqual(t, base, mustSecretsHash(t, result, changed4))
}

// TestCapabilityResponseHash_BindsEnclaveConfig same as above for the
// capability execute path.
func TestCapabilityResponseHash_BindsEnclaveConfig(t *testing.T) {
	params := validCapabilityParams()
	result := CapabilityResponseResult{Payload: "out"}
	base := mustCapabilityHash(t, result, params)

	changed := params
	c := *params.EnclaveConfig
	c.F = params.EnclaveConfig.F + 1
	changed.EnclaveConfig = &c
	require.NotEqual(t, base, mustCapabilityHash(t, result, changed))
}

// TestSecretsResponseHash_StableUnderSignerReordering proves that the
// EnclaveConfig.Signers ordering does not affect the hash. The relay-side
// comparison against onchain state is order-independent so the hash must be
// too, otherwise an enclave permuting Signers (or a different ordering
// emitted across reboots) would invalidate signatures over identical
// logical state.
func TestSecretsResponseHash_StableUnderSignerReordering(t *testing.T) {
	params := validSecretsParams()
	result := SecretsResponseResult{
		Secrets: []SecretEntry{
			{ID: SecretIdentifier{Key: "alpha", Namespace: "ns-a"}, Ciphertext: "c", EncryptedShares: []string{"s"}},
		},
	}
	base := mustSecretsHash(t, result, params)

	reversed := params
	rc := *params.EnclaveConfig
	rc.Signers = make([][]byte, len(params.EnclaveConfig.Signers))
	for i, s := range params.EnclaveConfig.Signers {
		rc.Signers[len(params.EnclaveConfig.Signers)-1-i] = s
	}
	reversed.EnclaveConfig = &rc
	require.Equal(t, base, mustSecretsHash(t, result, reversed))
}
