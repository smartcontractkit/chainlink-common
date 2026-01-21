package testutils_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicaction/basicactiontest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger/basictriggertest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/referenceaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/referenceaction/referenceactiontest"
	ocr3 "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap/ocr3captest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter/chainwritertest"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

func TestRunner(t *testing.T) {
	t.Parallel()
	t.Run("Run runs a workflow for a single execution", func(t *testing.T) {
		helper := &testHelper{t: t}
		workflow := createBasicTestWorkflow(helper.transformTrigger)

		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})

		triggerMock, actionMock, consensusMock, targetMock := setupAllRunnerMocks(t, runner)

		runner.Run(workflow)
		require.NoError(t, runner.Err())

		trigger := triggerMock.GetStep()
		assert.NoError(t, trigger.Error)
		assert.Equal(t, "cool", trigger.Output.CoolOutput)

		action := actionMock.GetStep("action")
		assert.True(t, action.WasRun)
		assert.Equal(t, basicaction.ActionInputs{InputThing: true}, action.Input)
		assert.NoError(t, action.Error)
		assert.Equal(t, "it was true", action.Output.AdaptedThing)

		assert.True(t, helper.transformTriggerCalled)
		consensus := consensusMock.GetStepDecoded("consensus")
		assert.Equal(t, "it was true", consensus.Output.AdaptedThing)
		require.NotNil(t, consensus.Input.Observations[0])

		rawConsensus := consensusMock.GetStep("consensus")
		target := targetMock.GetAllWrites()
		assert.Len(t, target.Errors, 0)
		assert.Len(t, target.Inputs, 1)
		assert.Equal(t, rawConsensus.Output, target.Inputs[0].SignedReport)
	})

	t.Run("Run allows hard-coded values", func(t *testing.T) {
		workflow := sdk.NewWorkflowSpecFactory()
		trigger := basictrigger.TriggerConfig{Name: "trigger", Number: 100}.New(workflow)
		hardCodedInput := basicaction.NewActionOutputsFromFields(sdk.ConstantDefinition("hard-coded"))
		tTransform := sdk.Compute2[basictrigger.TriggerOutputs, basicaction.ActionOutputs, bool](
			workflow,
			"transform",
			sdk.Compute2Inputs[basictrigger.TriggerOutputs, basicaction.ActionOutputs]{Arg0: trigger, Arg1: hardCodedInput},
			func(SDK sdk.Runtime, tr basictrigger.TriggerOutputs, hc basicaction.ActionOutputs) (bool, error) {
				assert.Equal(t, "hard-coded", hc.AdaptedThing)
				assert.NotNil(t, SDK)
				assert.Equal(t, "cool", tr.CoolOutput)
				return true, nil
			})

		action := basicaction.ActionConfig{CamelCaseInSchemaForTesting: "name", SnakeCaseInSchemaForTesting: 20}.
			New(workflow, "action", basicaction.ActionInput{InputThing: tTransform.Value()})

		consensus := ocr3.IdenticalConsensusConfig[basicaction.ActionOutputs]{
			Encoder:       "Test",
			EncoderConfig: ocr3.EncoderConfig{},
		}.New(workflow, "consensus", ocr3.IdenticalConsensusInput[basicaction.ActionOutputs]{Observation: action})

		chainwriter.TargetConfig{
			Address:    "0x123",
			DeltaStage: "2m",
			Schedule:   "oneAtATime",
		}.New(workflow, "chainwriter@1.0.0", chainwriter.TargetInput{SignedReport: consensus})

		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})
		_, _, _, targetMock := setupAllRunnerMocks(t, runner)

		runner.Run(workflow)
		require.NoError(t, runner.Err())
		target := targetMock.GetAllWrites()
		assert.Len(t, target.Inputs, 1)
	})

	t.Run("Run returns errors if capabilities were registered multiple times", func(t *testing.T) {
		helper := &testHelper{t: t}
		workflow := createBasicTestWorkflow(helper.transformTrigger)
		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})
		setupAllRunnerMocks(t, runner)
		setupAllRunnerMocks(t, runner)

		runner.Run(workflow)
		require.Error(t, runner.Err())
	})

	t.Run("Run captures errors", func(t *testing.T) {
		expectedErr := errors.New("nope")
		wf := createBasicTestWorkflow(func(SDK sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {
			return false, expectedErr
		})

		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})

		basictriggertest.Trigger(runner, func() (basictrigger.TriggerOutputs, error) {
			return basictrigger.TriggerOutputs{CoolOutput: "cool"}, nil
		})

		basicactiontest.Action(runner, func(input basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
			return basicaction.ActionOutputs{AdaptedThing: "it was true"}, nil
		})

		consensusMock := ocr3captest.IdenticalConsensus[basicaction.ActionOutputs](runner)

		runner.Run(wf)
		assert.True(t, errors.Is(runner.Err(), expectedErr))

		consensus := consensusMock.GetStep("consensus")
		assert.False(t, consensus.WasRun)
	})

	t.Run("Run fails if MockCapability is not provided for a step that is run", func(t *testing.T) {
		helper := &testHelper{t: t}
		workflow := createBasicTestWorkflow(helper.transformTrigger)

		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})

		basictriggertest.Trigger(runner, func() (basictrigger.TriggerOutputs, error) {
			return basictrigger.TriggerOutputs{CoolOutput: "cool"}, nil
		})

		ocr3captest.IdenticalConsensus[basicaction.ActionOutputs](runner)

		chainwritertest.Target(runner, "chainwriter@1.0.0", func(input chainwriter.TargetInputs) error {
			return nil
		})

		runner.Run(workflow)
		require.Error(t, runner.Err())
		require.Equal(t, "no mock found for capability basic-test-action@1.0.0 on step action", runner.Err().Error())
	})

	t.Run("Run registers and unregisters from capabilities", func(t *testing.T) {
		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})

		workflow, testTriggerConfig, testTargetConfig := registrationWorkflow()

		triggerMock := &mockRegistrationTester{t: t, expected: testTriggerConfig}
		executableMock := &mockRegistrationTester{t: t, expected: testTargetConfig}

		runner.MockCapability("trigger@1.0.0", nil, triggerMock)
		runner.MockCapability("target@1.0.0", nil, executableMock)

		runner.Run(workflow)

		assert.True(t, triggerMock.wasRegistered)
		assert.True(t, triggerMock.wasUnregistered)
		assert.True(t, executableMock.wasRegistered)
		assert.True(t, executableMock.wasUnregistered)
	})

	t.Run("Run captures register errors", func(t *testing.T) {
		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})

		workflow, _, _ := registrationWorkflow()

		triggerMock := &mockRegistrationTester{t: t, regErr: errors.New("foo")}
		executableMock := &mockRegistrationTester{t: t, regErr: errors.New("bar")}

		runner.MockCapability("trigger@1.0.0", nil, triggerMock)
		runner.MockCapability("target@1.0.0", nil, executableMock)

		runner.Run(workflow)

		actualErr := runner.Err()
		assert.True(t, errors.Is(actualErr, triggerMock.regErr))
		assert.True(t, errors.Is(actualErr, executableMock.regErr))
	})

	t.Run("Run captures unregister errors", func(t *testing.T) {
		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})

		workflow, _, _ := registrationWorkflow()

		triggerMock := &mockRegistrationTester{t: t, unregErr: errors.New("foo")}
		executableMock := &mockRegistrationTester{t: t, unregErr: errors.New("bar")}

		runner.MockCapability("trigger@1.0.0", nil, triggerMock)
		runner.MockCapability("target@1.0.0", nil, executableMock)

		runner.Run(workflow)

		actualErr := runner.Err()
		assert.True(t, errors.Is(actualErr, triggerMock.unregErr))
		assert.True(t, errors.Is(actualErr, executableMock.unregErr))
	})

	t.Run("GetRegisteredMock returns the mock for a step", func(t *testing.T) {
		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})
		expected := basicactiontest.ActionForStep(runner, "action", func(input basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
			return basicaction.ActionOutputs{}, nil
		})
		actual := runner.GetRegisteredMock(expected.ID(), "action")
		assert.Same(t, expected, actual)

		basicactiontest.Action(runner, func(input basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
			return basicaction.ActionOutputs{}, nil
		})
		actual = runner.GetRegisteredMock(expected.ID(), "action")
		assert.Same(t, expected, actual)
	})

	t.Run("GetRegisteredMock returns a default mock if step wasn't specified", func(t *testing.T) {
		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})
		expected := basicactiontest.Action(runner, func(input basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
			return basicaction.ActionOutputs{}, nil
		})
		actual := runner.GetRegisteredMock(expected.ID(), "action")
		assert.Same(t, expected, actual)
	})

	t.Run("GetRegisteredMock returns nil if no mock was registered", func(t *testing.T) {
		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})
		referenceactiontest.Action(runner, func(input referenceaction.SomeInputs) (referenceaction.SomeOutputs, error) {
			return referenceaction.SomeOutputs{}, nil
		})
		assert.Nil(t, runner.GetRegisteredMock("basic-test-action@1.0.0", "action"))
	})

	t.Run("GetRegisteredMock returns nil if no mock was registered for a step", func(t *testing.T) {
		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})
		differentStep := basicactiontest.ActionForStep(runner, "step", func(input basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
			return basicaction.ActionOutputs{}, nil
		})
		actual := runner.GetRegisteredMock(differentStep.ID(), "action")
		assert.Nil(t, actual)
	})
}

