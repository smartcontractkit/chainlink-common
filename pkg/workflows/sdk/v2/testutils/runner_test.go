package testutils_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testhelpers/v2"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/actionandtrigger"
	actionandtriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/actionandtrigger/action_and_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	basictriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger/basic_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
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

	runner := testutils.NewRunner(t, "unused")
	require.NoError(t, err)

	anyResult := "ok"

	workflows := sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig),
			func(_ *sdk.Environment[string], rt sdk.Runtime, input *basictrigger.Outputs) (string, error) {
				assert.True(t, proto.Equal(anyTrigger, input))
				return anyResult, nil
			},
		),
	}
	runWorkflows(runner, workflows)

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

	runner := testutils.NewRunner(t, "unused")
	require.NoError(t, err)

	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig),
			func(_ *sdk.Environment[string], rt sdk.Runtime, input *basictrigger.Outputs) (*bad, error) {
				return &bad{C: make(chan int, 1)}, nil
			},
		),
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

	runner := testutils.NewRunner(t, "unused")
	require.NoError(t, err)

	called := false
	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig1),
			func(_ *sdk.Environment[string], rt sdk.Runtime, in *basictrigger.Outputs) (struct{}, error) {
				called = true
				return struct{}{}, nil
			},
		),
		sdk.On(
			actionandtrigger.Trigger(anyConfig2),
			func(_ *sdk.Environment[string], rt sdk.Runtime, in *actionandtrigger.TriggerEvent) (*string, error) {
				assert.Fail(t, "trigger returned nil and shouldn't fire")
				return nil, nil
			},
		),
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

	runner := testutils.NewRunner(t, "unused")
	require.NoError(t, err)

	anyResult := "ok"
	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig),
			func(_ *sdk.Environment[string], rt sdk.Runtime, input *basictrigger.Outputs) (string, error) {
				return anyResult, nil
			},
		),
		sdk.On(
			actionandtrigger.Trigger(anyConfig2),
			func(_ *sdk.Environment[string], rt sdk.Runtime, in *actionandtrigger.TriggerEvent) (*string, error) {
				assert.Fail(t, "This trigger shouldn't fire")
				return nil, nil
			},
		),
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

	runner := testutils.NewRunner(t, "unused")
	require.NoError(t, err)

	anyResult := "ok"
	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig),
			func(_ *sdk.Environment[string], rt sdk.Runtime, input *basictrigger.Outputs) (string, error) {
				return anyResult, nil
			},
		),
		sdk.On(
			actionandtrigger.Trigger(anyConfig2),
			func(_ *sdk.Environment[string], rt sdk.Runtime, in *actionandtrigger.TriggerEvent) (*string, error) {
				assert.Fail(t, "This trigger shouldn't fire")
				return nil, nil
			},
		),
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

	runner := testutils.NewRunner(t, "unused")
	require.NoError(t, err)

	called := false
	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig1),
			func(_ *sdk.Environment[string], rt sdk.Runtime, in *basictrigger.Outputs) (any, error) {
				called = true
				return nil, nil
			},
		),
		sdk.On(
			actionandtrigger.Trigger(anyConfig2),
			func(_ *sdk.Environment[string], rt sdk.Runtime, in *actionandtrigger.TriggerEvent) (*string, error) {
				assert.Fail(t, "second trigger shouldn'tb fire")
				return nil, nil
			},
		),
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

	runner := testutils.NewRunner(t, "unused")
	require.NoError(t, err)
	runner.SetStrictTriggers(true)

	anyResult := "ok"
	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig),
			func(_ *sdk.Environment[string], rt sdk.Runtime, input *basictrigger.Outputs) (string, error) {
				return anyResult, nil
			},
		),
		sdk.On(
			actionandtrigger.Trigger(anyConfig2),
			func(_ *sdk.Environment[string], rt sdk.Runtime, in *actionandtrigger.TriggerEvent) (*string, error) {
				assert.Fail(t, "This trigger shouldn'tb fire")
				return nil, nil
			},
		),
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

	runner := testutils.NewRunner(t, "unused")
	require.NoError(t, err)
	runner.SetStrictTriggers(true)

	anyResult := "ok"
	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig),
			func(_ *sdk.Environment[string], rt sdk.Runtime, input *basictrigger.Outputs) (string, error) {
				return anyResult, nil
			},
		),
		sdk.On(
			actionandtrigger.Trigger(anyConfig2),
			func(_ *sdk.Environment[string], rt sdk.Runtime, in *actionandtrigger.TriggerEvent) (*string, error) {
				assert.Fail(t, "This trigger shouldn't fire")
				return nil, nil
			},
		),
	})

	_, _, err = runner.Result()
	assert.Error(t, err)
}

