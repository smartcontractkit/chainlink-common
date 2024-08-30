package testutils_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicaction/basicactiontest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger/basictriggertest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/referenceaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/referenceaction/referenceactiontest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap/ocr3captest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter/chainwritertest"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

func TestRunner(t *testing.T) {
	t.Parallel()
	t.Run("Run runs a workflow for a single execution", func(t *testing.T) {
		helper := &testHelper{t: t}
		workflow := createBasicTestWorkflow(helper.transformTrigger)

		runner := testutils.NewRunner()

		triggerMock, actionMock, consensusMock, targetMock := setupAllRunnerMocks(t, runner)

		require.NoError(t, runner.Run(workflow))

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
		require.Len(t, consensus.Input.Observations, 1)

		rawConsensus := consensusMock.GetStep("consensus")
		target := targetMock.GetAllWrites()
		assert.Len(t, target.Errors, 0)
		assert.Len(t, target.Inputs, 1)
		assert.Equal(t, rawConsensus.Output, target.Inputs[0].SignedReport)
	})

	t.Run("Run allows hard-coded values", func(t *testing.T) {
		workflow := sdk.NewWorkflowSpecFactory(sdk.NewWorkflowParams{Name: "tester", Owner: "ryan"})
		trigger := basictrigger.TriggerConfig{Name: "trigger", Number: 100}.New(workflow)
		hardCodedInput := basicaction.NewActionOutputsFromFields(sdk.ConstantDefinition("hard-coded"))
		tTransform := sdk.Compute2[basictrigger.TriggerOutputs, basicaction.ActionOutputs, bool](
			workflow,
			"transform",
			sdk.Compute2Inputs[basictrigger.TriggerOutputs, basicaction.ActionOutputs]{Arg0: trigger, Arg1: hardCodedInput},
			func(sdk sdk.Runtime, tr basictrigger.TriggerOutputs, hc basicaction.ActionOutputs) (bool, error) {
				assert.Equal(t, "hard-coded", hc.AdaptedThing)
				assert.NotNil(t, sdk)
				assert.Equal(t, "cool", tr.CoolOutput)
				return true, nil
			})

		action := basicaction.ActionConfig{CamelCaseInSchemaForTesting: "name", SnakeCaseInSchemaForTesting: 20}.
			New(workflow, "action", basicaction.ActionInput{InputThing: tTransform.Value()})

		consensus := ocr3cap.IdenticalConsensusConfig[basicaction.ActionOutputs]{
			Encoder:       "Test",
			EncoderConfig: ocr3cap.EncoderConfig{Abi: "Test"},
		}.New(workflow, "consensus", ocr3cap.IdenticalConsensusInput[basicaction.ActionOutputs]{Observations: action})

		chainwriter.TargetConfig{
			Address:    "0x123",
			DeltaStage: "2m",
			Schedule:   "oneAtATime",
		}.New(workflow, "chainwriter@1.0.0", chainwriter.TargetInput{SignedReport: consensus})

		runner := testutils.NewRunner()
		_, _, _, targetMock := setupAllRunnerMocks(t, runner)

		require.NoError(t, runner.Run(workflow))
		target := targetMock.GetAllWrites()
		assert.Len(t, target.Inputs, 1)
	})

	t.Run("Run returns errors if capabilities were registered multiple times", func(t *testing.T) {
		helper := &testHelper{t: t}
		workflow := createBasicTestWorkflow(helper.transformTrigger)
		runner := testutils.NewRunner()
		setupAllRunnerMocks(t, runner)
		setupAllRunnerMocks(t, runner)

		require.Error(t, runner.Run(workflow))
	})

	t.Run("Run captures errors", func(t *testing.T) {
		expectedErr := errors.New("nope")
		wf := createBasicTestWorkflow(func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {
			return false, expectedErr
		})

		runner := testutils.NewRunner()

		basictriggertest.Trigger(runner, func() (basictrigger.TriggerOutputs, error) {
			return basictrigger.TriggerOutputs{CoolOutput: "cool"}, nil
		})

		basicactiontest.Action(runner, func(input basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
			return basicaction.ActionOutputs{AdaptedThing: "it was true"}, nil
		})

		consensusMock := ocr3captest.IdenticalConsensus[basicaction.ActionOutputs](runner)

		err := runner.Run(wf)
		assert.True(t, errors.Is(err, expectedErr))

		consensus := consensusMock.GetStep("consensus")
		assert.False(t, consensus.WasRun)
	})

	t.Run("Run fails if MockCapability is not provided for a step that is run", func(t *testing.T) {
		helper := &testHelper{t: t}
		workflow := createBasicTestWorkflow(helper.transformTrigger)

		runner := testutils.NewRunner()

		basictriggertest.Trigger(runner, func() (basictrigger.TriggerOutputs, error) {
			return basictrigger.TriggerOutputs{CoolOutput: "cool"}, nil
		})

		ocr3captest.IdenticalConsensus[basicaction.ActionOutputs](runner)

		chainwritertest.Target(runner, "chainwriter@1.0.0", func(input chainwriter.TargetInputs) error {
			return nil
		})

		err := runner.Run(workflow)
		require.Error(t, err)
		require.Equal(t, "no mock found for capability basic-test-action@1.0.0 on step action", err.Error())
	})

	t.Run("Fails build if workflow spec generation fails", func(t *testing.T) {
		t.Skip("TODO https://smartcontract-it.atlassian.net/browse/KS-442")
		assert.Fail(t, "Not implemented")
	})

	t.Run("Fails build if not all leafs are targets", func(t *testing.T) {
		t.Skip("TODO https://smartcontract-it.atlassian.net/browse/KS-442")
		assert.Fail(t, "Not implemented")
	})

	t.Run("GetRegisteredMock returns the mock for a step", func(t *testing.T) {
		runner := testutils.NewRunner()
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
		runner := testutils.NewRunner()
		expected := basicactiontest.Action(runner, func(input basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
			return basicaction.ActionOutputs{}, nil
		})
		actual := runner.GetRegisteredMock(expected.ID(), "action")
		assert.Same(t, expected, actual)
	})

	t.Run("GetRegisteredMock returns nil if no mock was registered", func(t *testing.T) {
		runner := testutils.NewRunner()
		referenceactiontest.Action(runner, func(input referenceaction.SomeInputs) (referenceaction.SomeOutputs, error) {
			return referenceaction.SomeOutputs{}, nil
		})
		assert.Nil(t, runner.GetRegisteredMock("basic-test-action@1.0.0", "action"))
	})

	t.Run("GetRegisteredMock returns nil if no mock was registered for a step", func(t *testing.T) {
		runner := testutils.NewRunner()
		differentStep := basicactiontest.ActionForStep(runner, "step", func(input basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
			return basicaction.ActionOutputs{}, nil
		})
		actual := runner.GetRegisteredMock(differentStep.ID(), "action")
		assert.Nil(t, actual)
	})
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
	workflow := sdk.NewWorkflowSpecFactory(sdk.NewWorkflowParams{Name: "tester", Owner: "ryan"})
	trigger := basictrigger.TriggerConfig{Name: "trigger", Number: 100}.New(workflow)
	tTransform := sdk.Compute1[basictrigger.TriggerOutputs, bool](
		workflow,
		"transform",
		sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
		actionTransform)

	action := basicaction.ActionConfig{CamelCaseInSchemaForTesting: "name", SnakeCaseInSchemaForTesting: 20}.
		New(workflow, "action", basicaction.ActionInput{InputThing: tTransform.Value()})

	consensus := ocr3cap.IdenticalConsensusConfig[basicaction.ActionOutputs]{
		Encoder:       "Test",
		EncoderConfig: ocr3cap.EncoderConfig{Abi: "Test"},
	}.New(workflow, "consensus", ocr3cap.IdenticalConsensusInput[basicaction.ActionOutputs]{Observations: action})

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

func (helper *testHelper) transformTrigger(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {
	assert.NotNil(helper.t, sdk)
	assert.Equal(helper.t, "cool", outputs.CoolOutput)
	assert.False(helper.t, helper.transformTriggerCalled)
	helper.transformTriggerCalled = true
	return true, nil
}