type ComputeConfig struct {
	Fidelity sdk.SecretValue
}

func TestCompute(t *testing.T) {
	t.Run("Inputs don't loose integer types when any is deserialized to", func(t *testing.T) {
		workflow := sdk.NewWorkflowSpecFactory()
		trigger := basictrigger.TriggerConfig{Name: "foo", Number: 100}.New(workflow)
		toMap := sdk.Compute1(workflow, "tomap", sdk.Compute1Inputs[string]{Arg0: trigger.CoolOutput()}, func(runtime sdk.Runtime, i0 string) (map[string]any, error) {
			v, err := strconv.Atoi(i0)
			if err != nil {
				return nil, err
			}

			return map[string]any{"a": int64(v)}, nil
		})

		sdk.Compute1(workflow, "compute", sdk.Compute1Inputs[map[string]any]{Arg0: toMap.Value()}, func(runtime sdk.Runtime, input map[string]any) (any, error) {
			actual := input["a"]
			if int64(100) != actual {
				return nil, fmt.Errorf("expected uint64(100), got %v of type %T", actual, actual)
			}

			return actual, nil
		})

		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})
		basictriggertest.Trigger(runner, func() (basictrigger.TriggerOutputs, error) {
			return basictrigger.TriggerOutputs{CoolOutput: "100"}, nil
		})

		runner.Run(workflow)

		require.NoError(t, runner.Err())
	})

	t.Run("Config interpolates secrets", func(t *testing.T) {
		workflow := sdk.NewWorkflowSpecFactory()
		trigger := basictrigger.TriggerConfig{Name: "foo", Number: 100}.New(workflow)

		conf := ComputeConfig{
			Fidelity: sdk.Secret("fidelity"),
		}
		var gotC ComputeConfig
		sdk.Compute1WithConfig(workflow, "tomap", &sdk.ComputeConfig[ComputeConfig]{Config: conf}, sdk.Compute1Inputs[string]{Arg0: trigger.CoolOutput()}, func(runtime sdk.Runtime, c ComputeConfig, i0 string) (ComputeConfig, error) {
			gotC = c
			return c, nil
		})

		runner := testutils.NewRunner(t.Context(), &testutils.NoopRuntime{})
		secretToken := "superSuperSecretToken"
		runner.Secrets = map[string]string{
			"fidelity": secretToken,
		}
		basictriggertest.Trigger(runner, func() (basictrigger.TriggerOutputs, error) {
			return basictrigger.TriggerOutputs{CoolOutput: "100"}, nil
		})

		runner.Run(workflow)

		require.NoError(t, runner.Err())
		assert.Equal(t, gotC.Fidelity, sdk.SecretValue(secretToken))
	})
}

