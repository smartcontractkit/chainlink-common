package main

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func BuildWorkflow(config []byte) (*sdk.WorkflowSpecFactory, error) {
	// Do something that errors
	return nil, errors.New("oops: I couldn't build this workflow")
}

func main() {
	runner := wasm.NewRunner()
	workflow, err := BuildWorkflow(runner.Config())
	if err != nil {
		runner.ExitWithError(err)
	}
	runner.Run(workflow)
}
