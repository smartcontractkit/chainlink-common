package confidentialrelay

import (
	"crypto/sha256"

	"github.com/smartcontractkit/libocr/ragep2p/peeridhelper"
)

// computeRequestDomainSeparator is vendored verbatim from confidential-compute
// types.DomainSeparator. It MUST stay byte-identical to the source, or
// ComputeRequest.Hash will not match the digest the Workflow DON nodes signed and
// F+1 verification at the relay DON will fail. chainlink-common cannot import
// confidential-compute, so the byte-for-byte conformance check lives in that repo
// (which can import this package).
const computeRequestDomainSeparator = "CONFIDENTIAL_COMPUTE_PAYLOAD"

// signedComputeRequestSignaturePrefix is vendored verbatim from confidential-compute
// util.GetConfidentialComputePayloadPrefix(). Each Workflow DON node signs the peerid
// domain-separated payload over ComputeRequest.Hash() using this prefix; the relay DON
// reconstructs the same payload (via SignedComputeRequestSignaturePayload) to verify the
// F+1 signatures against the Workflow DON signer set. Note the trailing underscore: this
// is the signature prefix, distinct from computeRequestDomainSeparator (the hash prefix).
const signedComputeRequestSignaturePrefix = "CONFIDENTIAL_COMPUTE_PAYLOAD_"

// computeRequestLegacyVersion is vendored from confidential-compute
// types.ServiceConfidentialComputeVersionLegacy. Hash includes the Version field only
// when it equals this value, matching the source (confidential-compute is migrating
// Version out of the hash for newer versions). It MUST stay in sync with the source, or
// ComputeRequest.Hash will diverge from the digest the Workflow DON nodes signed once the
// enclave moves past the legacy version.
const computeRequestLegacyVersion = "0.0.6"

// SignedComputeRequestSignaturePayload reconstructs the exact payload a Workflow DON node
// signed over a ComputeRequest hash, so the relay DON can verify the signature with the
// node's public key.
func SignedComputeRequestSignaturePayload(computeRequestHash [32]byte) []byte {
	return peeridhelper.MakePeerIDSignatureDomainSeparatedPayload(signedComputeRequestSignaturePrefix, computeRequestHash[:])
}

// ComputeRequest is vendored from confidential-compute types.ComputeRequest. The
// relay DON cannot import confidential-compute (the dependency runs the other way),
// so the type and its canonical Hash are copied here. The enclave forwards the
// Workflow-DON-signed compute request to the relay, which reconstructs the hash and
// verifies the F+1 signatures over it.
//
// PublicData carries the marshaled WorkflowExecution (owner, orgid, workflowID,
// executionID); the relay unmarshals it via chainlink-protos to recover the
// authorized identity.
type ComputeRequest struct {
	RequestID                    [32]byte   `json:"requestID"`
	PublicData                   []byte     `json:"publicData"`
	Ciphertexts                  [][]byte   `json:"ciphertexts"`
	CiphertextNames              []string   `json:"CiphertextNames"`
	EncryptedDecryptionKeyShares [][][]byte `json:"encryptedDecryptionKeyShares"`
	EnclaveEphemeralPublicKey    []byte     `json:"enclaveEphemeralPublicKey"`
	MasterPublicKey              []byte     `json:"masterPublicKey"`
	AppID                        string     `json:"appID"`
	Version                      string     `json:"version"`
}

// Hash mirrors confidential-compute types.ComputeRequest.Hash byte-for-byte. It
// reuses this package's length-prefix helpers (writeBytes/writeString/
// writeLengthPrefix), which are identical to the source's writeWithLength/
// writeLengthPrefix. EncryptedDecryptionKeyShares is intentionally excluded, and
// Version is included only for the legacy version, both matching the source.
func (cr ComputeRequest) Hash() [32]byte {
	h := sha256.New()

	h.Write([]byte(computeRequestDomainSeparator))
	h.Write([]byte("\nComputeRequest\n"))

	h.Write(cr.RequestID[:])

	writeBytes(h, cr.PublicData)

	writeLengthPrefix(h, len(cr.CiphertextNames))
	for _, name := range cr.CiphertextNames {
		writeString(h, name)
	}

	writeLengthPrefix(h, len(cr.Ciphertexts))
	for _, ciphertext := range cr.Ciphertexts {
		writeBytes(h, ciphertext)
	}

	writeBytes(h, cr.EnclaveEphemeralPublicKey)
	writeBytes(h, cr.MasterPublicKey)

	writeString(h, cr.AppID)
	// Version is included in the hash only for the legacy version, matching
	// confidential-compute (which is migrating Version out of the hash).
	if cr.Version == computeRequestLegacyVersion {
		writeString(h, cr.Version)
	}

	var result [32]byte
	h.Sum(result[:0])
	return result
}

// SignedComputeRequest is vendored from confidential-compute
// types.SignedComputeRequest: a ComputeRequest plus one Workflow DON node's
// signature over ComputeRequest.Hash. The enclave forwards the F+1 signed requests
// to the relay DON as the authorization for a secrets request.
type SignedComputeRequest struct {
	ComputeRequest
	Signature   []byte            `json:"signature"`
	PerNodeData map[string]string `json:"perNodeData,omitempty"`
}
