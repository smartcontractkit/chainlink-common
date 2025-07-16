package gateway

import "encoding/json"

// HTTPTriggerStatus represents the status of an HTTP trigger request.
type HTTPTriggerStatus string

const (
	// HTTPTriggerStatusAccepted indicates the trigger request was accepted.
	HTTPTriggerStatusAccepted HTTPTriggerStatus = "ACCEPTED"
	// MethodWorkflowExecute is the method name for executing workflows.
	MethodWorkflowExecute string = "workflows.execute"
)

// HTTPTriggerRequest represents a request to trigger a workflow via HTTP.
type HTTPTriggerRequest struct {
	Workflow WorkflowSelector `json:"workflow"` // Selector for the workflow to execute.
	Input    json.RawMessage  `json:"input"`    // Input parameters for the workflow.
}

// WorkflowSelector specifies how to identify a workflow.
type WorkflowSelector struct {
	WorkflowOwner string `json:"workflowOwner,omitempty"` // Owner of the workflow.
	WorkflowName  string `json:"workflowName,omitempty"`  // Name of the workflow.
	WorkflowTag   string `json:"workflowTag,omitempty"`   // Tag for the workflow.
	WorkflowID    string `json:"workflowID,omitempty"`    // Unique ID of the workflow.
}

// HTTPTriggerResponse represents the response to an HTTP trigger request.
type HTTPTriggerResponse struct {
	WorkflowID          string            `json:"workflow_id,omitempty"`           // ID of the triggered workflow.
	WorkflowExecutionID string            `json:"workflow_execution_id,omitempty"` // ID of the workflow execution.
	Status              HTTPTriggerStatus `json:"status,omitempty"`                // Status of the trigger request.
}
