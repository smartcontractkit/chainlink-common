package pb_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	testWorkflowID  = "test-id-1"
	testConfigKey   = "test-key"
	testConfigValue = "test-value"
	testInputsKey   = "input-key"
	testInputsValue = "input-value"
	testError       = "test-error"
	anyReferenceID  = "anything"
)

func TestCapabilityRequestFromProto(t *testing.T) {
	_, err := pb.CapabilityRequestFromProto(nil)
	assert.ErrorContains(t, err, "could not convert nil proto")

	pr := pb.CapabilityRequest{
		Metadata: nil,
		Inputs:   values.ProtoMap(values.EmptyMap()),
		Config:   values.ProtoMap(values.EmptyMap()),
	}
	_, err = pb.CapabilityRequestFromProto(&pr)
	assert.ErrorContains(t, err, "could not convert nil metadata")

	inputs, err := values.NewMap(map[string]any{
		"hello": "world",
	})
	require.NoError(t, err)

	config, err := values.NewMap(map[string]any{
		"aConfigVersion": true,
	})
	require.NoError(t, err)
	pr = pb.CapabilityRequest{
		Metadata: &pb.RequestMetadata{
			WorkflowId: "<workflow-id>",
		},
		Inputs: values.ProtoMap(inputs),
		Config: values.ProtoMap(config),
	}
	_, err = pb.CapabilityRequestFromProto(&pr)
	require.NoError(t, err)

	pr.Metadata.ReferenceId = anyReferenceID
	_, err = pb.CapabilityRequestFromProto(&pr)
	require.NoError(t, err)
}

func TestCapabilityResponseFromProto(t *testing.T) {
	_, err := pb.CapabilityResponseFromProto(nil)
	assert.ErrorContains(t, err, "could not convert nil proto")

	pr := pb.CapabilityResponse{
		Value: values.ProtoMap(values.EmptyMap()),
		Error: "error: bang!",
	}
	_, err = pb.CapabilityResponseFromProto(&pr)
	require.NoError(t, err)
}

func TestMarshalUnmarshalRequest(t *testing.T) {
	req := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID:               "test-workflow-id",
			WorkflowExecutionID:      testWorkflowID,
			WorkflowOwner:            "0xaa",
			WorkflowName:             "test-workflow-name",
			WorkflowDonID:            1,
			WorkflowDonConfigVersion: 1,
			ReferenceID:              anyReferenceID,
		},
		Config: &values.Map{Underlying: map[string]values.Value{
			testConfigKey: &values.String{Underlying: testConfigValue},
		}},
		Inputs: &values.Map{Underlying: map[string]values.Value{
			testInputsKey: &values.String{Underlying: testInputsValue},
		}},
	}
	raw, err := pb.MarshalCapabilityRequest(req)
	require.NoError(t, err)

	unmarshaled, err := pb.UnmarshalCapabilityRequest(raw)
	require.NoError(t, err)

	require.Equal(t, req, unmarshaled)

	req.Metadata.ReferenceID = anyReferenceID
	raw, err = pb.MarshalCapabilityRequest(req)
	require.NoError(t, err)

	unmarshaled, err = pb.UnmarshalCapabilityRequest(raw)
	require.NoError(t, err)

	require.Equal(t, req, unmarshaled)
}

func TestMarshalUnmarshalResponse(t *testing.T) {
	v, err := values.NewMap(map[string]any{"hello": "world"})
	require.NoError(t, err)
	resp := capabilities.CapabilityResponse{
		Value: v,
		Err:   errors.New(testError),
	}
	raw, err := pb.MarshalCapabilityResponse(resp)
	require.NoError(t, err)

	unmarshaled, err := pb.UnmarshalCapabilityResponse(raw)
	require.NoError(t, err)

	require.Equal(t, resp, unmarshaled)
}
