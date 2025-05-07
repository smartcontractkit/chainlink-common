package host

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

const (
	nodagBinaryLocation             = "test/nodag/singlehandler/cmd/testmodule.wasm"
	nodagMultiTriggerBinaryLocation = "test/nodag/multihandler/cmd/testmodule.wasm"
	nodagBinaryCmd                  = "test/nodag/singlehandler/cmd"
	nodagMultiTriggerBinaryCmd      = "test/nodag/multihandler/cmd"
)

const anyNoDagExecId = "executionId"

func Test_NoDag_Run(t *testing.T) {
	t.Parallel()

	var (
		mc           = createNoDagMc(t)
		triggerIndex = int(0)
		capID        = (&basictrigger.Basic{}).Trigger(&basictrigger.Config{}).CapabilityID()
		binary       = createTestBinary(nodagBinaryCmd, nodagBinaryLocation, true, t)
	)

	m, err := NewModule(mc, binary)
	require.NoError(t, err)

	m.Start()
	defer m.Close()

	t.Run("Subscribe to triggers", func(t *testing.T) {
		ctx := t.Context()
		triggers, err := getTriggersSpec(ctx, m, []byte(""))
		require.NoError(t, err)

		require.Len(t, triggers.Subscriptions, 1)
		require.Equal(t,
			capID,
			triggers.Subscriptions[triggerIndex].Id,
		)
		configProto := triggers.Subscriptions[0].Payload
		config := &basictrigger.Config{}
		require.NoError(t, configProto.UnmarshalTo(config))
		require.Equal(t, "name", config.Name)
		require.Equal(t, int32(100), config.Number)
	})

	t.Run("Execute trigger", func(t *testing.T) {
		ctx := t.Context()
		// When a TriggerEvent occurs, Engine calls Execute with that Event.
		trigger := &basictrigger.Outputs{CoolOutput: "Hi"}
		wrapped, err := anypb.New(trigger)
		require.NoError(t, err)

		// TODO test config
		req := &wasmpb.ExecuteRequest{
			Id: anyNoDagExecId,
			Request: &wasmpb.ExecuteRequest_Trigger{
				Trigger: &sdkpb.Trigger{
					Id:      uint64(triggerIndex),
					Payload: wrapped,
				},
			},
		}
		response, err := m.Execute(ctx, req)
		require.NoError(t, err)

		require.Equal(t, anyNoDagExecId, response.Id)
		switch output := response.Result.(type) {
		case *wasmpb.ExecutionResult_Value:
			valuePb := output.Value
			value, err := values.FromProto(valuePb)
			require.NoError(t, err)
			unwrapped, err := value.Unwrap()
			require.NoError(t, err)
			require.Equal(t, "Hiresponse-1response-2", unwrapped)
		default:
			t.Fatalf("unexpected response type %T", output)
		}
	})
}

