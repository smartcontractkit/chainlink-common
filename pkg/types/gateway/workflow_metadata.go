package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	// Note: any addition to this list must be reflected in core's HandlerTypeForMethod function.
	MethodPushWorkflowMetadata         = "push_workflow_metadata"
	MethodPullWorkflowMetadata         = "pull_workflow_metadata"
	KeyTypeECDSA               KeyType = "ecdsa"
)

type KeyType string

// WorkflowMetadata represents the workflow metadata for HTTP triggers including auth data.
// This type is used for communication between the gateway handler in the gateway node and
// the gateway connector handler in the workflow node.
type WorkflowMetadata struct {
	WorkflowSelector WorkflowSelector
	AuthorizedKeys   []AuthorizedKey
}

func (wm *WorkflowMetadata) Digest() (string, error) {
	data, err := json.Marshal(wm)
	if err != nil {
		return "", err
	}
	hasher := sha256.New()
	hasher.Write(data)
	digestBytes := hasher.Sum(nil)

	return hex.EncodeToString(digestBytes), nil
}

type AuthorizedKey struct {
	KeyType   KeyType `json:"keyType"`
	PublicKey string  `json:"publicKey"`
}

// MarshalJSON implements custom JSON marshalling to ensure alphabetical order of keys for AuthorizedKey,
// and only includes non-empty fields.
func (r AuthorizedKey) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	if r.KeyType != "" {
		m["keyType"] = r.KeyType
	}
	if r.PublicKey != "" {
		m["publicKey"] = r.PublicKey
	}
	return marshalWithSortedKeys(m)
}
