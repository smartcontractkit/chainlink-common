package host

import (
	"context"
	_ "embed"
	"errors"
	"strings"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	nodagBinaryCmd                  = "test/nodag/singlehandler/cmd"
	nodagBinaryLocation             = nodagBinaryCmd + "/testmodule.wasm"
	nodagMultiTriggerBinaryCmd      = "test/nodag/multihandler/cmd"
	nodagMultiTriggerBinaryLocation = nodagMultiTriggerBinaryCmd + "/testmodule.wasm"
	nodagRandomBinaryCmd            = "test/nodag/randoms/cmd"
	nodagRandomBinaryLocation       = nodagRandomBinaryCmd + "/testmodule.wasm"
)

var wordList = []string{"Hello, ", "world", "!"}

func Test_NoDag_Run(t *testing.T) {
	t.Parallel()

	binary := createTestBinary(nodagBinaryCmd, nodagBinaryLocation, true, t)

	t.Run("NOK fails with unset ExecutionHelper for trigger", func(t *testing.T) {
		mc := defaultNoDAGModCfg(t)
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		ctx := t.Context()
		req := &sdkpb.ExecuteRequest{
			Request: &sdkpb.ExecuteRequest_Trigger{},
		}

		_, err = m.Execute(ctx, req, nil)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid capability executor")
	})

	t.Run("OK can subscribe without setting ExecutionHelper", func(t *testing.T) {
		mc := defaultNoDAGModCfg(t)
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		triggers, err := getTriggersSpec(t, m, []byte(""))
		require.NoError(t, err)
		require.Equal(t, len(triggers.Subscriptions), 1)
	})

	t.Run("OK executes happy path with two awaits", func(t *testing.T) {
		ctx := t.Context()
		wantResponse := strings.Join(wordList, "")
		lggr, observer := logger.TestObserved(t, zapcore.InfoLevel)
		mc := &ModuleConfig{
			Logger:         lggr,
			IsUncompressed: true,
		}
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetID().Return("Id")

		// wrap some common payload
		newWantedCapResponse := func(i int) *sdkpb.CapabilityResponse {
			action := &basicaction.Outputs{AdaptedThing: wordList[i]}
			anyAction, err := anypb.New(action)
			require.NoError(t, err)

			return &sdkpb.CapabilityResponse{
				Response: &sdkpb.CapabilityResponse_Payload{
					Payload: anyAction,
				}}
		}

		for i := 1; i < len(wordList); i++ {
			wantCapResp := newWantedCapResponse(i)
			mockExecutionHelper.EXPECT().CallCapability(mock.Anything, mock.Anything).
				Run(
					func(ctx context.Context, request *sdkpb.CapabilityRequest) {
						require.Equal(t, "basic-test-action@1.0.0", request.Id)
					},
				).
				Return(wantCapResp, nil).
				Once()
		}

		// When a TriggerEvent occurs, Engine calls Execute with that Event.
		trigger := &basictrigger.Outputs{CoolOutput: wordList[0]}
		wrapped, err := anypb.New(trigger)
		require.NoError(t, err)

		req := &sdkpb.ExecuteRequest{
			Request: &sdkpb.ExecuteRequest_Trigger{
				Trigger: &sdkpb.Trigger{
					Id:      uint64(0),
					Payload: wrapped,
				},
			},
		}

		response, err := m.Execute(ctx, req, mockExecutionHelper)
		require.NoError(t, err)

		logs := observer.TakeAll()
		require.Len(t, logs, 1)
		assert.True(t, strings.Contains(logs[0].Message, "Hi"))

		switch output := response.Result.(type) {
		case *sdkpb.ExecutionResult_Value:
			valuePb := output.Value
			value, err := values.FromProto(valuePb)
			require.NoError(t, err)
			unwrapped, err := value.Unwrap()
			require.NoError(t, err)
			require.Equal(t, wantResponse, unwrapped)
		default:
			t.Fatalf("unexpected response type %T", output)
		}
	})
}

