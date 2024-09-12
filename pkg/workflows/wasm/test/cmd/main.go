//go:build wasip1

package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/sdk"
)

func BuildWorkflow(config []byte) *workflows.WorkflowSpecFactory {
	workflow := workflows.NewWorkflowSpecFactory(
		workflows.NewWorkflowParams{
			Name:  "tester",
			Owner: "ryan",
		},
	)

	triggerCfg := basictrigger.TriggerConfig{Name: "trigger", Number: 100}
	_ = triggerCfg.New(workflow)

	return workflow
}

func main() {
	runner := sdk.NewRunner()
	workflow := BuildWorkflow(runner.Config())
	runner.Run(workflow)
}
