package confidentialrelay

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
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

// EnclaveConfig mirrors the confidential-compute EnclaveConfig fields the
// relay needs to verify against onchain DON state. The enclave fills this
// from its local types.EnclaveConfig before sending each relay request.
//
// PRIV-458: without this field the request's Nitro
// attestation cryptographically binds the request hash but does not let the
// relay compare the enclave's config against any reference. A malicious host
// can produce genuinely-attested requests over a forged enclave config and
// have them accepted unless the relay can see and verify the config value.
type EnclaveConfig struct {
	Signers         [][]byte `json:"signers"`
	MasterPublicKey []byte   `json:"master_public_key"`
	T               uint32   `json:"t"`
	F               uint32   `json:"f"`
}

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
	// EnclaveConfig is the enclave's current config, included so the relay can
	// verify it against onchain DON state after attestation validation. See
	// the EnclaveConfig type doc-comment for the threat model.
	//
	// Optional: a nil EnclaveConfig means the sender did not provide one (e.g.
	// an enclave on an older protocol). Validate and the canonical hash skip it
	// when nil, so a verifier on this version stays compatible with senders
	// that omit the field. When present, it is validated and hash-bound.
	EnclaveConfig *EnclaveConfig `json:"enclave_config,omitempty"`
	Attestation   string         `json:"attestation,omitempty"`

	// SignedComputeRequests carries the F+1 Workflow-DON-signed compute requests the
	// enclave forwards verbatim. The relay DON verifies the signatures over
	// ComputeRequest.Hash() against its Workflow DON signer set and reads the
	// authorized identity from PublicData (the WorkflowExecution proto). Like
	// Attestation, it is authorization input and is excluded from the response hash.
	SignedComputeRequests []SignedComputeRequest `json:"signed_compute_requests,omitempty"`
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

// Validate rejects request params that are missing fields the canonical hash binds to,
// or that carry a malformed value for a structured field. Without these fields a
// signature would cover a less-unique context than the caller believes, which would
// let a malicious gateway replay it across requests; alternate encodings of structured
// fields would similarly let two distinct logical requests collide on the same hash.
func (p SecretsRequestParams) Validate() error {
	if p.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}
	if p.Owner == "" {
		return errors.New("owner is required")
	}
	if err := validateOwnerAddress(p.Owner); err != nil {
		return err
	}
	if p.ExecutionID == "" {
		return errors.New("execution_id is required")
	}
	if err := validateExecutionID(p.ExecutionID); err != nil {
		return err
	}
	if p.EnclavePublicKey == "" {
		return errors.New("enclave_public_key is required")
	}
	if err := validateEnclavePublicKey(p.EnclavePublicKey); err != nil {
		return err
	}
	if len(p.Secrets) == 0 {
		return errors.New("secrets must be non-empty")
	}
	if err := validateSecretIdentifiers(p.Secrets); err != nil {
		return err
	}
	if p.EnclaveConfig != nil {
		if err := validateEnclaveConfig(*p.EnclaveConfig); err != nil {
			return err
		}
	}
	return nil
}

// Hash computes the canonical hash of the caller-provided request params and
// logical relay response body. Attestation is intentionally excluded, and the
// secrets and encrypted share slices are sorted before hashing. Returns an
// error if params fails Validate so a caller cannot accidentally produce a
// signature over an unbinding payload.
func (r *SecretsResponseResult) Hash(params SecretsRequestParams) ([32]byte, error) {
	if err := params.Validate(); err != nil {
		return [32]byte{}, fmt.Errorf("invalid secrets request params: %w", err)
	}

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
	return result, nil
}

// CapabilityRequestParams is the JSON-RPC params for "confidential.capability.execute".
// Owner, ExecutionID, and ReferenceID are required: they bind a relay-DON signature to a
// specific (workflow, execution, step) tuple. Without them the same signature could be
// replayed across distinct logical requests that happened to share the remaining fields.
type CapabilityRequestParams struct {
	WorkflowID   string `json:"workflow_id"`
	Owner        string `json:"owner"`
	ExecutionID  string `json:"execution_id"`
	OrgID        string `json:"org_id,omitempty"` // propagated into capability.RequestMetadata when CRE setting enables it
	ReferenceID  string `json:"reference_id"`
	CapabilityID string `json:"capability_id"`
	Payload      string `json:"payload"`
	// EnclaveConfig is the enclave's current config, included so the relay can
	// verify it against onchain DON state after attestation validation. See
	// the EnclaveConfig type doc-comment for the threat model.
	//
	// Optional: a nil EnclaveConfig means the sender did not provide one (e.g.
	// an enclave on an older protocol). Validate and the canonical hash skip it
	// when nil, so a verifier on this version stays compatible with senders
	// that omit the field. When present, it is validated and hash-bound.
	EnclaveConfig *EnclaveConfig `json:"enclave_config,omitempty"`
	Attestation   string         `json:"attestation,omitempty"`
}

