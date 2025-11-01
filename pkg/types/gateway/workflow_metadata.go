package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

const (
	// Note: any addition to this list must be reflected in the handler's Methods() function.
	MethodPushWorkflowMetadata         = "push_workflow_metadata"
	MethodPullWorkflowMetadata         = "pull_workflow_metadata"
	KeyTypeECDSAEVM            KeyType = "ecdsa_evm"
)

type KeyType string

// WorkflowMetadata represents the workflow metadata for HTTP triggers including auth data.
// This type is used for communication between the gateway handler in the gateway node and
// the gateway connector handler in the workflow node.
type WorkflowMetadata struct {
	WorkflowSelector WorkflowSelector
	AuthorizedKeys   []AuthorizedKey
}

// Digest returns a digest of the workflow metadata. This is used for aggregating metadata
// across multiple nodes. The digest is a SHA256 hash of the canonical JSON representation,
// ensuring deterministic output regardless of the order in which authorized keys are reported.
func (wm *WorkflowMetadata) Digest() (string, error) {
	JSONBytes, err := jsonv2.Marshal(wm, jsonv2.Deterministic(true))
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	canonicalJSONBytes := jsontext.Value(JSONBytes)
	err = canonicalJSONBytes.Canonicalize()
	if err != nil {
		return "", fmt.Errorf("error canonicalizing JSON: %w", err)
	}

	hasher := sha256.New()
	if _, err := hasher.Write(canonicalJSONBytes); err != nil {
		return "", fmt.Errorf("error writing to hasher: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

type AuthorizedKey struct {
	KeyType   KeyType `json:"keyType,omitempty"`
	PublicKey string  `json:"publicKey,omitempty"`
}
