package cmd_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/arrayaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicconsensus"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictarget"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/externalreferenceaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/nestedaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/referenceaction"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

//go:generate go run github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/generate-types --dir $GOFILE

// Notes:
//
//	This doesn't get "code coverage" because the generate command is executed before the test
//	Since the builder isn't implemented yet, the types added aren't tested yet.
//	They will be tested here and in the tests for the builder as well.
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
				Name:   "anything",
				Number: 123,
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
}

// onlyVerifySyntax allows testing of the syntax while the builder still isn't implemented.
// The fact that the code compiles, verifies that the generated code works for typing.
func onlyVerifySyntax(run func()) {
	defer func() {
		_ = recover()
	}()
	run()
}
