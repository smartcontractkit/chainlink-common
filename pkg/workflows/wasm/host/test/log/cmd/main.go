//go:build wasip1

package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func BuildWorkflow(config []byte) *sdk.WorkflowSpecFactory {
	workflow := sdk.NewWorkflowSpecFactory(
		sdk.NewWorkflowParams{
			Name:  "tester",
			Owner: "ryan",
		},
	)

	triggerCfg := basictrigger.TriggerConfig{Name: "trigger", Number: 100}
	_ = triggerCfg.New(workflow)

	return workflow
}

func main() {
	runner := wasm.NewRunner()
	runner.SDK.Logger.Infow("building workflow...", []interface{}{
		"test-string-field-key", "this is a test field content",
		"test-numeric-field-key", 6400000,
	}...)
	workflow := BuildWorkflow(runner.Config())
	runner.SDK.Logger.Info("running workflow...")
	runner.Run(workflow)
}
