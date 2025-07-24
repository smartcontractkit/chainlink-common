package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

const (
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
	wmCopy := *wm
	sortedKeys := make([]AuthorizedKey, len(wm.AuthorizedKeys))
	copy(sortedKeys, wm.AuthorizedKeys)
	sort.Slice(sortedKeys, func(i, j int) bool {
		if sortedKeys[i].KeyType != sortedKeys[j].KeyType {
			return string(sortedKeys[i].KeyType) < string(sortedKeys[j].KeyType)
		}
		return sortedKeys[i].PublicKey < sortedKeys[j].PublicKey
	})
	wmCopy.AuthorizedKeys = sortedKeys

	data, err := json.Marshal(wmCopy)
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
