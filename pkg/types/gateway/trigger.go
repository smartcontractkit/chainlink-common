package gateway

import (
	"bytes"
	"encoding/json"
	"sort"
)

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
	Input    json.RawMessage  `json:"input"`    // Input parameters for the workflow.
	Key      AuthorizedKey    `json:"key"`      // Signing key for the request
	Workflow WorkflowSelector `json:"workflow"` // Selector for the workflow to execute.
}

// WorkflowSelector specifies how to identify a workflow.
type WorkflowSelector struct {
	WorkflowID    string `json:"workflowID,omitempty"`    // Unique ID of the workflow.
	WorkflowName  string `json:"workflowName,omitempty"`  // Name of the workflow.
	WorkflowOwner string `json:"workflowOwner,omitempty"` // Owner of the workflow.
	WorkflowTag   string `json:"workflowTag,omitempty"`   // Tag for the workflow.
}

// HTTPTriggerResponse represents the response to an HTTP trigger request.
type HTTPTriggerResponse struct {
	WorkflowID          string            `json:"workflow_id,omitempty"`           // ID of the triggered workflow.
	WorkflowExecutionID string            `json:"workflow_execution_id,omitempty"` // ID of the workflow execution.
	Status              HTTPTriggerStatus `json:"status,omitempty"`                // Status of the trigger request.
}

// MarshalJSON implements custom JSON marshalling to ensure alphabetical order of keys for WorkflowSelector,
// and only includes non-empty fields.
func (ws WorkflowSelector) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	if ws.WorkflowID != "" {
		m["workflowID"] = ws.WorkflowID
	}
	if ws.WorkflowName != "" {
		m["workflowName"] = ws.WorkflowName
	}
	if ws.WorkflowOwner != "" {
		m["workflowOwner"] = ws.WorkflowOwner
	}
	if ws.WorkflowTag != "" {
		m["workflowTag"] = ws.WorkflowTag
	}
	return marshalWithSortedKeys(m)
}

// MarshalJSON implements custom JSON marshalling to ensure deterministic output
// with sorted keys at all levels for map[string]interface{}, including nested objects in the Input field.
func (r HTTPTriggerRequest) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})
	if len(r.Input) > 0 {
		var inputData interface{}
		if err := json.Unmarshal(r.Input, &inputData); err != nil {
			return nil, err
		}
		result["input"] = inputData
	}
	result["key"] = r.Key
	result["workflow"] = r.Workflow
	return marshalWithSortedKeys(result)
}

func marshalWithSortedKeys(data map[string]interface{}) ([]byte, error) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}

		// Marshal the key
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		// Marshal the value
		valueBytes, err := marshalJSONValue(data[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// marshalJSONValue marshals a JSON value, handling maps and arrays with sorted keys
func marshalJSONValue(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		return marshalWithSortedKeys(v)
	case []interface{}:
		return marshalArrayWithSortedKeys(v)
	default:
		return json.Marshal(v)
	}
}

// marshalArrayWithSortedKeys marshals an array, ensuring any nested maps have sorted keys
func marshalArrayWithSortedKeys(data []interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')

	for i, item := range data {
		if i > 0 {
			buf.WriteByte(',')
		}

		// Recursively marshal each item, handling nested maps/arrays
		itemBytes, err := marshalJSONValue(item)
		if err != nil {
			return nil, err
		}
		buf.Write(itemBytes)
	}

	buf.WriteByte(']')
	return buf.Bytes(), nil
}