// CapabilityResponseResult is the JSON-RPC result for "confidential.capability.execute".
type CapabilityResponseResult struct {
	Payload string `json:"payload,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Validate rejects request params that are missing fields the canonical hash binds to,
// or that carry a malformed value for a structured field. Without these fields a
// signature would cover a less-unique context than the caller believes, which would
// let a malicious gateway replay it across requests; alternate encodings of structured
// fields would similarly let two distinct logical requests collide on the same hash.
func (p CapabilityRequestParams) Validate() error {
	if p.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}
	if p.Owner == "" {
		return errors.New("owner is required")
	}
	if err := validateOwnerAddress(p.Owner); err != nil {
		return err
	}
	if p.ExecutionID == "" {
		return errors.New("execution_id is required")
	}
	if err := validateExecutionID(p.ExecutionID); err != nil {
		return err
	}
	if p.ReferenceID == "" {
		return errors.New("reference_id is required")
	}
	if p.CapabilityID == "" {
		return errors.New("capability_id is required")
	}
	if p.Payload == "" {
		return errors.New("payload is required")
	}
	if p.EnclaveConfig != nil {
		if err := validateEnclaveConfig(*p.EnclaveConfig); err != nil {
			return err
		}
	}
	return nil
}

// validateOwnerAddress enforces the canonical "0x-prefixed 20-byte hex" Ethereum address
// shape so two encodings of the same address (e.g., differing case or a missing prefix)
// cannot produce different hashes.
func validateOwnerAddress(s string) error {
	if len(s) != 42 || s[:2] != "0x" {
		return errors.New("owner must be a 0x-prefixed 20-byte hex address")
	}
	if _, err := hex.DecodeString(s[2:]); err != nil {
		return errors.New("owner must be a 0x-prefixed 20-byte hex address")
	}
	return nil
}

// validateExecutionID enforces 32-byte hex with no prefix.
func validateExecutionID(s string) error {
	if len(s) != 64 {
		return errors.New("execution_id must be 32 bytes hex-encoded (64 hex chars, no 0x prefix)")
	}
	if _, err := hex.DecodeString(s); err != nil {
		return errors.New("execution_id must be 32 bytes hex-encoded (64 hex chars, no 0x prefix)")
	}
	return nil
}

// validateEnclavePublicKey requires hex-encoded bytes; length is intentionally not pinned
// because the encoding length depends on the enclave's key type and is not yet contracted
// in this package.
func validateEnclavePublicKey(s string) error {
	if _, err := hex.DecodeString(s); err != nil {
		return errors.New("enclave_public_key must be hex-encoded")
	}
	return nil
}

// validateEnclaveConfig rejects configs missing fields the canonical hash binds to.
// Signers must be non-empty (the relay needs to compare against the onchain DON
// membership). F must be > 0 (a DON with no fault tolerance is not a configuration
// the relay will trust). MasterPublicKey is checked for presence only; encoding is
// the enclave's contract. T is allowed to be zero in case some future enclave
// configurations carry it implicitly, but in practice the enclave will set it.
func validateEnclaveConfig(c EnclaveConfig) error {
	if len(c.Signers) == 0 {
		return errors.New("enclave_config.signers must be non-empty")
	}
	for i, s := range c.Signers {
		if len(s) == 0 {
			return fmt.Errorf("enclave_config.signers[%d] is empty", i)
		}
	}
	if c.F == 0 {
		return errors.New("enclave_config.f must be > 0")
	}
	if len(c.MasterPublicKey) == 0 {
		return errors.New("enclave_config.master_public_key must be non-empty")
	}
	return nil
}

// validateSecretIdentifiers rejects any entry with an empty Key or Namespace because the
// canonical hash binds to them and an empty value would produce a signature over an
// ambiguous selector.
func validateSecretIdentifiers(secrets []SecretIdentifier) error {
	for i, s := range secrets {
		if s.Key == "" {
			return fmt.Errorf("secrets[%d].key is required", i)
		}
		if s.Namespace == "" {
			return fmt.Errorf("secrets[%d].namespace is required", i)
		}
	}
	return nil
}

// Hash computes the canonical hash of the caller-provided request params and
// logical relay response body. Attestation is intentionally excluded. Returns an
// error if params fails Validate so a caller cannot accidentally produce a
// signature over an unbinding payload.
func (r *CapabilityResponseResult) Hash(params CapabilityRequestParams) ([32]byte, error) {
	if err := params.Validate(); err != nil {
		return [32]byte{}, fmt.Errorf("invalid capability request params: %w", err)
	}

	h := sha256.New()
	h.Write([]byte(teeattestation.DomainSeparator))
	h.Write([]byte("\nCapabilityResponseResult\n"))

	writeCapabilityRequestParams(h, params)
	writeString(h, r.Payload)
	writeString(h, r.Error)

	var result [32]byte
	h.Sum(result[:0])
	return result, nil
}

// RelayResponseSignature is a single relay-DON node signature over a relay
// response hash.
type RelayResponseSignature struct {
	Signer    []byte `json:"signer"`
	Signature []byte `json:"signature"`
}

// SignedSecretsResponseResult is one relay-DON node's signed secrets response:
// the logical result plus that single node's signature over the response hash.
// A node signs only its own response, so it carries exactly one signature; the
// gateway forwards a SignedSecretsResponseBundle of these without merging or
// trusting them, and the enclave verifies each against the relay-DON signer set.
type SignedSecretsResponseResult struct {
	Result SecretsResponseResult `json:"result"`
	// Deprecated: use Signature. A relay node signs only its own response, so this
	// array always carries exactly one entry. Retained for backward compatibility
	// while chainlink and confidential-compute migrate to the single-signature
	// field; it will be removed once nothing reads it.
	Signatures []RelayResponseSignature `json:"signatures,omitempty"`
	// Signature is this relay node's single signature over the response hash.
	Signature RelayResponseSignature `json:"signature"`
}

// SignedCapabilityResponseResult is one relay-DON node's signed capability
// response: the logical result plus that single node's signature over the
// response hash. See SignedSecretsResponseResult for the trust model.
type SignedCapabilityResponseResult struct {
	Result CapabilityResponseResult `json:"result"`
	// Deprecated: use Signature. A relay node signs only its own response, so this
	// array always carries exactly one entry. Retained for backward compatibility
	// while chainlink and confidential-compute migrate to the single-signature
	// field; it will be removed once nothing reads it.
	Signatures []RelayResponseSignature `json:"signatures,omitempty"`
	// Signature is this relay node's single signature over the response hash.
	Signature RelayResponseSignature `json:"signature"`
}

// SignedSecretsResponseBundle is the gateway's response to the enclave: the
// unverified set of per-node signed responses the gateway collected. The gateway
// makes no quorum decision and holds no signer keys; it is a dumb fan-in. The
// enclave groups the responses by their canonical hash, verifies each signature
// against the relay-DON signer set, and accepts the result backed by F+1 valid
// distinct signers. Invalid or foreign signatures are skipped, not fatal.
type SignedSecretsResponseBundle struct {
	Responses []SignedSecretsResponseResult `json:"responses"`
}

// SignedCapabilityResponseBundle is the gateway's response to the enclave for a
// capability execution. See SignedSecretsResponseBundle for the trust model.
type SignedCapabilityResponseBundle struct {
	Responses []SignedCapabilityResponseResult `json:"responses"`
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
	// Hash the config only when present so a nil EnclaveConfig produces the
	// same digest a sender that omits the field would, preserving
	// compatibility across the optional rollout.
	if params.EnclaveConfig != nil {
		writeEnclaveConfig(h, *params.EnclaveConfig)
	}
}

func writeCapabilityRequestParams(h hash.Hash, params CapabilityRequestParams) {
	writeString(h, params.WorkflowID)
	writeString(h, params.Owner)
	writeString(h, params.ExecutionID)
	writeString(h, params.OrgID)
	writeString(h, params.ReferenceID)
	writeString(h, params.CapabilityID)
	writeString(h, params.Payload)
	// Hash the config only when present so a nil EnclaveConfig produces the
	// same digest a sender that omits the field would, preserving
	// compatibility across the optional rollout.
	if params.EnclaveConfig != nil {
		writeEnclaveConfig(h, *params.EnclaveConfig)
	}
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

// writeEnclaveConfig binds every field of EnclaveConfig into the hash with
// canonical length prefixes. Signers are sorted so two logically-equivalent
// configs that differ only in Signer ordering produce the same hash; the
// relay-side comparison against onchain state is order-independent so the
// hashing must be too.
func writeEnclaveConfig(h hash.Hash, c EnclaveConfig) {
	signers := append([][]byte(nil), c.Signers...)
	sort.Slice(signers, func(i, j int) bool { return bytes.Compare(signers[i], signers[j]) < 0 })
	writeLengthPrefix(h, len(signers))
	for _, s := range signers {
		writeBytes(h, s)
	}
	writeBytes(h, c.MasterPublicKey)

	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], c.T)
	h.Write(buf[:])
	binary.BigEndian.PutUint32(buf[:], c.F)
	h.Write(buf[:])
}

func writeLengthPrefix(h hash.Hash, length int) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(length))
	h.Write(buf[:])
}
