package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/arrayaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/arrayaction/arrayactiontest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicaction/basicactiontest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicconsensus"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictarget"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictarget/basictargettest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger/basictriggertest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/externalreferenceaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/externalreferenceaction/externalreferenceactiontest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/nestedaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/referenceaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/referenceaction/referenceactiontest"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testutils"
)

//go:generate go run github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/generate-types --dir $GOFILE

// Notes:
//
//	This doesn't get "code coverage" because the generate command is executed before the test
//	These tests only verify syntax to assure use is as intended, the test for the `workflows.WorkflowSpecFactory` and `testutils.Runner`
//	test their interactions with those components.  This is done to avoid duplication in testing effort
//	and allows specific testing of what interfaces should be implemented by generated code.
func TestTypeGeneration(t *testing.T) {
	t.Run("Basic trigger", func(t *testing.T) {
		onlyVerifySyntax(func() {
			factory := &workflows.WorkflowSpecFactory{}

			// assure the right type of trigger
			var trigger basictrigger.TriggerOutputsCap //nolint
			trigger = basictrigger.TriggerConfig{
				Name:   "anything",
				Number: 123,
			}.New(factory)

			// verify that the underlying interface is right
			var _ workflows.CapDefinition[basictrigger.TriggerOutputs] = trigger

			// verify the type is correct
			var expectedOutput workflows.CapDefinition[string] //nolint
			expectedOutput = trigger.CoolOutput()
			_ = expectedOutput
		})
	})

	t.Run("Basic action", func(t *testing.T) {
		onlyVerifySyntax(func() {
			factory := &workflows.WorkflowSpecFactory{}

			// assure the right type of action
			var action basicaction.ActionOutputsCap //nolint
			action = basicaction.ActionConfig{
				CamelCaseInSchemaForTesting: "anything",
				SnakeCaseInSchemaForTesting: 123,
			}.New(factory, "reference", basicaction.ActionInput{
				InputThing: workflows.CapDefinition[bool](nil),
			})

			// verify that the underlying interface is right
			var _ workflows.CapDefinition[basicaction.ActionOutputs] = action

			// verify the type is correct
			var expectedOutput workflows.CapDefinition[string] //nolint
			expectedOutput = action.AdaptedThing()
			_ = expectedOutput
		})
	})

	t.Run("Basic consensus", func(t *testing.T) {
		onlyVerifySyntax(func() {
			factory := &workflows.WorkflowSpecFactory{}

			// assure the right type of trigger
			var consensus basicconsensus.ConsensusOutputsCap //nolint
			consensus = basicconsensus.ConsensusConfig{
				Name:   "anything",
				Number: 123,
			}.New(factory, "reference", basicconsensus.ConsensusInput{
				InputThing: workflows.CapDefinition[bool](nil),
			})

			// verify that the underlying interface is right
			var _ workflows.CapDefinition[basicconsensus.ConsensusOutputs] = consensus

			// verify the type is correct
			var expectedConsensusField workflows.CapDefinition[[]string] //nolint
			expectedConsensusField = consensus.Consensus()
			_ = expectedConsensusField

			var expectedSigsField workflows.CapDefinition[[]string] //nolint
			expectedSigsField = consensus.Sigs()
			_ = expectedSigsField
		})
	})

	t.Run("Basic target", func(t *testing.T) {
		onlyVerifySyntax(func() {
			config := basictarget.TargetConfig{
				Name:   "anything",
				Number: 123,
			}

			// verify no output type
			var verifyCreationType func(w *workflows.WorkflowSpecFactory, input basictarget.TargetInput) //nolint
			verifyCreationType = config.New
			var _ = verifyCreationType
		})
	})
	t.Run("References", func(t *testing.T) {
		onlyVerifySyntax(func() {
			factory := &workflows.WorkflowSpecFactory{}

			// assure the right type of action
			var action referenceaction.SomeOutputsCap //nolint
			action = referenceaction.SomeConfig{
				Name:   "anything",
				Number: 123,
			}.New(factory, "reference", referenceaction.ActionInput{
				InputThing: workflows.CapDefinition[bool](nil),
			})

			// verify that the underlying interface is right
			var _ workflows.CapDefinition[referenceaction.SomeOutputs] = action

			// verify the type is correct
			var expectedOutput workflows.CapDefinition[string] //nolint
			expectedOutput = action.AdaptedThing()
			_ = expectedOutput
		})
	})

	t.Run("External references", func(t *testing.T) {
		onlyVerifySyntax(func() {
			factory := &workflows.WorkflowSpecFactory{}

			// assure the right type of action
			var trigger referenceaction.SomeOutputsCap
			config := externalreferenceaction.SomeConfig{
				Name:   "anything",
				Number: 123,
			}
			trigger = config.New(factory, "reference", referenceaction.ActionInput{
				InputThing: workflows.CapDefinition[bool](nil),
			})
			_ = trigger

			// verify that the type can be cast
			cast := referenceaction.SomeConfig(config)
			_ = cast
		})
	})

	t.Run("Nested types work", func(t *testing.T) {
		onlyVerifySyntax(func() {
			factory := &workflows.WorkflowSpecFactory{}

			// assure the right type of action
			var action nestedaction.ActionOutputsCap //nolint
			action = nestedaction.ActionConfig{
				Details: nestedaction.ActionConfigDetails{
					Name:   "anything",
					Number: 123,
				},
			}.New(factory, "reference", nestedaction.ActionInput{
				Metadata: workflows.CapDefinition[nestedaction.ActionInputsMetadata](nil),
			})

			// verify that the underlying interface is right
			var _ workflows.CapDefinition[nestedaction.ActionOutputs] = action

			// verify the types are correct
			var expectedOutput nestedaction.ActionOutputsResultsCap
			var expectedOutputRaw workflows.CapDefinition[nestedaction.ActionOutputsResults]
			expectedOutput = action.Results()
			expectedOutputRaw = expectedOutput
			_ = expectedOutputRaw

			var expectedUnderlyingFieldType workflows.CapDefinition[string] //nolint
			expectedUnderlyingFieldType = expectedOutput.AdaptedThing()
			_ = expectedUnderlyingFieldType
		})
	})

	t.Run("Array types work", func(t *testing.T) {
		onlyVerifySyntax(func() {
			factory := &workflows.WorkflowSpecFactory{}

			// assure the right type of action
			var action workflows.CapDefinition[[]arrayaction.ActionOutputsElem] //nolint
			action = arrayaction.ActionConfig{
				Details: arrayaction.ActionConfigDetails{
					Name:   "name",
					Number: 123,
				},
			}.New(factory, "reference", arrayaction.ActionInput{
				Metadata: workflows.CapDefinition[arrayaction.ActionInputsMetadata](nil),
			})
			_ = action
		})
	})

	t.Run("Creating a type from fields works", func(t *testing.T) {
		onlyVerifySyntax(func() {
			factory := &workflows.WorkflowSpecFactory{}
			var action referenceaction.SomeOutputsCap //nolint
			action = referenceaction.SomeConfig{
				Name:   "anything",
				Number: 123,
			}.New(factory, "reference", referenceaction.ActionInput{
				InputThing: workflows.CapDefinition[bool](nil),
			})

			// verify the type is correct
			var adapted basicaction.ActionOutputsCap //nolint
			adapted = basicaction.NewActionOutputsFromFields(action.AdaptedThing())
			_ = adapted
		})
	})

	t.Run("casing is respected from the json schema", func(t *testing.T) {
		workflow := workflows.NewWorkflowSpecFactory(workflows.NewWorkflowParams{Owner: "owner", Name: "name"})
		ai := basicaction.ActionConfig{CamelCaseInSchemaForTesting: "foo", SnakeCaseInSchemaForTesting: 12}.
			New(workflow, "ref", basicaction.ActionInput{InputThing: workflows.ConstantDefinition[bool](true)})
		spec, _ := workflow.Spec()
		require.Len(t, spec.Actions, 1)
		actual := spec.Actions[0]
		require.Equal(t, 12, actual.Config["snake_case_in_schema_for_testing"])
		require.Equal(t, "foo", actual.Config["camelCaseInSchemaForTesting"])
		require.True(t, actual.Inputs.Mapping["input_thing"].(bool))
		require.Equal(t, "$(ref.outputs.adapted_thing)", ai.AdaptedThing().Ref())
	})
}

