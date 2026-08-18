package privateregistry

import "context"

// WorkflowDeploymentAction defines operations for managing workflows in a workflow source.
// This interface is implemented by both the mock server (for testing) and the actual
// private workflow registry server (when built).
type WorkflowDeploymentAction interface {
	// AddWorkflow registers a new workflow with the source
	AddWorkflow(ctx context.Context, workflow *WorkflowRegistration) error

	// UpdateWorkflow updates the workflow's status configuration
	UpdateWorkflow(ctx context.Context, workflowID [32]byte, config *WorkflowStatusConfig) error

	// DeleteWorkflow removes the workflow from the source
	DeleteWorkflow(ctx context.Context, workflowID [32]byte) error
}

// WorkflowStatusConfig contains the desired state for a workflow's status
type WorkflowStatusConfig struct {
	// Paused indicates whether the workflow should be paused (true) or active (false)
	Paused bool
}

// WorkflowRegistration contains the data needed to register a workflow
type WorkflowRegistration struct {
	WorkflowID   [32]byte
	Owner        []byte
	WorkflowName string
	BinaryURL    string
	ConfigURL    string
	DonFamily    string
	Tag          string
	Attributes   []byte
}
