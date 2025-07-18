package gateway

const (
	MethodWorkflowPushAuthMetadata         = "push_auth_metadata"
	MethodWorkflowPullAuthMetadata         = "pull_auth_metadata"
	KeyTypeECDSA                   KeyType = "ecdsa"
)

type KeyType string

// WorkflowMetadata represents the workflow metadata for HTTP triggers including auth data.
// This type is used for communication between the gateway handler in the gateway node and
// the gateway connector handler in the workflow node.
type WorkflowMetadata struct {
	WorkflowSelector WorkflowSelector
	AuthorizedKeys   []AuthorizedKey
}

type AuthorizedKey struct {
	KeyType   KeyType
	PublicKey string
}
