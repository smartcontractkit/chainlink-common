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
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

func TestRunner_TriggerFires(t *testing.T) {
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

	anyResult := "ok"
	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
			assert.True(t, proto.Equal(anyTrigger, input))
			return anyResult, nil
		},
	)

	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, anyResult, result)
}

func TestRunner_TriggerRegistrationCanBeVerifiedWithoutTriggering(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig1 := &basictrigger.Config{Name: "a", Number: 1}
	anyConfig2 := &actionandtrigger.Config{Name: "b"}

	trigger1 := &basictriggermock.BasicCapability{
		Trigger: func(ctx context.Context, input *basictrigger.Config) (*basictrigger.Outputs, error) {
			assert.True(t, proto.Equal(anyConfig1, input))
			return &basictrigger.Outputs{CoolOutput: "1"}, nil
		},
	}

	trigger2 := &actionandtriggermock.BasicCapability{
		Trigger: func(_ context.Context, input *actionandtrigger.Config) (*actionandtrigger.TriggerEvent, error) {
			assert.True(t, proto.Equal(anyConfig2, input))
			return nil, nil
		},
	}

	require.NoError(t, reg.RegisterCapability(trigger1))
	require.NoError(t, reg.RegisterCapability(trigger2))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	called := false
	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig1),
		func(rt sdk.DonRuntime, in *basictrigger.Outputs) (any, error) {
			called = true
			return nil, nil
		},
	)

	sdk.SubscribeToDonTrigger(
		runner,
		actionandtrigger.Basic{}.Trigger(anyConfig2),
		func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (*string, error) {
			assert.Fail(t, "trigger returned nil and shouldn't fire")
			return nil, nil
		},
	)

	ran, _, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.True(t, called)
}

func TestRunner_MissingTriggersAreNotRequired(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyConfig2 := &actionandtrigger.Config{Name: "b"}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			return anyTrigger, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	anyResult := "ok"
	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
			return anyResult, nil
		},
	)

	sdk.SubscribeToDonTrigger(
		runner,
		actionandtrigger.Basic{}.Trigger(anyConfig2),
		func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (*string, error) {
			assert.Fail(t, "This trigger shouldn't fire")
			return nil, nil
		},
	)

	_, _, err = runner.Result()
	require.NoError(t, err)
}

func TestRunner_FiringTwoTriggersReturnsAnError(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig1 := &basictrigger.Config{Name: "a", Number: 1}
	anyConfig2 := &actionandtrigger.Config{Name: "b"}

	trigger1 := &basictriggermock.BasicCapability{
		Trigger: func(ctx context.Context, input *basictrigger.Config) (*basictrigger.Outputs, error) {
			assert.True(t, proto.Equal(anyConfig1, input))
			return &basictrigger.Outputs{CoolOutput: "1"}, nil
		},
	}

	trigger2 := &actionandtriggermock.BasicCapability{
		Trigger: func(_ context.Context, input *actionandtrigger.Config) (*actionandtrigger.TriggerEvent, error) {
			assert.True(t, proto.Equal(anyConfig2, input))
			return &actionandtrigger.TriggerEvent{CoolOutput: "abcd"}, nil
		},
	}

	require.NoError(t, reg.RegisterCapability(trigger1))
	require.NoError(t, reg.RegisterCapability(trigger2))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	called := false
	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig1),
		func(rt sdk.DonRuntime, in *basictrigger.Outputs) (any, error) {
			called = true
			return nil, nil
		},
	)

	sdk.SubscribeToDonTrigger(
		runner,
		actionandtrigger.Basic{}.Trigger(anyConfig2),
		func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (*string, error) {
			assert.Fail(t, "second trigger shouldn't fire")
			return nil, nil
		},
	)

	ran, _, err := runner.Result()
	require.True(t, errors.Is(err, testutils.TooManyTriggers{}))
	assert.True(t, ran)
	assert.True(t, called)
}

func TestRunner_StrictTriggers_FailsIfTriggerIsNotRegistered(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyConfig2 := &actionandtrigger.Config{Name: "b"}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			return anyTrigger, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)
	runner.SetStrictTriggers(true)

	anyResult := "ok"
	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
			return anyResult, nil
		},
	)

	sdk.SubscribeToDonTrigger(
		runner,
		actionandtrigger.Basic{}.Trigger(anyConfig2),
		func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (*string, error) {
			assert.Fail(t, "This trigger shouldn't fire")
			return nil, nil
		},
	)

	_, _, err = runner.Result()
	missing := &actionandtriggermock.BasicCapability{}
	assert.True(t, errors.Is(err, testutils.NoCapability(missing.ID())))
}

func TestRunner_CallCapabilityIsAsync(t *testing.T) {
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
	assert.Fail(t, "Not written yet")
}

func TestRuntime_DonRuntimeUseInNodeModeFails(t *testing.T) {
	assert.Fail(t, "Not written yet")
}

func TestNodeRunner_UsesNodeRuntimeCapability(t *testing.T) {
	assert.Fail(t, "Not written yet")
}

func TestRunner_Logs(t *testing.T) {
	assert.Fail(t, "Not written yet")
}
