package wasm

import (
	"context"
	"errors"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/internal/v2/sdkimpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	basicactionmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction/basic_actionmock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"
	nodeactionmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction/node_actionmock"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils"
)

func TestRuntimeBase_CallCapability(t *testing.T) {
	t.Run("Successful capability call", func(t *testing.T) {
		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}

		action, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)
		expectedInput := &basicaction.Inputs{InputThing: true}
		expectedOutput := &basicaction.Outputs{AdaptedThing: "adapted"}
		action.PerformAction = func(ctx context.Context, input *basicaction.Inputs) (*basicaction.Outputs, error) {
			assert.True(t, proto.Equal(expectedInput, input))
			return expectedOutput, nil
		}

		capability := &basicaction.BasicAction{}
		actual, err := capability.PerformAction(runtime, expectedInput).Await()
		require.NoError(t, err)

		assert.True(t, proto.Equal(expectedOutput, actual))
	})

	t.Run("awaitCapabilities capability errors", func(t *testing.T) {
		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}

		action, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)

		expectedErr := errors.New("error")
		action.PerformAction = func(ctx context.Context, input *basicaction.Inputs) (*basicaction.Outputs, error) {
			return nil, expectedErr
		}

		capability := &basicaction.BasicAction{}
		_, err = capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()

		assert.Equal(t, expectedErr, err)
	})

	t.Run("awaitCapabilities missing response", func(t *testing.T) {
		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}

		_, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)

		overrideCapabilityResponseForTest(t, func() ([]byte, error) {
			missing := &sdkpb.AwaitCapabilitiesResponse{}
			return proto.Marshal(missing)
		})

		capability := &basicaction.BasicAction{}
		_, err = capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()

		assert.ErrorContains(t, err, "cannot find response for ")
	})

	t.Run("callCapability host errors", func(t *testing.T) {
		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		capability := &basicaction.BasicAction{}
		_, err = capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()

		require.ErrorContains(t, err, "cannot find capability "+action.ID())
	})

	t.Run("awaitCapabilities host errors", func(t *testing.T) {
		_, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)

		expectedErr := errors.New("error")
		overrideCapabilityResponseForTest(t, func() ([]byte, error) { return nil, expectedErr })

		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		capability := &basicaction.BasicAction{}
		_, err = capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()

		require.Equal(t, expectedErr, err)
	})

	t.Run("awaitCapabilities unparsable response", func(t *testing.T) {
		_, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)
		overrideCapabilityResponseForTest(t, func() ([]byte, error) { return []byte("invalid"), nil })

		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		capability := &basicaction.BasicAction{}
		_, err = capability.PerformAction(runtime, &basicaction.Inputs{InputThing: true}).Await()

		require.Error(t, err)
	})
}

func TestRuntimeBase_Config(t *testing.T) {
	runtime := newTestRuntime(t)
	assert.Equal(t, anyConfig, runtime.Config())
}

func TestRuntimeBase_LogWriter(t *testing.T) {
	runtime := newTestRuntime(t)
	assert.IsType(t, &writer{}, runtime.LogWriter())
}