func Test_NoDag_Secrets(t *testing.T) {
	t.Parallel()

	binary := createTestBinary(nodagSecretsBinaryCmd, nodagSecretsBinaryLocation, true, t)

	t.Run("Returns an error if the secret doesn't exist", func(t *testing.T) {
		ctx := t.Context()
		lggr, _ := logger.TestObserved(t, zapcore.InfoLevel)
		mc := &ModuleConfig{
			Logger:         lggr,
			IsUncompressed: true,
		}
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("Id")

		mockExecutionHelper.EXPECT().GetSecrets(mock.Anything, mock.Anything).
			Return([]*sdkpb.SecretResponse{
				{
					Response: &sdkpb.SecretResponse_Error{
						Error: "could not find secret",
					},
				},
			}, nil).
			Once()

		// When a TriggerEvent occurs, Engine calls Execute with that Event.
		trigger := &basictrigger.Outputs{CoolOutput: wordList[0]}
		wrapped, err := anypb.New(trigger)
		require.NoError(t, err)

		req := &sdkpb.ExecuteRequest{
			Request: &sdkpb.ExecuteRequest_Trigger{
				Trigger: &sdkpb.Trigger{
					Id:      uint64(0),
					Payload: wrapped,
				},
			},
		}

		resp, err := m.Execute(ctx, req, mockExecutionHelper)
		require.NoError(t, err)

		assert.ErrorContains(t, errors.New(resp.GetError()), "could not find secret")
	})

	t.Run("Returns the secret if it exists", func(t *testing.T) {
		ctx := t.Context()
		lggr, _ := logger.TestObserved(t, zapcore.InfoLevel)
		mc := &ModuleConfig{
			Logger:         lggr,
			IsUncompressed: true,
		}
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("Id")

		mockExecutionHelper.EXPECT().GetSecrets(mock.Anything, mock.Anything).
			Return([]*sdkpb.SecretResponse{
				{
					Response: &sdkpb.SecretResponse_Secret{
						Secret: &sdkpb.Secret{Value: "Bar"},
					},
				},
			}, nil).
			Once()

		// When a TriggerEvent occurs, Engine calls Execute with that Event.
		trigger := &basictrigger.Outputs{CoolOutput: wordList[0]}
		wrapped, err := anypb.New(trigger)
		require.NoError(t, err)

		req := &sdkpb.ExecuteRequest{
			Request: &sdkpb.ExecuteRequest_Trigger{
				Trigger: &sdkpb.Trigger{
					Id:      uint64(0),
					Payload: wrapped,
				},
			},
		}

		resp, err := m.Execute(ctx, req, mockExecutionHelper)
		require.NoError(t, err)
		assert.Equal(t, "", resp.GetError())

		assert.Equal(t, "Bar", resp.GetValue().GetStringValue())
	})
}

