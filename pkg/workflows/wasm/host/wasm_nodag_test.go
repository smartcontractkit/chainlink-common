package host

import (
	"context"
	_ "embed"
	"errors"
	"strings"
	"testing"

	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

const (
	nodagBinaryLocation             = "test/nodag/singlehandler/cmd/testmodule.wasm"
	nodagMultiTriggerBinaryLocation = "test/nodag/multihandler/cmd/testmodule.wasm"
	nodagBinaryCmd                  = "test/nodag/singlehandler/cmd"
	nodagMultiTriggerBinaryCmd      = "test/nodag/multihandler/cmd"
)

var wordList = []string{"Hello, ", "world", "!"}

func Test_NoDag_Run(t *testing.T) {
	t.Parallel()

	binary := createTestBinary(nodagBinaryCmd, nodagBinaryLocation, true, t)

	t.Run("NOK fails with unset CapabilityExecutor for trigger", func(t *testing.T) {
		mc := defaultNoDAGModCfg(t)
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		ctx := t.Context()
		req := &wasmpb.ExecuteRequest{
			Request: &wasmpb.ExecuteRequest_Trigger{},
		}

		_, err = m.Execute(ctx, req, nil)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid capability executor")
	})

	t.Run("OK can subscribe without setting CapabilityExecutor", func(t *testing.T) {
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
		mc := &ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
		}
		m, err := NewModule(mc, binary)
		require.NoError(t, err)

		m.Start()
		defer m.Close()

		mockCapExecutor := NewMockCapabilityExecutor(t)

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
			mockCapExecutor.EXPECT().CallCapability(mock.Anything, mock.Anything).
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

		req := &wasmpb.ExecuteRequest{
			Request: &wasmpb.ExecuteRequest_Trigger{
				Trigger: &sdkpb.Trigger{
					Id:      uint64(0),
					Payload: wrapped,
				},
			},
		}

		response, err := m.Execute(ctx, req, mockCapExecutor)
		require.NoError(t, err)

		switch output := response.Result.(type) {
		case *wasmpb.ExecutionResult_Value:
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

func Test_NoDag_MultipleTriggers_Run(t *testing.T) {
	t.Parallel()

	mc := defaultNoDAGModCfg(t)
	capID := (&basictrigger.Basic{}).Trigger(&basictrigger.Config{}).CapabilityID()
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

		mockCapExecutor := NewMockCapabilityExecutor(t)

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
			mockCapExecutor.EXPECT().CallCapability(mock.Anything, mock.Anything).
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

		req := &wasmpb.ExecuteRequest{
			Request: &wasmpb.ExecuteRequest_Trigger{
				Trigger: &sdkpb.Trigger{
					Id:      uint64(1),
					Payload: wrapped,
				},
			},
		}
		response, err := m.Execute(ctx, req, mockCapExecutor)
		require.NoError(t, err)

		switch output := response.Result.(type) {
		case *wasmpb.ExecutionResult_Value:
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

func defaultNoDAGModCfg(t testing.TB) *ModuleConfig {
	return &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
	}
}

func getTriggersSpec(t *testing.T, m ModuleV2, config []byte) (*sdkpb.TriggerSubscriptionRequest, error) {
	execResult, err := m.Execute(t.Context(), &wasmpb.ExecuteRequest{
		Config:  config,
		Request: &wasmpb.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}},
	}, NewMockCapabilityExecutor(t))

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
