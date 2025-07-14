package gateway

const (
	MethodWorkflowPushAuthMetadata         = "push_auth_metadata"
	MethodWorkflowPullAuthMetadata         = "pull_auth_metadata"
	KeyTypeECDSA                   KeyType = "ecdsa"
)

type KeyType string

// WorkflowAuthMetadata represents the metadata for workflow authorization
// This type is used for communication between the gateway handler in the gateway node and
// the gateway connector handler in the workflow node.
type WorkflowAuthMetadata struct {
	WorkflowID     string
	AuthorizedKeys []AuthorizedKey
}

type AuthorizedKey struct {
	KeyType   KeyType
	PublicKey string
}
