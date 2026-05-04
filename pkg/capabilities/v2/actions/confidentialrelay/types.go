package confidentialrelay

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"sort"

	"github.com/smartcontractkit/chainlink-common/pkg/teeattestation"
	"github.com/smartcontractkit/libocr/ragep2p/peeridhelper"
)

const (
	MethodSecretsGet     = "confidential.secrets.get"
	MethodCapabilityExec = "confidential.capability.execute"

	DomainSecretsGet     = "ConfidentialSecretsGet"
	DomainCapabilityExec = "ConfidentialCapabilityExecute"

	// RelayResponseSignaturePrefix domain-separates signatures over relay
	// response hashes from other ed25519 payloads in the system.
	RelayResponseSignaturePrefix = "CONFIDENTIAL_RELAY_PAYLOAD_"
)

// SecretIdentifier identifies a secret by key and namespace.
type SecretIdentifier struct {
	Key       string `json:"key"`
	Namespace string `json:"namespace"`
}

// SecretsRequestParams is the JSON-RPC params for "confidential.secrets.get".
type SecretsRequestParams struct {
	WorkflowID       string             `json:"workflow_id"`
	Owner            string             `json:"owner"`            // Ethereum address (hex, 0x-prefixed)
	ExecutionID      string             `json:"execution_id"`     // 32 bytes, hex-encoded
	OrgID            string             `json:"org_id,omitempty"` // Organization identifier for org-based secret ownership
	Secrets          []SecretIdentifier `json:"secrets"`
	EnclavePublicKey string             `json:"enclave_public_key"`
	Attestation      string             `json:"attestation,omitempty"`
}

// SecretEntry is a single secret in the relay DON's response.
type SecretEntry struct {
	ID              SecretIdentifier `json:"id"`
	Ciphertext      string           `json:"ciphertext"`
	EncryptedShares []string         `json:"encrypted_shares"`
}

// SecretsResponseResult is the JSON-RPC result for "confidential.secrets.get".
// The enclave uses its own config for MasterPublicKey and threshold (config.T),
// so the relay handler only returns the encrypted shares per secret.
type SecretsResponseResult struct {
	Secrets []SecretEntry `json:"secrets"`
}

// Hash computes the canonical hash of the caller-provided request params and
// logical relay response body. Attestation is intentionally excluded, and the
// secrets and encrypted share slices are sorted before hashing.
func (r *SecretsResponseResult) Hash(params SecretsRequestParams) [32]byte {
	h := sha256.New()
	h.Write([]byte(teeattestation.DomainSeparator))
	h.Write([]byte("\nSecretsResponseResult\n"))

	writeSecretsRequestParams(h, params)

	secrets := append([]SecretEntry(nil), r.Secrets...)
	sortSecretEntries(secrets)

	writeLengthPrefix(h, len(secrets))
	for _, secret := range secrets {
		writeSecretIdentifier(h, secret.ID)
		writeString(h, secret.Ciphertext)

		shares := append([]string(nil), secret.EncryptedShares...)
		sort.Strings(shares)

		writeLengthPrefix(h, len(shares))
		for _, share := range shares {
			writeString(h, share)
		}
	}

	var result [32]byte
	h.Sum(result[:0])
	return result
}

// CapabilityRequestParams is the JSON-RPC params for "confidential.capability.execute".
type CapabilityRequestParams struct {
	WorkflowID   string `json:"workflow_id"`
	Owner        string `json:"owner,omitempty"`
	ExecutionID  string `json:"execution_id,omitempty"`
	ReferenceID  string `json:"reference_id,omitempty"`
	CapabilityID string `json:"capability_id"`
	Payload      string `json:"payload"`
	Attestation  string `json:"attestation,omitempty"`
}

// CapabilityResponseResult is the JSON-RPC result for "confidential.capability.execute".
type CapabilityResponseResult struct {
	Payload string `json:"payload,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Hash computes the canonical hash of the caller-provided request params and
// logical relay response body. Attestation is intentionally excluded.
func (r *CapabilityResponseResult) Hash(params CapabilityRequestParams) [32]byte {
	h := sha256.New()
	h.Write([]byte(teeattestation.DomainSeparator))
	h.Write([]byte("\nCapabilityResponseResult\n"))

	writeCapabilityRequestParams(h, params)
	writeString(h, r.Payload)
	writeString(h, r.Error)

	var result [32]byte
	h.Sum(result[:0])
	return result
}

// RelayResponseSignature is a single relay-DON node signature over a relay
// response hash.
type RelayResponseSignature struct {
	Signer    []byte `json:"signer"`
	Signature []byte `json:"signature"`
}

// SignedSecretsResponseResult wraps a logical secrets response with the relay
// signatures that attest to it.
type SignedSecretsResponseResult struct {
	Result     SecretsResponseResult    `json:"result"`
	Signatures []RelayResponseSignature `json:"signatures"`
}

// SignedCapabilityResponseResult wraps a logical capability response with the
// relay signatures that attest to it.
type SignedCapabilityResponseResult struct {
	Result     CapabilityResponseResult `json:"result"`
	Signatures []RelayResponseSignature `json:"signatures"`
}

// RelayResponseSignaturePayload prepares a relay response hash for signing with
// the standard peerid domain-separated payload format.
func RelayResponseSignaturePayload(responseHash [32]byte) []byte {
	return peeridhelper.MakePeerIDSignatureDomainSeparatedPayload(RelayResponseSignaturePrefix, responseHash[:])
}

func writeSecretsRequestParams(h hash.Hash, params SecretsRequestParams) {
	writeString(h, params.WorkflowID)
	writeString(h, params.Owner)
	writeString(h, params.ExecutionID)
	writeString(h, params.OrgID)

	secrets := append([]SecretIdentifier(nil), params.Secrets...)
	sortSecretIdentifiers(secrets)

	writeLengthPrefix(h, len(secrets))
	for _, secret := range secrets {
		writeSecretIdentifier(h, secret)
	}

	writeString(h, params.EnclavePublicKey)
}

func writeCapabilityRequestParams(h hash.Hash, params CapabilityRequestParams) {
	writeString(h, params.WorkflowID)
	writeString(h, params.Owner)
	writeString(h, params.ExecutionID)
	writeString(h, params.ReferenceID)
	writeString(h, params.CapabilityID)
	writeString(h, params.Payload)
}

func writeSecretIdentifier(h hash.Hash, id SecretIdentifier) {
	writeString(h, id.Key)
	writeString(h, id.Namespace)
}

func sortSecretIdentifiers(secrets []SecretIdentifier) {
	sort.Slice(secrets, func(i, j int) bool {
		if secrets[i].Namespace != secrets[j].Namespace {
			return secrets[i].Namespace < secrets[j].Namespace
		}
		return secrets[i].Key < secrets[j].Key
	})
}

func sortSecretEntries(secrets []SecretEntry) {
	sort.Slice(secrets, func(i, j int) bool {
		if secrets[i].ID.Namespace != secrets[j].ID.Namespace {
			return secrets[i].ID.Namespace < secrets[j].ID.Namespace
		}
		if secrets[i].ID.Key != secrets[j].ID.Key {
			return secrets[i].ID.Key < secrets[j].ID.Key
		}
		return secrets[i].Ciphertext < secrets[j].Ciphertext
	})
}

func writeString(h hash.Hash, s string) {
	writeBytes(h, []byte(s))
}

func writeBytes(h hash.Hash, b []byte) {
	writeLengthPrefix(h, len(b))
	h.Write(b)
}

func writeLengthPrefix(h hash.Hash, length int) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(length))
	h.Write(buf[:])
}
