package confidentialrelay

import "crypto/sha256"

// computeRequestDomainSeparator is vendored verbatim from confidential-compute
// types.DomainSeparator. It MUST stay byte-identical to the source, or
// ComputeRequest.Hash will not match the digest the Workflow DON nodes signed and
// F+1 verification at the relay DON will fail. chainlink-common cannot import
// confidential-compute, so the byte-for-byte conformance check lives in that repo
// (which can import this package).
const computeRequestDomainSeparator = "CONFIDENTIAL_COMPUTE_PAYLOAD"

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
// writeLengthPrefix. EncryptedDecryptionKeyShares is intentionally excluded,
// matching the source.
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
	writeString(h, cr.Version)

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