func registrationWorkflow() (*sdk.WorkflowSpecFactory, map[string]any, map[string]any) {
	workflow := sdk.NewWorkflowSpecFactory()
	testTriggerConfig := map[string]any{"something": "from nothing"}
	trigger := sdk.Step[int]{
		Definition: sdk.StepDefinition{
			ID:             "trigger@1.0.0",
			Ref:            "trigger",
			Inputs:         sdk.StepInputs{},
			Config:         testTriggerConfig,
			CapabilityType: capabilities.CapabilityTypeTrigger,
		},
	}
	trigger.AddTo(workflow)

	testTargetConfig := map[string]any{"nothing": "from something"}
	target := sdk.Step[int]{
		Definition: sdk.StepDefinition{
			ID:             "target@1.0.0",
			Ref:            "target",
			Inputs:         sdk.StepInputs{Mapping: map[string]any{"foo": "$(trigger.outputs)"}},
			Config:         testTargetConfig,
			CapabilityType: capabilities.CapabilityTypeTarget,
		},
	}
	target.AddTo(workflow)
	return workflow, testTriggerConfig, testTargetConfig
}

func setupAllRunnerMocks(t *testing.T, runner *testutils.Runner) (*testutils.TriggerMock[basictrigger.TriggerOutputs], *testutils.Mock[basicaction.ActionInputs, basicaction.ActionOutputs], *ocr3captest.IdenticalConsensusMock[basicaction.ActionOutputs], *testutils.TargetMock[chainwriter.TargetInputs]) {
	triggerMock := basictriggertest.Trigger(runner, func() (basictrigger.TriggerOutputs, error) {
		return basictrigger.TriggerOutputs{CoolOutput: "cool"}, nil
	})

	actionMock := basicactiontest.Action(runner, func(input basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
		assert.True(t, input.InputThing)
		return basicaction.ActionOutputs{AdaptedThing: "it was true"}, nil
	})

	consensusMock := ocr3captest.IdenticalConsensus[basicaction.ActionOutputs](runner)

	targetMock := chainwritertest.Target(runner, "chainwriter@1.0.0", func(input chainwriter.TargetInputs) error {
		return nil
	})
	return triggerMock, actionMock, consensusMock, targetMock
}

