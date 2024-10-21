package wasm

import (
	"encoding/base64"
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func Test_Runner_Config_InvalidRequest(t *testing.T) {
	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	runner := &Runner{
		sendResponse: responseFn,
		args:         []string{"wasm", "bla"},
	}
	c := runner.Config()
	assert.Nil(t, c)
	assert.Equal(t, unknownID, gotResponse.Id)
	assert.Contains(t, gotResponse.ErrMsg, "could not decode request")
}

func Test_Runner_Config_InvalidRequest_NotEnoughArgs(t *testing.T) {
	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	runner := &Runner{
		sendResponse: responseFn,
		args:         []string{"wasm"},
	}
	c := runner.Config()
	assert.Nil(t, c)
	assert.Equal(t, unknownID, gotResponse.Id)
	assert.Contains(t, gotResponse.ErrMsg, "request must contain a payload")
}

func marshalRequest(req *wasmpb.Request) (string, error) {
	rpb, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(rpb), nil
}

func Test_Runner_Config(t *testing.T) {
	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	cfg := []byte(`{"hello": "world"}`)
	request := &wasmpb.Request{
		Id:     uuid.New().String(),
		Config: cfg,
	}
	str, err := marshalRequest(request)
	require.NoError(t, err)
	runner := &Runner{
		sendResponse: responseFn,
		args:         []string{"wasm", str},
	}
	c := runner.Config()
	assert.Equal(t, cfg, c)
	assert.Nil(t, gotResponse)
}

func TestRunner_Run_ExecuteCompute(t *testing.T) {
	workflow := sdk.NewWorkflowSpecFactory(
		sdk.NewWorkflowParams{
			Name:  "tester",
			Owner: "cedric",
		},
	)

	trigger := basictrigger.TriggerConfig{Name: "trigger", Number: 100}.New(workflow)
	computeFn := func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {
		return true, nil
	}
	sdk.Compute1(
		workflow,
		"compute",
		sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
		computeFn,
	)

	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	m, err := values.NewMap(map[string]any{
		"cool_output": "a trigger event",
	})
	require.NoError(t, err)

	req := capabilities.CapabilityRequest{
		Config: values.EmptyMap(),
		Inputs: m,
		Metadata: capabilities.RequestMetadata{
			ReferenceID: "compute",
		},
	}
	reqpb := capabilitiespb.CapabilityRequestToProto(req)
	request := &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_ComputeRequest{
			ComputeRequest: &wasmpb.ComputeRequest{
				Request: reqpb,
			},
		},
	}
	str, err := marshalRequest(request)
	require.NoError(t, err)
	runner := &Runner{
		args:         []string{"wasm", str},
		sendResponse: responseFn,
		sdkFactory: func(cfg *RuntimeConfig, id string, md *capabilities.RequestMetadata, _ ...func(*RuntimeConfig)) *Runtime {
			return nil
		},
	}
	runner.Run(workflow)

	assert.NotNil(t, gotResponse.GetComputeResponse())

	resp := gotResponse.GetComputeResponse().GetResponse()
	assert.Equal(t, resp.Error, "")

	m, err = values.FromMapValueProto(resp.Value)
	require.NoError(t, err)

	unw, err := values.Unwrap(m)
	require.NoError(t, err)

	assert.Equal(t, unw.(map[string]any)["Value"].(bool), true)
}

func TestRunner_Run_GetWorkflowSpec(t *testing.T) {
	workflow := sdk.NewWorkflowSpecFactory(
		sdk.NewWorkflowParams{
			Name:  "tester",
			Owner: "cedric",
		},
	)

	trigger := basictrigger.TriggerConfig{Name: "trigger", Number: 100}.New(workflow)
	computeFn := func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {
		return true, nil
	}
	sdk.Compute1(
		workflow,
		"compute",
		sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
		computeFn,
	)

	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	request := &wasmpb.Request{
		Id: uuid.New().String(),
		Message: &wasmpb.Request_SpecRequest{
			SpecRequest: &emptypb.Empty{},
		},
	}
	str, err := marshalRequest(request)
	require.NoError(t, err)
	runner := &Runner{
		args:         []string{"wasm", str},
		sendResponse: responseFn,
	}
	runner.Run(workflow)

	resp := gotResponse.GetSpecResponse()
	assert.NotNil(t, resp)

	spc, err := wasmpb.ProtoToWorkflowSpec(resp)
	require.NoError(t, err)

	gotSpec, err := workflow.Spec()
	require.NoError(t, err)

	// Do some massaging due to protos lossy conversion of types
	gotSpec.Triggers[0].Inputs.Mapping = map[string]any{}
	gotSpec.Triggers[0].Config["number"] = int64(gotSpec.Triggers[0].Config["number"].(uint64))
	assert.Equal(t, &gotSpec, spc)
}

// Test_createEmitFn validates the runtime's emit function implementation.  Uses mocks of the
// imported wasip1 emit function.
func Test_createEmitFn(t *testing.T) {
	var (
		l         = logger.Test(t)
		sdkConfig = &RuntimeConfig{
			MaxFetchResponseSizeBytes: 1_000,
			Metadata: &capabilities.RequestMetadata{
				WorkflowID:          "workflow_id",
				WorkflowExecutionID: "workflow_execution_id",
				WorkflowName:        "workflow_name",
				WorkflowOwner:       "workflow_owner_address",
			},
		}
		giveMsg    = "testing guest"
		giveLabels = map[string]any{
			"some-key": "some-value",
		}
	)

	t.Run("success", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 0
		}
		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.NoError(t, err)
	})

	t.Run("successfully read error message when emit fails", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			// marshall the protobufs
			b, err := proto.Marshal(&wasmpb.EmitMessageResponse{
				Error: &wasmpb.Error{
					Message: assert.AnError.Error(),
				},
			})
			assert.NoError(t, err)

			// write the marshalled response message to memory
			resp := unsafe.Slice((*byte)(respptr), len(b))
			copy(resp, b)

			// write the length of the response to memory in little endian
			respLen := unsafe.Slice((*byte)(resplenptr), uint32Size)
			binary.LittleEndian.PutUint32(respLen, uint32(len(b)))

			return 0
		}
		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.Error(t, err)
		assert.ErrorContains(t, err, assert.AnError.Error())
	})

	t.Run("fail to deserialize response from memory", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			// b is a non-protobuf byte slice
			b := []byte(assert.AnError.Error())

			// write the marshalled response message to memory
			resp := unsafe.Slice((*byte)(respptr), len(b))
			copy(resp, b)

			// write the length of the response to memory in little endian
			respLen := unsafe.Slice((*byte)(resplenptr), uint32Size)
			binary.LittleEndian.PutUint32(respLen, uint32(len(b)))

			return 0
		}

		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "invalid wire-format data")
	})

	t.Run("fail with nonzero code from emit", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 42
		}
		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "emit failed with errno 42")
	})
}
