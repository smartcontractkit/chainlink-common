package confidentialrelay

import (
	"strings"
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

func validSecretsParams() SecretsRequestParams {
	return SecretsRequestParams{
		WorkflowID:       "wf-1",
		Owner:            validOwnerA,
		ExecutionID:      validExecutionID,
		OrgID:            "org-1",
		EnclavePublicKey: validEnclavePubKey,
		EnclaveConfig:    validEnclaveConfig(),
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
		EnclaveConfig: validEnclaveConfig(),
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

// TestSecretsRequestParams_Validate_RequiresEnclaveConfig covers that the
// Validate gate rejects requests missing the EnclaveConfig.
func TestSecretsRequestParams_Validate_RequiresEnclaveConfig(t *testing.T) {
	p := validSecretsParams()
	p.EnclaveConfig = EnclaveConfig{}
	require.Error(t, p.Validate())
}

// TestCapabilityRequestParams_Validate_RequiresEnclaveConfig same as above
// for the capability execute path.
func TestCapabilityRequestParams_Validate_RequiresEnclaveConfig(t *testing.T) {
	p := validCapabilityParams()
	p.EnclaveConfig = EnclaveConfig{}
	require.Error(t, p.Validate())
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
	changed.EnclaveConfig.F = params.EnclaveConfig.F + 1
	require.NotEqual(t, base, mustSecretsHash(t, result, changed))

	changed2 := params
	changed2.EnclaveConfig.T = params.EnclaveConfig.T + 1
	require.NotEqual(t, base, mustSecretsHash(t, result, changed2))

	changed3 := params
	changed3.EnclaveConfig.MasterPublicKey = append([]byte(nil), params.EnclaveConfig.MasterPublicKey...)
	changed3.EnclaveConfig.MasterPublicKey[0] ^= 0xff
	require.NotEqual(t, base, mustSecretsHash(t, result, changed3))

	changed4 := params
	changed4.EnclaveConfig.Signers = append([][]byte(nil), params.EnclaveConfig.Signers...)
	changed4.EnclaveConfig.Signers = append(changed4.EnclaveConfig.Signers, []byte{0xff})
	require.NotEqual(t, base, mustSecretsHash(t, result, changed4))
}

// TestCapabilityResponseHash_BindsEnclaveConfig same as above for the
// capability execute path.
func TestCapabilityResponseHash_BindsEnclaveConfig(t *testing.T) {
	params := validCapabilityParams()
	result := CapabilityResponseResult{Payload: "out"}
	base := mustCapabilityHash(t, result, params)

	changed := params
	changed.EnclaveConfig.F = params.EnclaveConfig.F + 1
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
	reversed.EnclaveConfig.Signers = make([][]byte, len(params.EnclaveConfig.Signers))
	for i, s := range params.EnclaveConfig.Signers {
		reversed.EnclaveConfig.Signers[len(params.EnclaveConfig.Signers)-1-i] = s
	}
	require.Equal(t, base, mustSecretsHash(t, result, reversed))
}

func validWorkflowAuthz() WorkflowAuthz {
	return WorkflowAuthz{
		Owner:       validOwnerA,
		OrgID:       "org-1",
		WorkflowID:  "wf-1",
		ExecutionID: validExecutionID,
	}
}

func mustWorkflowAuthzHash(t *testing.T, w WorkflowAuthz) [32]byte {
	t.Helper()
	h, err := w.Hash()
	require.NoError(t, err)
	return h
}

func TestWorkflowAuthz_Hash_Deterministic(t *testing.T) {
	w := validWorkflowAuthz()
	require.Equal(t, mustWorkflowAuthzHash(t, w), mustWorkflowAuthzHash(t, w))
}

// Every field WorkflowAuthz claims to bind must actually change the hash, or a
// compromised enclave could mutate that field without invalidating the F+1
// signatures the relay verifies.
func TestWorkflowAuthz_Hash_BindsEveryField(t *testing.T) {
	base := mustWorkflowAuthzHash(t, validWorkflowAuthz())

	mutations := map[string]func(*WorkflowAuthz){
		"owner":       func(w *WorkflowAuthz) { w.Owner = validOwnerB },
		"org_id":      func(w *WorkflowAuthz) { w.OrgID = "org-2" },
		"workflow_id": func(w *WorkflowAuthz) { w.WorkflowID = "wf-2" },
		"execution_id": func(w *WorkflowAuthz) {
			w.ExecutionID = "2222222222222222222222222222222222222222222222222222222222222222"
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			w := validWorkflowAuthz()
			mutate(&w)
			require.NotEqual(t, base, mustWorkflowAuthzHash(t, w), "hash must change when %s changes", name)
		})
	}
}

// Owner and ExecutionID are hex with case-insensitive validators, so the hash must
// be invariant to hex case or a signer and verifier could disagree.
func TestWorkflowAuthz_Hash_CanonicalHexCase(t *testing.T) {
	lower := validWorkflowAuthz()
	lower.Owner = "0x" + strings.Repeat("a", 40)
	lower.ExecutionID = strings.Repeat("a", 64)

	upper := lower
	upper.Owner = "0x" + strings.Repeat("A", 40)
	upper.ExecutionID = strings.Repeat("A", 64)

	require.Equal(t, mustWorkflowAuthzHash(t, lower), mustWorkflowAuthzHash(t, upper))
}

// The relay reconstructs WorkflowAuthz from the request the enclave forwards. Its
// hash must equal the one the Workflow DON signed, or a faithful request would
// fail verification. This is the core round-trip the whole scheme rests on.
func TestWorkflowAuthz_ReconstructionMatchesSignedHash(t *testing.T) {
	w := validWorkflowAuthz()
	signed := mustWorkflowAuthzHash(t, w)

	p := SecretsRequestParams{
		WorkflowID:  w.WorkflowID,
		Owner:       w.Owner,
		ExecutionID: w.ExecutionID,
		OrgID:       w.OrgID,
	}
	require.Equal(t, w, p.WorkflowAuthz())
	require.Equal(t, signed, mustWorkflowAuthzHash(t, p.WorkflowAuthz()))
}

func TestWorkflowAuthz_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.NoError(t, validWorkflowAuthz().Validate())
	})
	t.Run("empty org_id is allowed", func(t *testing.T) {
		w := validWorkflowAuthz()
		w.OrgID = ""
		require.NoError(t, w.Validate())
	})

	cases := map[string]func(*WorkflowAuthz){
		"empty owner":            func(w *WorkflowAuthz) { w.Owner = "" },
		"malformed owner":        func(w *WorkflowAuthz) { w.Owner = "0xnothex" },
		"empty workflow_id":      func(w *WorkflowAuthz) { w.WorkflowID = "" },
		"empty execution_id":     func(w *WorkflowAuthz) { w.ExecutionID = "" },
		"malformed execution_id": func(w *WorkflowAuthz) { w.ExecutionID = "abc" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			w := validWorkflowAuthz()
			mutate(&w)
			require.Error(t, w.Validate())
		})
	}
}

func TestWorkflowAuthz_Hash_RejectsInvalid(t *testing.T) {
	w := validWorkflowAuthz()
	w.Owner = ""
	_, err := w.Hash()
	require.Error(t, err)
}

// The signing payload must be domain-separated from relay-response signatures so
// a WorkflowAuthz signature can never be replayed as one (and vice versa).
func TestWorkflowAuthzSignaturePayload_DomainSeparated(t *testing.T) {
	h := mustWorkflowAuthzHash(t, validWorkflowAuthz())

	got := WorkflowAuthzSignaturePayload(h)
	require.Equal(t, peeridhelper.MakePeerIDSignatureDomainSeparatedPayload(WorkflowAuthzSignaturePrefix, h[:]), got)
	require.NotEqual(t, RelayResponseSignaturePayload(h), got)
}