func Test_NoDag_MultipleTriggers_Run(t *testing.T) {
	t.Parallel()

	mc := defaultNoDAGModCfg(t)
	capID := basictrigger.Trigger(&basictrigger.Config{}).CapabilityID()
	binary := createTestBinary(nodagMultiTriggerBinaryCmd, nodagMultiTriggerBinaryLocation, true, t)

	t.Run("OK subscribe to triggers with identical capability IDs", func(t *testing.T) {
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		triggers, err := getTriggersSpec(t, m, []byte(""))
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

		// Assert on subscriptions
		require.Len(t, triggers.Subscriptions, 2)
		for idx := range len(triggers.Subscriptions) {
			// expect same capability ID for all triggers
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

	t.Run("OK executes happy path with multiple triggers for same capability", func(t *testing.T) {
		ctx := t.Context()
		wantResponse := strings.Join(wordList, "") + "true"
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		mockExecutionHelper := NewMockExecutionHelper(t)
		mockExecutionHelper.EXPECT().GetID().Return("Id")

		newWantedCapResponse := func(i int) *sdkpb.CapabilityResponse {
			action := &basicaction.Outputs{AdaptedThing: wordList[i]}
			anyAction, err := anypb.New(action)
			require.NoError(t, err)

			return &sdkpb.CapabilityResponse{
				Response: &sdkpb.CapabilityResponse_Payload{
					Payload: anyAction,
				}}
		}

		for i := 1; i < len(wordList); i++ {
			wantCapResp := newWantedCapResponse(i)
			mockExecutionHelper.EXPECT().CallCapability(mock.Anything, mock.Anything).
				Run(
					func(ctx context.Context, request *sdkpb.CapabilityRequest) {
						require.Equal(t, "basic-test-action@1.0.0", request.Id)
					},
				).
				Return(wantCapResp, nil).
				Once()
		}

		// When a TriggerEvent occurs, Engine calls Execute with that Event.
		trigger := &basictrigger.Outputs{CoolOutput: wordList[0]}
		wrapped, err := anypb.New(trigger)
		require.NoError(t, err)

		req := &sdkpb.ExecuteRequest{
			Request: &sdkpb.ExecuteRequest_Trigger{
				Trigger: &sdkpb.Trigger{
					Id:      uint64(1),
					Payload: wrapped,
				},
			},
		}
		response, err := m.Execute(ctx, req, mockExecutionHelper)
		require.NoError(t, err)

		switch output := response.Result.(type) {
		case *sdkpb.ExecutionResult_Value:
			valuePb := output.Value
			value, err := values.FromProto(valuePb)
			require.NoError(t, err)
			unwrapped, err := value.Unwrap()
			require.NoError(t, err)
			require.Equal(t, wantResponse, unwrapped)
		default:
			t.Fatalf("unexpected response type %T", output)
		}
	})
}

func Test_NoDag_Random(t *testing.T) {
	t.Parallel()

	mc := defaultNoDAGModCfg(t)
	lggr, observed := logger.TestObserved(t, zapcore.DebugLevel)
	mc.Logger = lggr

	binary := createTestBinary(nodagRandomBinaryCmd, nodagRandomBinaryLocation, true, t)

	m, err := NewModule(mc, binary)
	require.NoError(t, err)

	// Test binary executes node mode code conditionally based on the value >= 100
	anyId := "Id"
	gte100Exec := NewMockExecutionHelper(t)
	gte100Exec.EXPECT().GetID().Return(anyId)
	gte100 := &nodeaction.NodeOutputs{OutputThing: 120}
	gte100Payload, err := anypb.New(gte100)
	require.NoError(t, err)

	gte100Exec.EXPECT().CallCapability(mock.Anything, mock.Anything).Return(&sdkpb.CapabilityResponse{
		Response: &sdkpb.CapabilityResponse_Payload{
			Payload: gte100Payload,
		},
	}, nil)

	m.Start()
	defer m.Close()

	trigger := &basictrigger.Outputs{CoolOutput: "trigger1"}
	triggerPayload, err := anypb.New(trigger)
	require.NoError(t, err)
	anyRequest := &sdkpb.ExecuteRequest{
		Request: &sdkpb.ExecuteRequest_Trigger{
			Trigger: &sdkpb.Trigger{
				Id:      uint64(0),
				Payload: triggerPayload,
			},
		},
	}
	execution1Result, err := m.Execute(t.Context(), anyRequest, gte100Exec)
	require.NoError(t, err)
	wrappedValue1, err := values.FromProto(execution1Result.GetValue())
	require.NoError(t, err)
	value1, err := wrappedValue1.Unwrap()
	require.NoError(t, err)

	t.Run("Same execution id gives the same randoms, even if random is called in node mode", func(t *testing.T) {
		// Clear from any previous test
		observed.TakeAll()

		lt100Exec := NewMockExecutionHelper(t)
		lt100Exec.EXPECT().GetID().Return(anyId)
		lt100 := &nodeaction.NodeOutputs{OutputThing: 120}
		lt100Payload, err := anypb.New(lt100)
		require.NoError(t, err)

		lt100Exec.EXPECT().CallCapability(mock.Anything, mock.Anything).Return(&sdkpb.CapabilityResponse{
			Response: &sdkpb.CapabilityResponse_Payload{
				Payload: lt100Payload,
			},
		}, nil)

		exectuion2Result, err := m.Execute(t.Context(), anyRequest, lt100Exec)
		require.NoError(t, err)
		wrappedValue2, err := values.FromProto(exectuion2Result.GetValue())
		require.NoError(t, err)
		value2, err := wrappedValue2.Unwrap()
		require.NoError(t, err)
		require.Equal(t, value1, value2, "Expected the same random number to be generated for the same trigger")
	})

	t.Run("Different execution id give different randoms", func(t *testing.T) {
		require.NoError(t, err)

		gte100Exec2 := NewMockExecutionHelper(t)
		gte100Exec2.EXPECT().GetID().Return("differentId")

		gte100Exec2.EXPECT().CallCapability(mock.Anything, mock.Anything).Return(&sdkpb.CapabilityResponse{
			Response: &sdkpb.CapabilityResponse_Payload{
				Payload: gte100Payload,
			},
		}, nil)

		executionResult2, err := m.Execute(t.Context(), anyRequest, gte100Exec2)
		require.NoError(t, err)
		wrappedValue2, err := values.FromProto(executionResult2.GetValue())
		require.NoError(t, err)
		value2, err := wrappedValue2.Unwrap()
		require.NoError(t, err)
		require.NotEqual(t, value1, value2, "Expected different random numbers for different triggers")
	})
}

func defaultNoDAGModCfg(t testing.TB) *ModuleConfig {
	return &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
	}
}

func getTriggersSpec(t *testing.T, m ModuleV2, config []byte) (*sdkpb.TriggerSubscriptionRequest, error) {
	helper := NewMockExecutionHelper(t)
	helper.EXPECT().GetWorkflowExecutionID().Return("Id")
	execResult, err := m.Execute(t.Context(), &sdkpb.ExecuteRequest{
		Config:  config,
		Request: &sdkpb.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}},
	}, helper)

	if err != nil {
		return nil, err
	}

	switch r := execResult.Result.(type) {
	case *sdkpb.ExecutionResult_TriggerSubscriptions:
		return r.TriggerSubscriptions, nil
	case *sdkpb.ExecutionResult_Error:
		return nil, errors.New(r.Error)
	default:
		return nil, errors.New("unexpected response from WASM binary: got nil spec response")
	}
}