func TestDonRuntime_RunInNodeMode(t *testing.T) {
	t.Run("Successful consensus", func(t *testing.T) {
		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		nodeMock, err := nodeactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)
		anyObservation := int32(10)
		anyMedian := int64(11)
		nodeMock.PerformAction = func(ctx context.Context, input *nodeaction.NodeInputs) (*nodeaction.NodeOutputs, error) {
			return &nodeaction.NodeOutputs{OutputThing: anyObservation}, nil
		}
		reg := testutils.GetRegistry(t)
		require.NoError(t, reg.RegisterCapability(&mockConsensus{
			t:           t,
			observation: int64(anyObservation),
			resp:        anyMedian,
		}))

		result, err := sdk.RunInNodeMode(runtime, func(runtime sdk.NodeRuntime) (int64, error) {
			capability := &nodeaction.BasicAction{}
			value, err := capability.PerformAction(runtime, &nodeaction.NodeInputs{InputThing: true}).Await()
			require.NoError(t, err)
			return int64(value.OutputThing), nil
		}, sdkpb.SimpleConsensusType_MEDIAN).Await()
		require.NoError(t, err)
		assert.Equal(t, anyMedian, result)
	})

	t.Run("Failed consensus", func(t *testing.T) {
		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		anyError := errors.New("error")
		reg := testutils.GetRegistry(t)
		require.NoError(t, reg.RegisterCapability(&mockConsensus{
			t:   t,
			err: anyError,
		}))

		_, err := sdk.RunInNodeMode(runtime, func(runtime sdk.NodeRuntime) (int64, error) {
			return 0, anyError
		}, sdkpb.SimpleConsensusType_MEDIAN).Await()
		require.ErrorContains(t, err, anyError.Error())
	})

	t.Run("Does not allow usage of DON runtime in Node mode", func(t *testing.T) {
		drt := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		actionMock, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)
		anyObservation := int32(10)
		actionMock.PerformAction = func(ctx context.Context, input *basicaction.Inputs) (*basicaction.Outputs, error) {
			return &basicaction.Outputs{AdaptedThing: "foo"}, nil
		}
		reg := testutils.GetRegistry(t)
		require.NoError(t, reg.RegisterCapability(&mockConsensus{
			t:           t,
			observation: int64(anyObservation),
			err:         sdk.DonModeCallInNodeMode(),
		}))

		_, err = sdk.RunInNodeMode(drt, func(runtime sdk.NodeRuntime) (int64, error) {
			capability := &basicaction.BasicAction{}
			_, err := capability.PerformAction(drt, &basicaction.Inputs{InputThing: true}).Await()
			return 0, err
		}, sdkpb.SimpleConsensusType_MEDIAN).Await()
		assert.Equal(t, sdk.DonModeCallInNodeMode(), err)
	})

	t.Run("Does not allow usage of Node runtime in DON mode", func(t *testing.T) {
		runtime := &sdkimpl.DonRuntime{RuntimeBase: newTestRuntime(t)}
		nodeMock, err := nodeactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)
		anyObservation := int32(10)
		anyMedian := int64(11)
		nodeMock.PerformAction = func(ctx context.Context, input *nodeaction.NodeInputs) (*nodeaction.NodeOutputs, error) {
			return &nodeaction.NodeOutputs{OutputThing: anyObservation}, nil
		}
		reg := testutils.GetRegistry(t)
		require.NoError(t, reg.RegisterCapability(&mockConsensus{
			t:           t,
			observation: int64(anyObservation),
			resp:        anyMedian,
		}))

		var nrt sdk.NodeRuntime
		_, _ = sdk.RunInNodeMode(runtime, func(runtime sdk.NodeRuntime) (int64, error) {
			nrt = runtime
			return int64(anyObservation), nil
		}, sdkpb.SimpleConsensusType_MEDIAN).Await()
		capability := &nodeaction.BasicAction{}
		_, err = capability.PerformAction(nrt, &nodeaction.NodeInputs{InputThing: true}).Await()
		assert.Equal(t, sdk.NodeModeCallInDonMode(), err)
	})
}

func newTestRuntime(t *testing.T) sdkimpl.RuntimeBase {
	initRunnerAndRuntimeForTest(t, anyExecutionId)
	return newRuntime(executionId, anyMaxResponseSize, anyConfig)
}

type mockConsensus struct {
	t           *testing.T
	observation int64
	err         error
	resp        int64
}

func (m *mockConsensus) Invoke(ctx context.Context, request *sdkpb.CapabilityRequest) *sdkpb.CapabilityResponse {
	assert.Equal(m.t, anyExecutionId, request.ExecutionId)
	assert.Equal(m.t, "consensus@1.0.0", request.Id)
	assert.Equal(m.t, "BuiltIn", request.Method)
	consensus := &sdkpb.BuiltInConsensusRequest{}
	require.NoError(m.t, proto.Unmarshal(request.Payload.Value, consensus))
	switch ct := consensus.PrimitiveConsensus.Consensus.(type) {
	case *sdkpb.PrimitiveConsensus_Simple:
		assert.Equal(m.t, sdkpb.SimpleConsensusType_MEDIAN, ct.Simple)
	default:
		assert.Fail(m.t, "unexpected consensus type")
	}

	resp := &sdkpb.CapabilityResponse{}
	if m.err == nil {
		o, ok := consensus.Observation.(*sdkpb.BuiltInConsensusRequest_Value)
		require.True(m.t, ok)
		assert.Equal(m.t, m.observation, o.Value.GetInt64Value())
		consensusResp := &valuespb.Value{Value: &valuespb.Value_Int64Value{Int64Value: m.resp}}
		a, err := anypb.New(consensusResp)
		require.NoError(m.t, err)

		resp.Response = &sdkpb.CapabilityResponse_Payload{Payload: a}
	} else {
		assert.Equal(m.t, m.err.Error(), consensus.Observation.(*sdkpb.BuiltInConsensusRequest_Error).Error)
		resp.Response = &sdkpb.CapabilityResponse_Error{Error: m.err.Error()}
	}

	return resp
}

func (m *mockConsensus) InvokeTrigger(ctx context.Context, request *sdkpb.TriggerSubscription) (*sdkpb.Trigger, error) {
	return nil, errors.New("not a trigger")
}

func (m *mockConsensus) ID() string {
	return "consensus@1.0.0"
}