func TestMockGeneration(t *testing.T) {
	t.Run("Basic trigger", func(t *testing.T) {
		runner := testutils.NewRunner()
		capMock := basictriggertest.Trigger(runner, func() (basictrigger.TriggerOutputs, error) {
			return basictrigger.TriggerOutputs{}, nil
		})

		// verify type is correct
		var mock *testutils.TriggerMock[basictrigger.TriggerOutputs] //nolint
		mock = capMock
		_ = mock
	})

	t.Run("Basic action", func(t *testing.T) {
		runner := testutils.NewRunner()

		// nolint value is never used but it's assigned to mock to verify the type
		capMock := basicactiontest.Action(runner, func(_ basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
			return basicaction.ActionOutputs{}, nil
		})

		specificMock := basicactiontest.ActionForStep(runner, "step", func(_ basicaction.ActionInputs) (basicaction.ActionOutputs, error) {
			return basicaction.ActionOutputs{}, nil
		})

		// verify type is correct
		var mock *testutils.Mock[basicaction.ActionInputs, basicaction.ActionOutputs] //nolint
		// nolint
		mock = capMock
		mock = specificMock
		_ = mock
	})

	t.Run("Basic target", func(t *testing.T) {
		runner := testutils.NewRunner()
		capMock := basictargettest.Target(runner, func(_ basictarget.TargetInputs) error {
			return nil
		})

		// verify type is correct
		var mock *testutils.TargetMock[basictarget.TargetInputs] //nolint
		mock = capMock
		_ = mock
	})

	t.Run("References", func(t *testing.T) {
		runner := testutils.NewRunner()

		// nolint value is never used but it's assigned to mock to verify the type
		capMock := referenceactiontest.Action(runner, func(_ referenceaction.SomeInputs) (referenceaction.SomeOutputs, error) {
			return referenceaction.SomeOutputs{}, nil
		})

		specificMock := referenceactiontest.ActionForStep(runner, "step", func(_ referenceaction.SomeInputs) (referenceaction.SomeOutputs, error) {
			return referenceaction.SomeOutputs{}, nil
		})

		// verify type is correct
		var mock *testutils.Mock[referenceaction.SomeInputs, referenceaction.SomeOutputs] //nolint
		// nolint
		mock = capMock
		mock = specificMock
		_ = mock
	})

	t.Run("External references", func(t *testing.T) {
		runner := testutils.NewRunner()

		// nolint value is never used but it's assigned to mock to verify the type
		capMock := externalreferenceactiontest.Action(runner, func(_ referenceaction.SomeInputs) (referenceaction.SomeOutputs, error) {
			return referenceaction.SomeOutputs{}, nil
		})

		specificMock := externalreferenceactiontest.ActionForStep(runner, "step", func(_ referenceaction.SomeInputs) (referenceaction.SomeOutputs, error) {
			return referenceaction.SomeOutputs{}, nil
		})

		// verify type is correct
		var mock *testutils.Mock[referenceaction.SomeInputs, referenceaction.SomeOutputs] //nolint

		// nolint ineffectual assignment is ok, it's for testing the type.
		mock = capMock
		mock = specificMock
		_ = mock
	})

	// no need to test nesting, we don't generate anything different for the mock's

	t.Run("Array action", func(t *testing.T) {
		runner := testutils.NewRunner()
		// nolint value is never used but it's assigned to mock to verify the type
		capMock := arrayactiontest.Action(runner, func(_ arrayaction.ActionInputs) ([]arrayaction.ActionOutputsElem, error) {
			return []arrayaction.ActionOutputsElem{}, nil
		})

		specificMock := arrayactiontest.ActionForStep(runner, "step", func(_ arrayaction.ActionInputs) ([]arrayaction.ActionOutputsElem, error) {
			return []arrayaction.ActionOutputsElem{}, nil
		})

		// verify type is correct
		var mock *testutils.Mock[arrayaction.ActionInputs, []arrayaction.ActionOutputsElem] //nolint

		// nolint ineffectual assignment is ok, it's for testing the type.
		mock = capMock
		mock = specificMock
		_ = mock
	})
}

// onlyVerifySyntax allows testing of the syntax while the builder still isn't implemented.
// The fact that the code compiles, verifies that the generated code works for typing.
func onlyVerifySyntax(_ func()) {}