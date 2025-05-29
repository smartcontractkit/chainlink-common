package testutils_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	basicactionmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction/basic_actionmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	basictriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger/basic_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"
	nodeactionmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction/node_actionmock"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils"
)

func TestRuntime_CallCapability(t *testing.T) {
	t.Run("response too large", func(t *testing.T) {
		trigger, err := basictriggermock.NewBasicCapability(t)
		require.NoError(t, err)
		trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			return &basictrigger.Outputs{CoolOutput: "cool"}, nil
		}

		action, err := basicactionmock.NewBasicActionCapability(t)
		require.NoError(t, err)
		action.PerformAction = func(_ context.Context, input *basicaction.Inputs) (*basicaction.Outputs, error) {
			return &basicaction.Outputs{AdaptedThing: strings.Repeat("a", 1000)}, nil
		}

		runner := testutils.NewDonRunner(t, nil)
		runner.SetMaxResponseSizeBytes(1)
		runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
			Handlers: []sdk.Handler[sdk.DonRuntime]{
				sdk.NewDonHandler(
					basictrigger.Basic{}.Trigger(&basictrigger.Config{}),
					func(rt sdk.DonRuntime, _ *basictrigger.Outputs) (string, error) {
						workflowAction1 := &basicaction.BasicAction{}
						call := workflowAction1.PerformAction(rt, &basicaction.Inputs{InputThing: true})
						_, err := call.Await()
						return "", err
					},
				)},
		})

		_, _, err = runner.Result()
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), sdk.ResponseBufferTooSmall))
	})
}

func TestRuntime_ReturnsErrorsFromCapabilitiesThatDoNotExist(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		return &basictrigger.Outputs{CoolOutput: "cool"}, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, _ *basictrigger.Outputs) (string, error) {
					workflowAction1 := &basicaction.BasicAction{}
					call := workflowAction1.PerformAction(rt, &basicaction.Inputs{InputThing: true})
					_, err := call.Await()
					return "", err
				},
			)},
	})

	_, _, err = runner.Result()
	require.Error(t, err)
}

func TestRuntime_ConsensusReturnsTheObservation(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig, config))
		return anyTrigger, nil
	}
	require.NoError(t, err)

	anyValue := int32(100)
	nodeCapability, err := nodeactionmock.NewBasicActionCapability(t)
	require.NoError(t, err)
	nodeCapability.PerformAction = func(_ context.Context, _ *nodeaction.NodeInputs) (*nodeaction.NodeOutputs, error) {
		return &nodeaction.NodeOutputs{OutputThing: anyValue}, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (int32, error) {
					consensus := sdk.RunInNodeMode(rt, func(nodeRuntime sdk.NodeRuntime) (int32, error) {
						action := &nodeaction.BasicAction{}
						resp, err := action.PerformAction(nodeRuntime, &nodeaction.NodeInputs{InputThing: true}).Await()
						require.NoError(t, err)
						return resp.OutputThing, nil
					}, sdk.ConsensusMedianAggregation[int32]())

					consensusResult, err := consensus.Await()
					require.NoError(t, err)
					return consensusResult, nil

				},
			)},
	})

	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, anyValue, result)
}

func TestRuntime_ConsensusReturnsTheDefaultValue(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig, config))
		return anyTrigger, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	anyValue := int32(100)
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (int32, error) {
					consensus := sdk.RunInNodeMode(
						rt,
						func(nodeRuntime sdk.NodeRuntime) (int32, error) {
							return 0, errors.New("no consensus")
						},
						sdk.ConsensusMedianAggregation[int32]().WithDefault(anyValue))

					consensusResult, err := consensus.Await()
					require.NoError(t, err)
					return consensusResult, nil
				},
			)},
	})

	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, anyValue, result)
}

func TestRuntime_ConsensusReturnsErrors(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig, config))
		return anyTrigger, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	anyErr := errors.New("no consensus")
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (int32, error) {
					consensus := sdk.RunInNodeMode(
						rt,
						func(nodeRuntime sdk.NodeRuntime) (int32, error) {
							return 0, anyErr
						},
						sdk.ConsensusMedianAggregation[int32]())

					return consensus.Await()
				},
			)},
	})

	_, _, err = runner.Result()
	require.Equal(t, err, anyErr)
}
