package pb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	testWorkflowID    = "test-id-1"
	testConfigKey     = "test-key"
	testConfigValue   = "test-value"
	testInputsKey     = "input-key"
	testInputsValue   = "input-value"
	testError         = "test-error"
	anyReferenceID    = "anything"
	testWorkflowOwner = "testowner"
	testWorkflowName  = "00112233445566778899"
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

	anyMsg := &anypb.Any{
		TypeUrl: "example.com/type",
		Value:   []byte("test-bytes"),
	}

	pr = pb.CapabilityRequest{
		Metadata: &pb.RequestMetadata{
			WorkflowId: "<workflow-id>",
		},
		Inputs:        values.ProtoMap(inputs),
		Config:        values.ProtoMap(config),
		ConfigPayload: anyMsg,
	}
	out, err := pb.CapabilityRequestFromProto(&pr)
	require.NoError(t, err)
	require.True(t, proto.Equal(anyMsg, out.ConfigPayload))

	pr.Metadata.ReferenceId = anyReferenceID
	out, err = pb.CapabilityRequestFromProto(&pr)
	require.NoError(t, err)
	require.Equal(t, anyReferenceID, out.Metadata.ReferenceID)
}

func TestCapabilityResponseFromProto(t *testing.T) {
	_, err := pb.CapabilityResponseFromProto(nil)
	assert.ErrorContains(t, err, "could not convert nil proto")

	pr := pb.CapabilityResponse{
		Value: values.ProtoMap(values.EmptyMap()),
	}
	resp, err := pb.CapabilityResponseFromProto(&pr)
	require.NoError(t, err)
	assert.Equal(t, capabilities.CapabilityResponse{Value: values.EmptyMap()}, resp)
}

func TestMarshalUnmarshalRequest(t *testing.T) {
	req := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID:               "test-workflow-id",
			WorkflowExecutionID:      testWorkflowID,
			WorkflowOwner:            "0xaa",
			WorkflowName:             testWorkflowName,
			WorkflowDonID:            1,
			WorkflowDonConfigVersion: 1,
			ReferenceID:              anyReferenceID,
			DecodedWorkflowName:      "test-workflow-name",
			SpendLimits: []capabilities.SpendLimit{
				{SpendType: "COMPUTE", Limit: "1000"},
				{SpendType: "GAS_12345", Limit: "1000000"},
			},
			WorkflowTag: "test-workflow-tag",
		},
		Config: &values.Map{Underlying: map[string]values.Value{
			testConfigKey: &values.String{Underlying: testConfigValue},
		}},
		Inputs: &values.Map{Underlying: map[string]values.Value{
			testInputsKey: &values.String{Underlying: testInputsValue},
		}},
		ConfigPayload: &anypb.Any{
			TypeUrl: "example.com/type",
			Value:   []byte("any-bytes"),
		},
		Method:       "call-it",
		CapabilityId: "test-capability-id",
	}

	raw, err := pb.MarshalCapabilityRequest(req)
	require.NoError(t, err)

	unmarshaled, err := pb.UnmarshalCapabilityRequest(raw)
	require.NoError(t, err)

	require.EqualValues(t, req.Metadata, unmarshaled.Metadata)
	require.EqualValues(t, req.Config, unmarshaled.Config)
	require.EqualValues(t, req.Inputs, unmarshaled.Inputs)
	require.True(t, proto.Equal(req.ConfigPayload, unmarshaled.ConfigPayload))

	req.Metadata.ReferenceID = anyReferenceID
	raw, err = pb.MarshalCapabilityRequest(req)
	require.NoError(t, err)

	unmarshaled, err = pb.UnmarshalCapabilityRequest(raw)
	require.NoError(t, err)

	require.EqualValues(t, req.Metadata, unmarshaled.Metadata)
	require.EqualValues(t, req.Config, unmarshaled.Config)
	require.EqualValues(t, req.Inputs, unmarshaled.Inputs)
	require.True(t, proto.Equal(req.ConfigPayload, unmarshaled.ConfigPayload))
}

func TestTriggerResponseFromProto(t *testing.T) {
	t.Run("with event outputs", func(t *testing.T) {
		outMap := &pb.TriggerEvent{
			Id:          "id",
			TriggerType: "type",
			Outputs: values.ProtoMap(&values.Map{
				Underlying: map[string]values.Value{
					"a": &values.String{Underlying: "b"},
				},
			}),
		}
		protoResp := &pb.TriggerResponse{
			Event: outMap,
			Error: "",
		}
		resp, err := pb.TriggerResponseFromProto(protoResp)
		require.NoError(t, err)
		assert.Nil(t, resp.Err)
		assert.Equal(t, "id", resp.Event.ID)
		assert.Equal(t, "type", resp.Event.TriggerType)
		assert.NotNil(t, resp.Event.Outputs)
		assert.Equal(t, "b", resp.Event.Outputs.Underlying["a"].(*values.String).Underlying)
	})

	t.Run("with error only", func(t *testing.T) {
		protoResp := &pb.TriggerResponse{
			Error: "something went wrong",
		}
		resp, err := pb.TriggerResponseFromProto(protoResp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Err)
		assert.Equal(t, "something went wrong", resp.Err.Error())
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := pb.TriggerResponseFromProto(nil)
		assert.ErrorContains(t, err, "could not unmarshal nil")
	})
}

func TestMarshalUnmarshalTriggerRegistrationRequest(t *testing.T) {
	req := capabilities.TriggerRegistrationRequest{
		TriggerID: "test-trigger-id",
		Metadata: capabilities.RequestMetadata{
			WorkflowID:               "test-workflow-id",
			WorkflowExecutionID:      testWorkflowID,
			WorkflowOwner:            testWorkflowOwner,
			WorkflowName:             testWorkflowName,
			WorkflowDonID:            2,
			WorkflowDonConfigVersion: 3,
			ReferenceID:              anyReferenceID,
			DecodedWorkflowName:      "decoded-workflow-name",
			SpendLimits: []capabilities.SpendLimit{
				{SpendType: "GAS", Limit: "5000"},
			},
			WorkflowTag: "workflow-tag",
		},
		Config: &values.Map{Underlying: map[string]values.Value{
			testConfigKey: &values.String{Underlying: testConfigValue},
		}},
		Payload: &anypb.Any{
			TypeUrl: "example.com/payload",
			Value:   []byte("payload-bytes"),
		},
		Method: "trigger-method",
	}

	raw, err := pb.MarshalTriggerRegistrationRequest(req)
	require.NoError(t, err)

	unmarshaled, err := pb.UnmarshalTriggerRegistrationRequest(raw)
	require.NoError(t, err)

	require.EqualValues(t, req.TriggerID, unmarshaled.TriggerID)
	require.EqualValues(t, req.Metadata, unmarshaled.Metadata)
	require.EqualValues(t, req.Config, unmarshaled.Config)
	require.True(t, proto.Equal(req.Payload, unmarshaled.Payload))
	require.EqualValues(t, req.Method, unmarshaled.Method)
}
