// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package mapactiontest

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/mapaction"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

// Action registers a new capability mock with the runner
// if another mock is registered for the same capability with for a step, it will take priority for that step.
func Action(runner *testutils.Runner, fn func(input mapaction.ActionInputs) (mapaction.ActionOutputs, error)) *testutils.Mock[mapaction.ActionInputs, mapaction.ActionOutputs] {
	mock := testutils.MockCapability[mapaction.ActionInputs, mapaction.ActionOutputs]("mapaction@1.0.0", fn)
	runner.MockCapability("mapaction@1.0.0", nil, mock)
	return mock
}

// ActionForStep registers a new capability mock with the runner, but only for a given step.
// if another mock was registered for the same capability without a step, this mock will take priority for that step.
func ActionForStep(runner *testutils.Runner, step string, mockFn func(input mapaction.ActionInputs) (mapaction.ActionOutputs, error)) *testutils.Mock[mapaction.ActionInputs, mapaction.ActionOutputs] {
	fn := mockFn
	mock := testutils.MockCapability[mapaction.ActionInputs, mapaction.ActionOutputs]("mapaction@1.0.0", fn)
	runner.MockCapability("mapaction@1.0.0", &step, mock)
	return mock
}