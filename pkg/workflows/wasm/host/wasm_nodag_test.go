package host

import (
	"context"
	_ "embed"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmdagpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/legacy/pb"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

// TODO this should go generate the capability clients

const (
	nodagBinaryLocation = "test/nodag/cmd/testmodule.wasm"
	nodagBinaryCmd      = "test/nodag/cmd"
)

func Test_NoDag_Run(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	lock := sync.Mutex{}
	anyExecId := "executionId"
	numRequests := byte(0)
	numAwaits := byte(0)

	mc := &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
		CallCapability: func(ctx context.Context, req *wasmpb.CapabilityRequest) ([sdk.IdLen]byte, error) {
			require.Equal(t, anyExecId, req.Id)
			require.Equal(t, "basic-action@1.0.0", req.Id)
			inputs := basicaction.Inputs{}
			err := req.Payload.UnmarshalTo(&inputs)
			require.NoError(t, err)
			lock.Lock()
			defer lock.Unlock()

			require.Equal(t, inputs.InputThing, numRequests == 0)

			id := [sdk.IdLen]byte{}
			b := byte(0)
			if inputs.InputThing {
				b = byte(1)
			}
			for i := 0; i < sdk.IdLen; i++ {
				id[i] = b
			}

			return id, nil
		},
		AwaitCapabilities: func(ctx context.Context, req *wasmpb.AwaitCapabilitiesRequest) (*wasmpb.AwaitCapabilitiesResponse, error) {
			require.Equal(t, anyExecId, req.ExecId)
			require.Len(t, req.Ids, 1)
			lock.Lock()
			defer lock.Unlock()
			for i := 0; i < sdk.IdLen; i++ {
				require.Equal(t, numAwaits, req.Ids[0][i])
			}
			numAwaits++
			resp := &basicaction.Outputs{AdaptedThing: fmt.Sprintf("response-%d", numAwaits)}
			payload, err := anypb.New(resp)
			require.NoError(t, err)

			return &wasmpb.AwaitCapabilitiesResponse{
				Responses: map[string]*wasmpb.CapabilityResponse{
					req.Ids[0][:]: {
						Response: &wasmpb.CapabilityResponse_Payload{
							Payload: payload,
						},
					},
				},
			}, nil
		},
	}

	triggerID := "basic-trigger@1.0.0"
	binary := createTestBinary(nodagBinaryCmd, nodagBinaryLocation, true, t)
	triggers, err := GetTriggersSpec(ctx, mc, binary, []byte(""))
	require.NoError(t, err)

	require.Len(t, triggers, 1)
	require.Equal(t, triggerID, triggers[0].Id)
	require.Equal(t, anyExecId, triggers[0].ExecId)
	configProto := triggers[0].Payload
	config := &basictrigger.Config{}
	require.NoError(t, configProto.UnmarshalTo(config))
	require.Equal(t, "name", config.Name)
	require.Equal(t, int32(100), config.Number)

	m, err := NewModule(mc, binary)
	require.NoError(t, err)
	m.Start()

	// When a TriggerEvent occurs, Engine calls Run() with that Event.
	trigger := &basictrigger.Outputs{CoolOutput: "Hi"}
	wrapped, err := anypb.New(trigger)
	require.NoError(t, err)

	// TODO test config
	req := &wasmpb.ExecuteRequest{
		Id: anyExecId,
		Request: &wasmpb.ExecuteRequest_Trigger{
			Trigger: &wasmpb.Trigger{
				Id:      triggerID,
				Payload: wrapped,
			},
		},
	}
	response, err := m.Execute(ctx, req)
	require.NoError(t, err)

	require.Equal(t, anyExecId, response.Id)
	switch output := response.Result.(type) {
	case *wasmpb.ExecutionResult_Value:
		valuePb := output.Value
		value, err := values.FromProto(valuePb)
		require.NoError(t, err)
		unwrapped, err := value.Unwrap()
		require.NoError(t, err)
		require.Equal(t, "Hiresponse-0response-1", unwrapped)
	default:
		t.Fatalf("unexpected response type %T", output)
	}
}

