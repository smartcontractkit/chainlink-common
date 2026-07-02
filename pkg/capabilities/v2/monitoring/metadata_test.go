package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

func TestRequestLogKVs(t *testing.T) {
	t.Parallel()

	kvs := RequestLogKVs(capabilities.RequestMetadata{
		WorkflowID:               "wf-1",
		WorkflowOwner:            "owner-1",
		WorkflowExecutionID:      "exec-1",
		DecodedWorkflowName:      "my-workflow",
		WorkflowDonID:            3,
		WorkflowDonConfigVersion: 2,
		ReferenceID:              "ref-1",
	})
	require.Len(t, kvs, 16)

	fields := make(map[string]any, len(kvs)/2)
	for i := 0; i+1 < len(kvs); i += 2 {
		fields[kvs[i].(string)] = kvs[i+1]
	}
	assert.Equal(t, "wf-1", fields["workflowID"])
	assert.Equal(t, "owner-1", fields["workflowOwner"])
	assert.Equal(t, "exec-1", fields["workflowExecutionID"])
	assert.Equal(t, "my-workflow", fields["workflowName"])
	assert.Equal(t, uint32(3), fields["workflowDonID"])
	assert.Equal(t, uint32(2), fields["workflowDonConfigVersion"])
	assert.Equal(t, "ref-1", fields["referenceID"])
	assert.Equal(t, "exec-1:ref-1", fields["requestID"])
}

func TestActionMetricAttributes(t *testing.T) {
	t.Parallel()

	capAttrsFn := func() []attribute.KeyValue {
		return []attribute.KeyValue{attribute.String(LabelCapabilityID, "cap-1")}
	}
	attrs := ActionMetricAttributes("WriteReport", capabilities.RequestMetadata{WorkflowDonID: 3}, capAttrsFn, []attribute.KeyValue{
		attribute.String("region", "us-east"),
	})
	require.Len(t, attrs, 4)
	assert.Equal(t, attribute.String(LabelCapabilityID, "cap-1"), attrs[0])
	assert.Equal(t, attribute.String(LabelMethod, "WriteReport"), attrs[1])
	assert.Equal(t, attribute.Int64(LabelWorkflowDonID, 3), attrs[2])
	assert.Equal(t, attribute.String("region", "us-east"), attrs[3])
}

func TestActionMetricAttributesNilCapabilityFn(t *testing.T) {
	t.Parallel()

	attrs := ActionMetricAttributes("WriteReport", capabilities.RequestMetadata{WorkflowDonID: 3}, nil, nil)
	require.Len(t, attrs, 2)
	assert.Equal(t, attribute.String(LabelMethod, "WriteReport"), attrs[0])
	assert.Equal(t, attribute.Int64(LabelWorkflowDonID, 3), attrs[1])
}