func Test_NoDag_MultipleTriggers_Run(t *testing.T) {
	t.Parallel()

	var (
		mc           = createNoDagMc(t)
		triggerIndex = int(0)
		capID        = (&basictrigger.Basic{}).Trigger(&basictrigger.Config{}).CapabilityID()
		binary       = createTestBinary(nodagMultiTriggerBinaryCmd, nodagMultiTriggerBinaryLocation, true, t)
	)

	m, err := NewModule(mc, binary)
	require.NoError(t, err)

	m.Start()
	defer m.Close()

	t.Run("Subscribe to triggers", func(t *testing.T) {
		ctx := t.Context()
		triggers, err := getTriggersSpec(ctx, m, []byte(""))
		require.NoError(t, err)

		expectedConfigs := []*basictrigger.Config{
			{
				Name:   "name",
				Number: 100,
			},
			{
				Name:   "second-trigger",
				Number: 200,
			},
		}

		require.Len(t, triggers.Subscriptions, 2)

		// expect same name for all subscriptions/triggers
		for idx := range len(triggers.Subscriptions) {
			require.Equal(t,
				capID,
				triggers.Subscriptions[idx].Id,
			)
			configProto := triggers.Subscriptions[idx].Payload
			config := &basictrigger.Config{}
			require.NoError(t, configProto.UnmarshalTo(config))
			require.Equal(t, expectedConfigs[idx].Name, config.Name)
			require.Equal(t, expectedConfigs[idx].Number, config.Number)
		}
	})

	t.Run("Execute trigger", func(t *testing.T) {
		ctx := t.Context()
		// When a TriggerEvent occurs, Engine calls Execute with that Event.
		trigger := &basictrigger.Outputs{CoolOutput: "Hi"}
		wrapped, err := anypb.New(trigger)
		require.NoError(t, err)

		// Trigger second handler
		req := &wasmpb.ExecuteRequest{
			Id: anyNoDagExecId,
			Request: &wasmpb.ExecuteRequest_Trigger{
				Trigger: &sdkpb.Trigger{
					Id:      uint64(triggerIndex + 1),
					Payload: wrapped,
				},
			},
		}
		response, err := m.Execute(ctx, req)
		require.NoError(t, err)

		require.Equal(t, anyNoDagExecId, response.Id)
		switch output := response.Result.(type) {
		case *wasmpb.ExecutionResult_Value:
			valuePb := output.Value
			value, err := values.FromProto(valuePb)
			require.NoError(t, err)
			unwrapped, err := value.Unwrap()
			require.NoError(t, err)
			require.Equal(t, "Hiresponse-1response-2"+"true", unwrapped)
		default:
			t.Fatalf("unexpected response type %T", output)
		}
	})
}

func createNoDagMc(t *testing.T) *ModuleConfig {
	lock := sync.Mutex{}
	numRequests := byte(0)
	numAwaits := byte(0)

	mc := &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
		CallCapability: func(ctx context.Context, req *sdkpb.CapabilityRequest) ([sdk.IdLen]byte, error) {
			require.Equal(t, anyNoDagExecId, req.ExecutionId)
			require.Equal(t, "basic-test-action@1.0.0", req.Id)
			inputs := basicaction.Inputs{}
			err := req.Payload.UnmarshalTo(&inputs)
			require.NoError(t, err)
			lock.Lock()
			defer lock.Unlock()

			require.Equal(t, inputs.InputThing, numRequests != 0)
			numRequests++

			id := [sdk.IdLen]byte{}
			b := byte(0)
			if inputs.InputThing {
				b = byte(1)
			}
			for i := 0; i < sdk.IdLen; i++ {
				id[i] = 'a' + b
			}

			return id, nil
		},
		AwaitCapabilities: func(ctx context.Context, req *sdkpb.AwaitCapabilitiesRequest) (*sdkpb.AwaitCapabilitiesResponse, error) {
			require.Equal(t, anyNoDagExecId, req.ExecId)
			require.Len(t, req.Ids, 1)
			lock.Lock()
			defer lock.Unlock()
			for i := 0; i < sdk.IdLen; i++ {
				require.Equal(t, 'a'+numAwaits, req.Ids[0][i])
			}
			numAwaits++
			resp := &basicaction.Outputs{AdaptedThing: fmt.Sprintf("response-%d", numAwaits)}
			payload, err := anypb.New(resp)
			require.NoError(t, err)

			return &sdkpb.AwaitCapabilitiesResponse{
				Responses: map[string]*sdkpb.CapabilityResponse{
					req.Ids[0][:]: {
						Response: &sdkpb.CapabilityResponse_Payload{
							Payload: payload,
						},
					},
				},
			}, nil
		},
	}
	return mc
}

func getTriggersSpec(ctx context.Context, m ModuleV2, config []byte) (*sdkpb.TriggerSubscriptionRequest, error) {
	execResult, err := m.Execute(ctx, &wasmpb.ExecuteRequest{
		Id:      uuid.New().String(),
		Config:  config,
		Request: &wasmpb.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}},
	})

	if err != nil {
		return nil, err
	}

	switch r := execResult.Result.(type) {
	case *wasmpb.ExecutionResult_TriggerSubscriptions:
		return r.TriggerSubscriptions, nil
	case *wasmpb.ExecutionResult_Error:
		return nil, errors.New(r.Error)
	default:
		return nil, errors.New("unexpected response from WASM binary: got nil spec response")
	}
}
