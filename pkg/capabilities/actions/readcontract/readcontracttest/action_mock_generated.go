// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package readcontracttest

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/actions/readcontract"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

// Action registers a new capability mock with the runner
// if another mock is registered for the same capability with for a step, it will take priority for that step.
func Action(runner *testutils.Runner, id string, fn func(input readcontract.Input) (readcontract.Output, error)) *testutils.Mock[readcontract.Input, readcontract.Output] {
	mock := testutils.MockCapability[readcontract.Input, readcontract.Output](id, fn)
	runner.MockCapability(id, nil, mock)
	return mock
}

// ActionForStep registers a new capability mock with the runner, but only for a given step.
// if another mock was registered for the same capability without a step, this mock will take priority for that step.
func ActionForStep(runner *testutils.Runner, id string, step string, mockFn func(input readcontract.Input) (readcontract.Output, error)) *testutils.Mock[readcontract.Input, readcontract.Output] {
	fn := mockFn
	mock := testutils.MockCapability[readcontract.Input, readcontract.Output](id, fn)
	runner.MockCapability(id, &step, mock)
	return mock
}
