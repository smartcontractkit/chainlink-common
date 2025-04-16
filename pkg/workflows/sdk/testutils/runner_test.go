package testutils_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/actionandtrigger"
	actionandtriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/actionandtrigger/action_and_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basictrigger"
	basictriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basictrigger/basic_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/nodetrigger"
	nodetriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/nodetrigger/node_triggermock"
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
	assert.True(t, strings.Contains(err.Error(), "too many triggers fired during execution"))
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

func TestRunner_CanStartInNodeMode(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &nodetrigger.Config{Name: "name", Number: 123}
	anyTrigger := &nodetrigger.Outputs{CoolOutput: "cool"}

	trigger := &nodetriggermock.NodeEventCapability{
		Trigger: func(_ context.Context, config *nodetrigger.Config) (*nodetrigger.Outputs, error) {
			assert.True(t, proto.Equal(anyConfig, config))
			return anyTrigger, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	runner, err := testutils.NewNodeRunner(ctx, nil, reg)
	require.NoError(t, err)

	anyResult := "ok"
	sdk.SubscribeToNodeTrigger(
		runner,
		nodetrigger.NodeEvent{}.Trigger(anyConfig),
		func(rt sdk.NodeRuntime, input *nodetrigger.Outputs) (string, error) {
			assert.True(t, proto.Equal(anyTrigger, input))
			return anyResult, nil
		},
	)

	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, anyResult, result)
}

func TestRunner_Logs(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			return anyTrigger, nil
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	runner.SetDefaultLogger()

	anyResult := "ok"
	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
			logger := slog.Default()
			logger.Info(anyResult)
			logger.Warn(anyResult + "2")
			return anyResult, nil
		},
	)

	_, _, err = runner.Result()
	require.NoError(t, err)

	expected := []string{
		"level=INFO msg=ok\n",
		"level=WARN msg=ok2\n",
	}

	var actual []string
	for _, log := range runner.Logs() {
		// Extract only the level and msg fields
		parts := strings.Split(log, " ")
		var filtered []string
		for _, part := range parts {
			if strings.HasPrefix(part, "level=") || strings.HasPrefix(part, "msg=") {
				filtered = append(filtered, part)
			}
		}
		actual = append(actual, strings.Join(filtered, " "))
	}

	assert.Equal(t, expected, actual)

}

func TestRunner_ReturnsTriggerErrorsWithoutRunningTheWorkflow(t *testing.T) {
	ctx := context.Background()
	reg := &testutils.Registry{}

	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyError := errors.New("some error")

	trigger := &basictriggermock.BasicCapability{
		Trigger: func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
			return nil, anyError
		},
	}
	require.NoError(t, reg.RegisterCapability(trigger))

	trigger2 := &actionandtriggermock.BasicCapability{
		Trigger: func(ctx context.Context, input *actionandtrigger.Config) (*actionandtrigger.TriggerEvent, error) {
			assert.Fail(t, "workflow should halt if a trigger has an error")
			return nil, nil
		},
	}

	require.NoError(t, reg.RegisterCapability(trigger2))

	runner, err := testutils.NewDonRunner(ctx, nil, reg)
	require.NoError(t, err)

	sdk.SubscribeToDonTrigger(
		runner,
		basictrigger.Basic{}.Trigger(anyConfig),
		func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
			assert.Fail(t, "This trigger shouldn't fire as there is already an error")
			return "", nil
		},
	)

	sdk.SubscribeToDonTrigger(
		runner,
		actionandtrigger.Basic{}.Trigger(&actionandtrigger.Config{Name: "b"}),
		func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (string, error) {
			assert.Fail(t, "This trigger shouldn't fire")
			return "", nil
		})

	_, _, err = runner.Result()
	assert.Equal(t, anyError, err)
}
