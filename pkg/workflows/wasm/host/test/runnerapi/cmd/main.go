package main

import (
	wasm "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/legacy"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	sdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/legacy"
)

func BuildWorkflow(config []byte) *sdk.WorkflowSpecFactory {
	workflow := sdk.NewWorkflowSpecFactory()

	triggerCfg := basictrigger.TriggerConfig{Name: "trigger", Number: 100}
	_ = triggerCfg.New(workflow)

	return workflow
}

func main() {
	runner := wasm.NewRunner()
	workflow := BuildWorkflow(runner.Config())
	runner.Run(workflow)
}
