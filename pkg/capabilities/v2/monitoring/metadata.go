package monitoring

import (
	"encoding/hex"

	"go.opentelemetry.io/otel/attribute"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

// ValOrUnknown returns the value if it is not empty, otherwise it returns "unknown".
// This avoids issues during exporting OTel metrics to Prometheus.
func ValOrUnknown(val string) string {
	if val == "" {
		return "unknown"
	}
	return val
}

// RequestID combines workflow execution ID and reference ID for log correlation.
func RequestID(workflowExecutionID, reference string) string {
	return workflowExecutionID + ":" + reference
}

// RequestLogKVs returns per-request structured log fields added by generated server code.
func RequestLogKVs(metadata capabilities.RequestMetadata) []any {
	return []any{
		"workflowID", metadata.WorkflowID,
		"workflowOwner", metadata.WorkflowOwner,
		"workflowExecutionID", metadata.WorkflowExecutionID,
		"workflowName", workflowName(metadata),
		"workflowDonID", metadata.WorkflowDonID,
		"workflowDonConfigVersion", metadata.WorkflowDonConfigVersion,
		"referenceID", metadata.ReferenceID,
		"requestID", RequestID(metadata.WorkflowExecutionID, metadata.ReferenceID),
	}
}

func workflowName(metadata capabilities.RequestMetadata) string {
	if metadata.DecodedWorkflowName != "" {
		return metadata.DecodedWorkflowName
	}
	bytes, err := hex.DecodeString(metadata.WorkflowName)
	if err != nil {
		return ValOrUnknown(metadata.WorkflowName)
	}
	return string(bytes)
}

// ActionMetricAttributes combines capability, request, and optional per-input metric labels.
func ActionMetricAttributes(method string, metadata capabilities.RequestMetadata, capAttrsFn func() []attribute.KeyValue, extra []attribute.KeyValue) []attribute.KeyValue {
	var capAttrs []attribute.KeyValue
	if capAttrsFn != nil {
		capAttrs = capAttrsFn()
	}
	attrs := make([]attribute.KeyValue, 0, len(capAttrs)+2+len(extra))
	attrs = append(attrs, capAttrs...)
	attrs = append(attrs,
		attribute.String(LabelMethod, ValOrUnknown(method)),
		attribute.Int64(LabelWorkflowDonID, int64(metadata.WorkflowDonID)),
	)
	return append(attrs, extra...)
}
