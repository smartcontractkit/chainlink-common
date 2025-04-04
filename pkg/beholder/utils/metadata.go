package utils

import (
	"encoding/hex"

	"go.opentelemetry.io/otel/attribute"
)

const (
	// WorkflowExecutionIDShortLen is the length of the short version of the WorkflowExecutionId (label)
	WorkflowExecutionIDShortLen = 3 // first 3 characters, 16^3 = 4.096 possibilities (mid-high cardinality)
)

// TODO: Refactor as a proto referenced from the other proto files (telemetry messages)
type ExecutionMetadata struct {
	// Execution Context - Source
	SourceId string
	// Execution Context - Chain
	ChainFamilyName string
	ChainId         string
	NetworkName     string
	NetworkNameFull string
	// Execution Context - Workflow (capabilities.RequestMetadata)
	WorkflowId               string
	WorkflowOwner            string
	WorkflowExecutionId      string
	WorkflowName             string
	WorkflowDonId            uint32
	WorkflowDonConfigVersion uint32
	ReferenceId              string
	// Execution Context - Capability
	CapabilityType           string
	CapabilityId             string
	CapabilityTimestampStart uint32
	CapabilityTimestampEmit  uint32
}

// Attributes returns common attributes used for metrics
func (m ExecutionMetadata) Attributes() []attribute.KeyValue {
	// Decode workflow name attribute for output
	workflowName := m.decodeWorkflowName()

	return []attribute.KeyValue{
		// Execution Context - Source
		attribute.String("source_id", ValOrUnknown(m.SourceId)),
		// Execution Context - Chain
		attribute.String("chain_family_name", ValOrUnknown(m.ChainFamilyName)),
		attribute.String("chain_id", ValOrUnknown(m.ChainId)),
		attribute.String("network_name", ValOrUnknown(m.NetworkName)),
		attribute.String("network_name_full", ValOrUnknown(m.NetworkNameFull)),
		// Execution Context - Workflow (capabilities.RequestMetadata)
		attribute.String("workflow_id", ValOrUnknown(m.WorkflowId)),
		attribute.String("workflow_owner", ValOrUnknown(m.WorkflowOwner)),
		// Notice: We lower the cardinality on the WorkflowExecutionId so it can be used by metrics
		// This label has good chances to be unique per workflow, in a reasonable bounded time window
		// TODO: enable this when sufficiently tested (PromQL queries like alerts might need to change if this is used)
		// attribute.String("workflow_execution_id_short", ValShortOrUnknown(m.WorkflowExecutionId, WorkflowExecutionIDShortLen)),
		attribute.String("workflow_name", ValOrUnknown(workflowName)),
		attribute.Int64("workflow_don_id", int64(m.WorkflowDonId)),
		attribute.Int64("workflow_don_config_version", int64(m.WorkflowDonConfigVersion)),
		attribute.String("reference_id", ValOrUnknown(m.ReferenceId)),
		// Execution Context - Capability
		attribute.String("capability_type", ValOrUnknown(m.CapabilityType)),
		attribute.String("capability_id", ValOrUnknown(m.CapabilityId)),
		// Notice: we don't include the timestamps here (high cardinality)
	}
}

// decodeWorkflowName decodes the workflow name from hex string to raw string (underlying, output)
func (m ExecutionMetadata) decodeWorkflowName() string {
	bytes, err := hex.DecodeString(m.WorkflowName)
	if err != nil {
		// This should never happen
		bytes = []byte("unknown-decode-error")
	}
	return string(bytes)
}

// This is needed to avoid issues during exporting OTel metrics to Prometheus
// For more details see https://smartcontract-it.atlassian.net/browse/INFOPLAT-1349
// ValOrUnknown returns the value if it is not empty, otherwise it returns "unknown"
func ValOrUnknown(val string) string {
	if val == "" {
		return "unknown"
	}
	return val
}

// ValShortOrUnknown returns the short len value if not empty or available, otherwise it returns "unknown"
func ValShortOrUnknown(val string, _len int) string {
	if val == "" || _len <= 0 {
		return "unknown"
	}
	if _len > len(val) {
		return val
	}
	return val[:_len]
}