type actionTransform func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error)

func createBasicTestWorkflow(actionTransform actionTransform) *sdk.WorkflowSpecFactory {
	workflow := sdk.NewWorkflowSpecFactory()
	trigger := basictrigger.TriggerConfig{Name: "trigger", Number: 100}.New(workflow)
	tTransform := sdk.Compute1[basictrigger.TriggerOutputs, bool](
		workflow,
		"transform",
		sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
		actionTransform)

	action := basicaction.ActionConfig{CamelCaseInSchemaForTesting: "name", SnakeCaseInSchemaForTesting: 20}.
		New(workflow, "action", basicaction.ActionInput{InputThing: tTransform.Value()})

	consensus := ocr3.IdenticalConsensusConfig[basicaction.ActionOutputs]{
		Encoder: "Test", EncoderConfig: ocr3.EncoderConfig{},
	}.New(workflow, "consensus", ocr3.IdenticalConsensusInput[basicaction.ActionOutputs]{Observation: action})

	chainwriter.TargetConfig{
		Address:    "0x123",
		DeltaStage: "2m",
		Schedule:   "oneAtATime",
	}.New(workflow, "chainwriter@1.0.0", chainwriter.TargetInput{SignedReport: consensus})

	return workflow
}

type testHelper struct {
	t                      *testing.T
	transformTriggerCalled bool
}

func (helper *testHelper) transformTrigger(runtime sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {
	assert.NotNil(helper.t, runtime)
	assert.Equal(helper.t, "cool", outputs.CoolOutput)
	assert.False(helper.t, helper.transformTriggerCalled)
	helper.transformTriggerCalled = true
	return true, nil
}

type mockRegistrationTester struct {
	t               *testing.T
	wasRegistered   bool
	wasUnregistered bool
	regErr          error
	unregErr        error
	expected        map[string]any
}

var _ capabilities.TriggerCapability = &mockRegistrationTester{}
var _ capabilities.ExecutableCapability = &mockRegistrationTester{}

func (m *mockRegistrationTester) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	return m.register(request.Config)
}

func (m *mockRegistrationTester) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	return m.unregister(request.Config)
}

func (m *mockRegistrationTester) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	assert.False(m.t, m.wasUnregistered)
	val, err := values.NewMap(map[string]any{"foo": "bar"})
	require.NoError(m.t, err)
	return capabilities.CapabilityResponse{Value: val}, nil
}

func (m *mockRegistrationTester) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	return capabilities.CapabilityInfo{}, nil
}

func (m *mockRegistrationTester) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	ch := make(chan capabilities.TriggerResponse)
	ch <- capabilities.TriggerResponse{Event: capabilities.TriggerEvent{
		TriggerType: "test",
		ID:          "test@1.0.0",
		Outputs:     &values.Map{Underlying: map[string]values.Value{"foo": values.NewString("bar")}},
	}}

	return ch, m.register(request.Config)
}

func (m *mockRegistrationTester) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	return m.unregister(request.Config)
}

func (m *mockRegistrationTester) AckEvent(ctx context.Context, triggerId string, eventId string, workflowId string) error {
	return nil
}

func (m *mockRegistrationTester) register(wrappedConfig *values.Map) error {
	m.wasRegistered = true
	m.verifyConfig(wrappedConfig)
	return m.regErr
}

func (m *mockRegistrationTester) unregister(wrappedConfig *values.Map) error {
	m.wasUnregistered = true
	m.verifyConfig(wrappedConfig)
	return m.unregErr
}

func (m *mockRegistrationTester) verifyConfig(wrappedConfig *values.Map) {
	if m.expected == nil {
		return
	}

	var actual map[string]any
	err := wrappedConfig.UnwrapTo(&actual)
	require.NoError(m.t, err)
	assert.True(m.t, reflect.DeepEqual(m.expected, actual))
}
