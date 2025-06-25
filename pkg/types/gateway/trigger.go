package gateway

import "encoding/json"

const (
	HTTPTriggerStatusAccepted HTTPTriggerStatus = "ACCEPTED"
	MethodWorkflowExecute     string            = "workflows.execute"
)

type HTTPTriggerRequest struct {
	Workflow WorkflowSelector `json:"workflow"`
	Input    json.RawMessage  `json:"input"`
}

type WorkflowSelector struct {
	WorkflowOwner string `json:"workflowOwner,omitempty"`
	WorkflowName  string `json:"workflowName,omitempty"`
	WorkflowLabel string `json:"workflowLabel,omitempty"`
	WorkflowID    string `json:"workflowID,omitempty"`
}

type HTTPTriggerResponse struct {
	WorkflowID          string            `json:"workflow_id,omitempty"`
	WorkflowExecutionID string            `json:"workflow_execution_id,omitempty"`
	Status              HTTPTriggerStatus `json:"status,omitempty"`
}

type HTTPTriggerStatus string
