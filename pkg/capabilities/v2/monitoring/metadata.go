package monitoring

import (
	"encoding/hex"
	"fmt"

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
func ActionMetricAttributes(method string, metadata capabilities.RequestMetadata, capAttrsFn func() []attribute.KeyValue, extraKVs []any) []attribute.KeyValue {
	var capAttrs []attribute.KeyValue
	if capAttrsFn != nil {
		capAttrs = capAttrsFn()
	}
	attrs := make([]attribute.KeyValue, 0, len(capAttrs)+2+len(extraKVs)/2)
	attrs = append(attrs, capAttrs...)
	attrs = append(attrs,
		attribute.String(LabelMethod, ValOrUnknown(method)),
		attribute.Int64(LabelWorkflowDonID, int64(metadata.WorkflowDonID)),
	)
	return append(attrs, KVsToAttributes(extraKVs)...)
}

// KVsToAttributes converts alternating key/value pairs into OTel attributes.
func KVsToAttributes(kvs []any) []attribute.KeyValue {
	if len(kvs) == 0 {
		return nil
	}
	attrs := make([]attribute.KeyValue, 0, len(kvs)/2)
	for i := 0; i+1 < len(kvs); i += 2 {
		key, ok := kvs[i].(string)
		if !ok {
			continue
		}
		attrs = append(attrs, attribute.String(key, fmtAny(kvs[i+1])))
	}
	return attrs
}

func fmtAny(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmtStringer:
		return t.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

type fmtStringer interface {
	String() string
}
