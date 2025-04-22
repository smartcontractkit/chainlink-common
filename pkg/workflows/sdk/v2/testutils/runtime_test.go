package testutils_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/actionandtrigger"
	actionandtriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/actionandtrigger/action_and_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basicaction"
	basicactionmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basicaction/basic_actionmock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basictrigger"
	basictriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basictrigger/basic_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/nodeaction"
	nodeactionmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/nodeaction/node_actionmock"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

func TestRuntime_CallCapabilityIsAsync(t *testing.T) {
	ctx := context.Background()

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig, config))
		return anyTrigger, nil
	}

	ch := make(chan struct{}, 1)
	anyResult1 := "ok1"
	action1, err := basicactionmock.NewBasicActionCapability(t)
	require.NoError(t, err)
	action1.PerformAction = func(_ context.Context, input *basicaction.Inputs) (*basicaction.Outputs, error) {
		<-ch
		return &basicaction.Outputs{AdaptedThing: anyResult1}, nil
	}

	anyResult2 := "ok2"
	action2, err := actionandtriggermock.NewBasicCapability(t)
	action2.Action = func(ctx context.Context, input *actionandtrigger.Input) (*actionandtrigger.Output, error) {
		return &actionandtrigger.Output{Welcome: anyResult2}, nil
	}

	runner, err := testutils.NewDonRunner(t, ctx, nil)
	require.NoError(t, err)

	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, _ *basictrigger.Outputs) (string, error) {
					workflowAction1 := &basicaction.BasicAction{}
					call1 := workflowAction1.PerformAction(rt, &basicaction.Inputs{InputThing: true})

					workflowAction2 := &actionandtrigger.Basic{}
					call2 := workflowAction2.Action(rt, &actionandtrigger.Input{Name: "input"})
					result2, err := call2.Await()
					require.NoError(t, err)
					ch <- struct{}{}
					result1, err := call1.Await()
					require.NoError(t, err)
					return result1.AdaptedThing + result2.Welcome, nil
				}),
		},
	})

	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, anyResult1+anyResult2, result)
}

func TestRuntime_NodeRuntimeUseInDonModeFails(t *testing.T) {
	ctx := context.Background()

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig, config))
		return anyTrigger, nil
	}

	nodeCapability, err := nodeactionmock.NewBasicActionCapability(t)
	require.NoError(t, err)
	nodeCapability.PerformAction = func(_ context.Context, _ *nodeaction.NodeInputs) (*nodeaction.NodeOutputs, error) {
		assert.Fail(t, "node capability should not be called")
		return nil, fmt.Errorf("should not be called")
	}

	runner, err := testutils.NewDonRunner(t, ctx, nil)
	require.NoError(t, err)

	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (*nodeaction.NodeOutputs, error) {
					var nrt sdk.NodeRuntime
					sdk.RunInNodeMode(rt, func(nodeRuntime sdk.NodeRuntime) (int32, error) {
						nrt = nodeRuntime
						return 0, err
					}, pb.SimpleConsensusType_MEDIAN)
					na := nodeaction.BasicAction{}
					return na.PerformAction(nrt, &nodeaction.NodeInputs{InputThing: true}).Await()
				},
			)},
	})

	_, _, err = runner.Result()
	assert.Equal(t, sdk.NodeModeCallInDonMode(), err)
}

func TestRuntime_DonRuntimeUseInNodeModeFails(t *testing.T) {
	ctx := context.Background()

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig, config))
		return anyTrigger, nil
	}

	capability, err := basicactionmock.NewBasicActionCapability(t)
	require.NoError(t, err)
	capability.PerformAction = func(_ context.Context, _ *basicaction.Inputs) (*basicaction.Outputs, error) {
		assert.Fail(t, "should not be called")
		return nil, errors.New("should not be called")
	}

	runner, err := testutils.NewDonRunner(t, ctx, nil)
	require.NoError(t, err)

	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (int32, error) {
					consensus := sdk.RunInNodeMode(rt, func(nodeRuntime sdk.NodeRuntime) (int32, error) {
						action := basicaction.BasicAction{}
						_, err := action.PerformAction(rt, &basicaction.Inputs{InputThing: true}).Await()
						return 0, err
					}, pb.SimpleConsensusType_MEDIAN)

					return consensus.Await()
				},
			)},
	})

	_, _, err = runner.Result()
	assert.Equal(t, sdk.DonModeCallInNodeMode(), err)
}

func TestRuntime_ReturnsErrorsFromCapabilitiesThatDoNotExist(t *testing.T) {
	ctx := context.Background()

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		return &basictrigger.Outputs{CoolOutput: "cool"}, nil
	}

	runner, err := testutils.NewDonRunner(t, ctx, nil)
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
	notRegistered := basicactionmock.BasicActionCapability{}
	require.Equal(t, testutils.NoCapability(notRegistered.ID()), err)
	assert.ErrorContains(t, err, "Capability not found")
	assert.ErrorContains(t, err, notRegistered.ID())
}

func TestRuntime_NumericalConsensusShouldReturnErrorIfInputIsNotNumerical(t *testing.T) {
	assert.Fail(t, "Not written yet")
}

func TestRuntime_ConsensusReturnsTheObservation(t *testing.T) {
	ctx := context.Background()

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

	runner, err := testutils.NewDonRunner(t, ctx, nil)
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
						// TODO should test to make sure median and median of fields are correct, maybe it's just always median though
					}, pb.SimpleConsensusType_MEDIAN)

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
	ctx := context.Background()

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig, config))
		return anyTrigger, nil
	}

	runner, err := testutils.NewDonRunner(t, ctx, nil)
	require.NoError(t, err)

	anyValue := int32(100)
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (int32, error) {
					consensusType := &sdk.PrimitiveConsensusWithDefault[int32]{
						SimpleConsensusType: pb.SimpleConsensusType_MEDIAN,
						DefaultValue:        anyValue,
					}
					consensus := sdk.RunInNodeMode(
						rt,
						func(nodeRuntime sdk.NodeRuntime) (int32, error) {
							return 0, errors.New("no consensus")
						},
						consensusType)

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
	ctx := context.Background()

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig, config))
		return anyTrigger, nil
	}

	runner, err := testutils.NewDonRunner(t, ctx, nil)
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
						pb.SimpleConsensusType_MEDIAN)

					return consensus.Await()
				},
			)},
	})

	_, _, err = runner.Result()
	require.Equal(t, err, anyErr)
}
