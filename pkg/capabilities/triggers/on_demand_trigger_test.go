package triggers

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const testID = "test-id-1"

func TestOnDemand(t *testing.T) {
	tr := NewOnDemand()
	ctx := tests.Context(t)

	callback := make(chan capabilities.CapabilityResponse, 10)

	req := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowExecutionID: testID,
		},
	}
	err := tr.RegisterTrigger(ctx, callback, req)
	require.NoError(t, err)

	er := capabilities.CapabilityResponse{
		Value: &values.String{Underlying: testID},
	}

	err = tr.FanOutEvent(ctx, er)
	require.NoError(t, err)

	assert.Len(t, callback, 1)
	assert.Equal(t, er, <-callback)
}

func TestOnDemand_ChannelDoesntExist(t *testing.T) {
	tr := NewOnDemand()
	ctx := tests.Context(t)

	er := capabilities.CapabilityResponse{
		Value: &values.String{Underlying: testID},
	}
	err := tr.SendEvent(ctx, testID, er)
	assert.ErrorContains(t, err, "no registration")
}

func TestOnDemand_(t *testing.T) {
	tr := NewOnDemand()
	ctx := tests.Context(t)

	req := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID: "hello",
		},
	}
	callback := make(chan capabilities.CapabilityResponse, 10)

	err := tr.RegisterTrigger(ctx, callback, req)
	require.NoError(t, err)

	er := capabilities.CapabilityResponse{
		Value: &values.String{Underlying: testID},
	}
	err = tr.SendEvent(ctx, "hello", er)
	require.NoError(t, err)

	assert.Len(t, callback, 1)
	assert.Equal(t, er, <-callback)
}

func TestOnDemandTrigger_GenerateConfigSchema(t *testing.T) {
	ts := NewOnDemand()
	schema := ts.GetRequestConfigJSONSchema()
	require.NotNil(t, schema)
	require.NoError(t, schema.Err)
	schemaStr := schema.Value.(*values.String)

	var shouldUpdate = false
	if shouldUpdate {
		err := os.WriteFile("./testdata/fixtures/ondemand/config_schema.json", []byte(schemaStr.Underlying), 0600)
		require.NoError(t, err)
	}

	fixture, err := os.ReadFile("./testdata/fixtures/ondemand/config_schema.json")
	require.NoError(t, err)
	
	if diff := cmp.Diff(fixture, []byte(schemaStr.Underlying), transformJSON); diff != "" {
		t.Errorf("TestOnDemandTrigger_GenerateConfigSchema() mismatch (-want +got):\n%s", diff)
		t.FailNow()
	}
}
