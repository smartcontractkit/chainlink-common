package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type foo struct {
	thing string
}

func (f *foo) doAThing() {
	_ = f.thing
}

func BuildWorkflow(config []byte) *sdk.WorkflowSpecFactory {
	workflow := sdk.NewWorkflowSpecFactory()

	triggerCfg := basictrigger.TriggerConfig{Name: "trigger", Number: 100}
	trigger := triggerCfg.New(workflow)

	sdk.Compute1[basictrigger.TriggerOutputs, bool](
		workflow,
		"transform",
		sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
		func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {
			var f *foo
			f.doAThing()
			return false, nil
		})

	return workflow
}

func main() {
	runner := wasm.NewRunner()
	workflow := BuildWorkflow(runner.Config())
	runner.Run(workflow)
}
