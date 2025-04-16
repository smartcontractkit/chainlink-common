package testutils_test

import (
	"context"
	"errors"
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
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func TestRuntime_CallCapabilityIsAsync(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			assert.True(t, proto.Equal(anyConfig, config))
			return anyTrigger, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	ch := make(chan struct{}, 1)
	anyResult1 := "ok1"
	action1 := &basicactionmock.BasicActionCapability{
		PerformAction: func(_ context.Context, input *basicaction.Inputs) (*basicaction.Outputs, error) {
			<-ch
			return &basicaction.Outputs{AdaptedThing: anyResult1}, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(action1))

	anyResult2 := "ok2"
	action2 := &actionandtriggermock.BasicCapability{
		Action: func(ctx context.Context, input *actionandtrigger.Input) (*actionandtrigger.Output, error) {
			return &actionandtrigger.Output{Welcome: anyResult2}, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(action2))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	sdk.SubscribeToDonTrigger(
		runner,
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
		},
	)

	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, anyResult1+anyResult2, result)
}

func TestRuntime_NodeRuntimeUseInDonModeFails(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			assert.True(t, proto.Equal(anyConfig, config))
			return anyTrigger, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	capability := &basicactionmock.BasicActionCapability{
		PerformAction: func(_ context.Context, _ *basicaction.Inputs) (*basicaction.Outputs, error) {
			assert.Fail(t, "should not be called")
			return nil, errors.New("should not be called")
		},
	}

	require.NoError(t, reg.RegisterCapability(capability))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	var nrt sdk.NodeRuntime
	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, input *basictrigger.Outputs) (int32, error) {
			consensus := sdk.RunInNodeModeWithBuiltInConsensus(rt, func(nodeRuntime sdk.NodeRuntime) (int32, error) {
				nrt = nodeRuntime
				return 0, err
			}, pb.SimpleConsensusType_MEDIAN)

			_, err = consensus.Await()
			require.Error(t, err)

			return consensusResult, nil
		},
	)

	_, _, err = runner.Result()
	assert.Equal(t, sdk.NodeModeCallInDonMode, err)
}

func TestRuntime_DonRuntimeUseInNodeModeFails(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			assert.True(t, proto.Equal(anyConfig, config))
			return anyTrigger, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	capability := &basicactionmock.BasicActionCapability{
		PerformAction: func(_ context.Context, _ *basicaction.Inputs) (*basicaction.Outputs, error) {
			assert.Fail(t, "should not be called")
			return nil, errors.New("should not be called")
		},
	}

	require.NoError(t, reg.RegisterCapability(capability))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, input *basictrigger.Outputs) (int32, error) {
			consensus := sdk.RunInNodeModeWithBuiltInConsensus(rt, func(nodeRuntime sdk.NodeRuntime) (int32, error) {
				action := basicaction.BasicAction{}
				_, err := action.PerformAction(rt, &basicaction.Inputs{InputThing: true}).Await()
				return 0, err
			}, pb.SimpleConsensusType_MEDIAN)

			consensusResult, err := consensus.Await()
			require.NoError(t, err)
			return consensusResult, nil

		},
	)

	_, _, err = runner.Result()
	assert.Equal(t, sdk.DonModeCallInNodeMode, err)
}

func TestRuntime_ReturnsErrorsFromCapabilitiesThatDoNotExist(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			return &basictrigger.Outputs{CoolOutput: "cool"}, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, _ *basictrigger.Outputs) (string, error) {
			workflowAction1 := &basicaction.BasicAction{}
			call := workflowAction1.PerformAction(rt, &basicaction.Inputs{InputThing: true})
			_, err := call.Await()
			return "", err
		},
	)

	_, _, err = runner.Result()
	notRegistered := basicactionmock.BasicActionCapability{}
	require.Equal(t, testutils.NoCapability(notRegistered.ID()), err)
	assert.ErrorContains(t, err, "Capability not found")
	assert.ErrorContains(t, err, notRegistered.ID())
}

func TestRuntime_ConsensusReturnsTheObservation(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			assert.True(t, proto.Equal(anyConfig, config))
			return anyTrigger, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	anyValue := int32(100)
	nodeCapability := &nodeactionmock.BasicActionCapability{
		PerformAction: func(_ context.Context, _ *nodeaction.NodeInputs) (*nodeaction.NodeOutputs, error) {
			return &nodeaction.NodeOutputs{OutputThing: anyValue}, nil
		},
	}

	require.NoError(t, reg.RegisterCapability(nodeCapability))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, input *basictrigger.Outputs) (int32, error) {
			consensus := sdk.RunInNodeModeWithBuiltInConsensus(rt, func(nodeRuntime sdk.NodeRuntime) (int32, error) {
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
	)

	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, anyValue, result)
}

func TestRuntime_ConsensusReturnsTheDefaultValue(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			assert.True(t, proto.Equal(anyConfig, config))
			return anyTrigger, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	anyValue := int32(100)
	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, input *basictrigger.Outputs) (int32, error) {
			consensusType := &sdk.PrimitiveConsensusWithDefault[int32]{
				SimpleConsensusType: pb.SimpleConsensusType_MEDIAN,
				DefaultValue:        anyValue,
			}
			consensus := sdk.RunInNodeModeWithBuiltInConsensus(
				rt,
				func(nodeRuntime sdk.NodeRuntime) (int32, error) {
					return 0, errors.New("no consensus")
				},
				consensusType)

			consensusResult, err := consensus.Await()
			require.NoError(t, err)
			return consensusResult, nil
		},
	)

	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, anyValue, result)
}

func TestRuntime_ConsensusReturnsErrors(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			assert.True(t, proto.Equal(anyConfig, config))
			return anyTrigger, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	anyErr := errors.New("no consensus")
	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, input *basictrigger.Outputs) (int32, error) {
			consensus := sdk.RunInNodeModeWithBuiltInConsensus(
				rt,
				func(nodeRuntime sdk.NodeRuntime) (int32, error) {
					return 0, anyErr
				},
				pb.SimpleConsensusType_MEDIAN)

			return consensus.Await()
		},
	)

	_, _, err = runner.Result()
	require.Equal(t, err, anyErr)

}
