package testutils_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testhelpers/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/actionandtrigger"
	actionandtriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/actionandtrigger/action_and_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	basictriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger/basic_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodetrigger"
	nodetriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodetrigger/node_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils"
)

func TestRunner_TriggerFires(t *testing.T) {
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

	anyResult := "ok"
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
					assert.True(t, proto.Equal(anyTrigger, input))
					return anyResult, nil
				},
			)},
	})

	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, anyResult, result)
}

func TestRunner_HasErrorsWhenReturnCannotMarshal(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}
	type bad struct {
		C chan int
	}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig, config))
		return anyTrigger, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (*bad, error) {
					return &bad{C: make(chan int, 1)}, nil
				},
			)},
	})

	_, _, err = runner.Result()
	require.ErrorContains(t, err, "could not wrap")
}

func TestRunner_TriggerRegistrationCanBeVerifiedWithoutTriggering(t *testing.T) {
	anyConfig1 := &basictrigger.Config{Name: "a", Number: 1}
	anyConfig2 := &actionandtrigger.Config{Name: "b"}

	trigger1, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger1.Trigger = func(ctx context.Context, input *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig1, input))
		return &basictrigger.Outputs{CoolOutput: "1"}, nil
	}

	trigger2, err := actionandtriggermock.NewBasicCapability(t)
	trigger2.Trigger = func(_ context.Context, input *actionandtrigger.Config) (*actionandtrigger.TriggerEvent, error) {
		assert.True(t, proto.Equal(anyConfig2, input))
		return nil, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	called := false
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig1),
				func(rt sdk.DonRuntime, in *basictrigger.Outputs) (any, error) {
					called = true
					return nil, nil
				},
			),
			sdk.NewDonHandler(
				actionandtrigger.Basic{}.Trigger(anyConfig2),
				func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (*string, error) {
					assert.Fail(t, "trigger returned nil and shouldn'tb fire")
					return nil, nil
				},
			),
		},
	})

	ran, _, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.True(t, called)
}

func TestRunner_MissingTriggersAreNotRequired(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyConfig2 := &actionandtrigger.Config{Name: "b"}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		return anyTrigger, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	anyResult := "ok"
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
					return anyResult, nil
				},
			),
			sdk.NewDonHandler(
				actionandtrigger.Basic{}.Trigger(anyConfig2),
				func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (*string, error) {
					assert.Fail(t, "This trigger shouldn'tb fire")
					return nil, nil
				},
			),
		},
	})

	_, _, err = runner.Result()
	require.NoError(t, err)
}

func TestRunner_MissingTriggerStubsAreNotRequired(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyConfig2 := &actionandtrigger.Config{Name: "b"}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		return anyTrigger, nil
	}

	_, err = actionandtriggermock.NewBasicCapability(t)
	require.NoError(t, err)

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	anyResult := "ok"
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
					return anyResult, nil
				},
			),
			sdk.NewDonHandler(
				actionandtrigger.Basic{}.Trigger(anyConfig2),
				func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (*string, error) {
					assert.Fail(t, "This trigger shouldn'tb fire")
					return nil, nil
				},
			),
		},
	})

	_, _, err = runner.Result()
	require.NoError(t, err)
}

func TestRunner_FiringTwoTriggersReturnsAnError(t *testing.T) {
	anyConfig1 := &basictrigger.Config{Name: "a", Number: 1}
	anyConfig2 := &actionandtrigger.Config{Name: "b"}

	trigger1, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger1.Trigger = func(ctx context.Context, input *basictrigger.Config) (*basictrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig1, input))
		return &basictrigger.Outputs{CoolOutput: "1"}, nil
	}

	trigger2, err := actionandtriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger2.Trigger = func(_ context.Context, input *actionandtrigger.Config) (*actionandtrigger.TriggerEvent, error) {
		assert.True(t, proto.Equal(anyConfig2, input))
		return &actionandtrigger.TriggerEvent{CoolOutput: "abcd"}, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	called := false
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig1),
				func(rt sdk.DonRuntime, in *basictrigger.Outputs) (any, error) {
					called = true
					return nil, nil
				},
			),
			sdk.NewDonHandler(
				actionandtrigger.Basic{}.Trigger(anyConfig2),
				func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (*string, error) {
					assert.Fail(t, "second trigger shouldn'tb fire")
					return nil, nil
				},
			),
		},
	})

	ran, _, err := runner.Result()
	require.True(t, errors.Is(err, testutils.TooManyTriggers{}))
	assert.True(t, strings.Contains(err.Error(), "too many triggers fired during execution"))
	assert.True(t, ran)
	assert.True(t, called)
}

