// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package externalreferenceactiontest

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/referenceaction"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/legacy/testutils"
)

// Action registers a new capability mock with the runner
// if another mock is registered for the same capability with for a step, it will take priority for that step.
func Action(runner *testutils.Runner, fn func(input referenceaction.SomeInputs) (referenceaction.SomeOutputs, error)) *testutils.Mock[referenceaction.SomeInputs, referenceaction.SomeOutputs] {
	mock := testutils.MockCapability[referenceaction.SomeInputs, referenceaction.SomeOutputs]("external-reference-test-action@1.0.0", fn)
	runner.MockCapability("external-reference-test-action@1.0.0", nil, mock)
	return mock
}

// ActionForStep registers a new capability mock with the runner, but only for a given step.
// if another mock was registered for the same capability without a step, this mock will take priority for that step.
func ActionForStep(runner *testutils.Runner, step string, mockFn func(input referenceaction.SomeInputs) (referenceaction.SomeOutputs, error)) *testutils.Mock[referenceaction.SomeInputs, referenceaction.SomeOutputs] {
	fn := mockFn
	mock := testutils.MockCapability[referenceaction.SomeInputs, referenceaction.SomeOutputs]("external-reference-test-action@1.0.0", fn)
	runner.MockCapability("external-reference-test-action@1.0.0", &step, mock)
	return mock
}