func Test_NoDag_With_Legacy_Run(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	lock := sync.Mutex{}
	anyExecId := "executionId"
	numRequests := byte(0)
	numAwaits := byte(0)

	mc := &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
		CallCapability: func(ctx context.Context, req *wasmpb.CapabilityRequest) ([sdk.IdLen]byte, error) {
			require.Equal(t, anyExecId, req.Id)
			require.Equal(t, "basic-action@1.0.0", req.Id)
			inputs := basicaction.Inputs{}
			err := req.Payload.UnmarshalTo(&inputs)
			require.NoError(t, err)
			lock.Lock()
			defer lock.Unlock()

			require.Equal(t, inputs.InputThing, numRequests == 0)

			id := [sdk.IdLen]byte{}
			b := byte(0)
			if inputs.InputThing {
				b = byte(1)
			}
			for i := 0; i < sdk.IdLen; i++ {
				id[i] = b
			}

			return id, nil
		},
		AwaitCapabilities: func(ctx context.Context, req *wasmpb.AwaitCapabilitiesRequest) (*wasmpb.AwaitCapabilitiesResponse, error) {
			require.Equal(t, anyExecId, req.ExecId)
			require.Len(t, req.Ids, 1)
			lock.Lock()
			defer lock.Unlock()
			for i := 0; i < sdk.IdLen; i++ {
				require.Equal(t, numAwaits, req.Ids[0][i])
			}
			numAwaits++
			resp := &basicaction.Outputs{AdaptedThing: fmt.Sprintf("response-%d", numAwaits)}
			payload, err := anypb.New(resp)
			require.NoError(t, err)

			return &wasmpb.AwaitCapabilitiesResponse{
				Responses: map[string]*wasmpb.CapabilityResponse{
					req.Ids[0][:]: {
						Response: &wasmpb.CapabilityResponse_Payload{
							Payload: payload,
						},
					},
				},
			}, nil
		},
	}

	triggerID := "basic-trigger@1.0.0"
	triggerRef := "trigger-0"
	binary := createTestBinary(nodagBinaryCmd, nodagBinaryLocation, true, t)
	spec, err := GetWorkflowSpec(ctx, mc, binary, []byte(""))
	require.NoError(t, err)

	// Engine expects only triggers to be included in the returned spec. Trigger subscriptions are performed.
	//
	//  [trigger]
	//      |
	//  (promise)
	require.Len(t, spec.Triggers, 1)
	require.Equal(t, triggerID, spec.Triggers[0].ID)
	require.Equal(t, triggerRef, spec.Triggers[0].Ref)
	configProto := spec.Triggers[0].ConfigProto
	config := &basictrigger.Config{}
	require.NoError(t, configProto.UnmarshalTo(config))
	require.Equal(t, "name", config.Name)
	require.Equal(t, int32(100), config.Number)
	require.Len(t, spec.Actions, 0)
	require.Len(t, spec.Consensus, 0)
	require.Len(t, spec.Targets, 0)

	m, err := NewModule(mc, binary)
	require.NoError(t, err)
	m.Start()

	// When a TriggerEvent occurs, Engine calls Run() with that Event.
	trigger := &basictrigger.Outputs{CoolOutput: "Hi"}
	wrapped, err := anypb.New(trigger)

	// TODO test config
	req := &wasmdagpb.Request{
		Id: anyExecId,
		Message: &wasmdagpb.Request_ComputeRequest{
			ComputeRequest: &wasmdagpb.ComputeRequest{
				Request: &pb.CapabilityRequest{
					Request: wrapped,
				},
			},
		},
		TriggerId: triggerID,
	}
	response, err := m.Run(ctx, req)
	require.NoError(t, err)
	require.Empty(t, response.ErrMsg)
	require.Equal(t, anyExecId, response.Id)
	switch output := response.Message.(type) {
	case *wasmdagpb.Response_ComputeResponse:
		valuePb := output.ComputeResponse.Response.Value.Fields["Output"]
		value, err := values.FromProto(valuePb)
		require.NoError(t, err)
		unwrapped, err := value.Unwrap()
		require.NoError(t, err)
		require.Equal(t, "Hiresponse-0response-1", unwrapped)
	default:
		t.Fatalf("unexpected response type %T", output)
	}
}