func TestRunner_Logs(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		return anyTrigger, nil
	}

	runner := testutils.NewRunner(t, "unused")
	require.NoError(t, err)

	anyResult := "ok"
	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig),
			func(wcx *sdk.Environment[string], _ sdk.Runtime, input *basictrigger.Outputs) (string, error) {
				logger := wcx.Logger
				logger.Info(anyResult)
				logger.Warn(anyResult + "2")
				return anyResult, nil
			},
		),
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

func TestRunner_SecretsProvider(t *testing.T) {
	anyConfig := &basictrigger.Config{Name: "name", Number: 123}
	anyTrigger := &basictrigger.Outputs{CoolOutput: "cool"}

	trigger, err := basictriggermock.NewBasicCapability(t)
	require.NoError(t, err)
	trigger.Trigger = func(_ context.Context, config *basictrigger.Config) (*basictrigger.Outputs, error) {
		return anyTrigger, nil
	}

	runner := testutils.NewRunner(t, "unused")

	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig),
			func(wcx *sdk.Environment[string], _ sdk.Runtime, input *basictrigger.Outputs) (string, error) {
				secret, err := wcx.GetSecret(&pb.SecretRequest{Id: "Foo"}).Await()
				if err != nil {
					return "", err
				}
				return secret.Value, nil
			},
		),
	})

	_, _, err = runner.Result()
	assert.ErrorContains(t, err, "could not find secret /Foo")

	runner = testutils.NewRunner(t, "unused")

	expectedSecret := "bar"
	runner.SetSecret("", "Foo", expectedSecret)

	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig),
			func(wcx *sdk.Environment[string], _ sdk.Runtime, input *basictrigger.Outputs) (string, error) {
				secret, err := wcx.GetSecret(&pb.SecretRequest{Id: "Foo"}).Await()
				if err != nil {
					return "", err
				}

				assert.Equal(t, expectedSecret, secret.Value)
				return secret.Value, nil
			},
		),
	})

	_, _, err = runner.Result()
	assert.NoError(t, err)
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

	runner := testutils.NewRunner(t, "unused")
	require.NoError(t, err)

	runWorkflows(runner, sdk.Workflow[string]{
		sdk.On(
			basictrigger.Trigger(anyConfig),
			func(_ *sdk.Environment[string], rt sdk.Runtime, input *basictrigger.Outputs) (string, error) {
				assert.Fail(t, "This trigger shouldn't fire as there is already an error")
				return "", nil
			},
		),
		sdk.On(
			actionandtrigger.Trigger(&actionandtrigger.Config{Name: "b"}),
			func(_ *sdk.Environment[string], rt sdk.Runtime, in *actionandtrigger.TriggerEvent) (string, error) {
				assert.Fail(t, "This trigger should not fire")
				return "", nil
			},
		),
	})

	_, _, err = runner.Result()
	assert.Equal(t, anyError, err)
}

func TestRunner_FullWorkflow(t *testing.T) {
	testhelpers.SetupExpectedCalls(t)
	runner := testutils.NewRunner(t, "unused")
	testhelpers.RunTestWorkflow(runner)
	ran, result, err := runner.Result()
	require.NoError(t, err)
	assert.True(t, ran)
	assert.Equal(t, testhelpers.TestWorkflowExpectedResult(), result)
	logs := runner.Logs()
	require.Len(t, logs, 1)
	assert.True(t, strings.Contains(logs[0], "Hi"))
}

func runWorkflows(runner sdk.Runner[string], workflows sdk.Workflow[string]) {
	runner.Run(func(wcx *sdk.Environment[string]) (sdk.Workflow[string], error) {
		return workflows, nil
	})
}