func TestRunner_StrictTriggers_FailsIfTriggerIsNotRegistered(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyConfig2 := &actionandtrigger.Config{Name: "b"}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		return anyTrigger, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)
	runner.SetStrictTriggers(true)

	anyResult := "ok"
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
					return anyResult, nil
				},
			),
			sdk.NewDonHandler(
				actionandtrigger.Basic{}.Trigger(anyConfig2),
				func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (*string, error) {
					assert.Fail(t, "This trigger shouldn'tb fire")
					return nil, nil
				},
			),
		},
	})

	_, _, err = runner.Result()
	assert.Error(t, err)
}

func TestRunner_StrictTriggers_FailsIfTriggerIsNotStubbed(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyConfig2 := &actionandtrigger.Config{Name: "b"}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		return anyTrigger, nil
	}

	_, err = actionandtriggermock.NewBasicCapability(t)
	require.NoError(t, err)

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)
	runner.SetStrictTriggers(true)

	anyResult := "ok"
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
					return anyResult, nil
				},
			),
			sdk.NewDonHandler(
				actionandtrigger.Basic{}.Trigger(anyConfig2),
				func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (*string, error) {
					assert.Fail(t, "This trigger shouldn'tb fire")
					return nil, nil
				},
			),
		},
	})

	_, _, err = runner.Result()
	assert.Error(t, err)
}

func TestRunner_CanStartInNodeMode(t *testing.T) {
	anyConfig := &nodetrigger.Config{Name: "name", Number: 123}
	anyTrigger := &nodetrigger.Outputs{CoolOutput: "cool"}

	trigger, err := nodetriggermock.NewNodeEventCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *nodetrigger.Config) (*nodetrigger.Outputs, error) {
		assert.True(t, proto.Equal(anyConfig, config))
		return anyTrigger, nil
	}

	runner := testutils.NewNodeRunner(t, nil)
	require.NoError(t, err)

	anyResult := "ok"
	runner.Run(&sdk.WorkflowArgs[sdk.NodeRuntime]{
		Handlers: []sdk.Handler[sdk.NodeRuntime]{
			sdk.NewNodeHandler(
				nodetrigger.NodeEvent{}.Trigger(anyConfig),
				func(rt sdk.NodeRuntime, input *nodetrigger.Outputs) (string, error) {
					assert.True(t, proto.Equal(anyTrigger, input))
					return anyResult, nil
				},
			)},
	})

	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, anyResult, result)
}

func TestRunner_Logs(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		return anyTrigger, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	runner.SetDefaultLogger()

	anyResult := "ok"
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
					logger := slog.Default()
					logger.Info(anyResult)
					logger.Warn(anyResult + "2")
					return anyResult, nil
				},
			)},
	})

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
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyError := errors.New("some error")

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		return nil, anyError
	}

	trigger2, err := actionandtriggermock.NewBasicCapability(t)
	trigger2.Trigger = func(ctx context.Context, input *actionandtrigger.Config) (*actionandtrigger.TriggerEvent, error) {
		assert.Fail(t, "workflow should halt if a trigger has an error")
		return nil, nil
	}

	runner := testutils.NewDonRunner(t, nil)
	require.NoError(t, err)

	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basictrigger.Basic{}.Trigger(anyConfig),
				func(rt sdk.DonRuntime, input *basictrigger.Outputs) (string, error) {
					assert.Fail(t, "This trigger shouldn'tb fire as there is already an error")
					return "", nil
				},
			),
			sdk.NewDonHandler(
				actionandtrigger.Basic{}.Trigger(&actionandtrigger.Config{Name: "b"}),
				func(rt sdk.DonRuntime, in *actionandtrigger.TriggerEvent) (string, error) {
					assert.Fail(t, "This trigger should not fire")
					return "", nil
				}),
		},
	})

	_, _, err = runner.Result()
	assert.Equal(t, anyError, err)
}

func TestRunner_FullWorkflow(t *testing.T) {
	testhelpers.SetupExpectedCalls(t)
	runner := testutils.NewDonRunner(t, nil)
	runner.SetDefaultLogger()
	testhelpers.RunTestWorkflow(runner)
	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, testhelpers.TestWorkflowExpectedResult(), result)
	logs := runner.Logs()
	assert.Len(t, logs, 1)
	assert.True(t, strings.Contains(logs[0], "Hi"))
}
