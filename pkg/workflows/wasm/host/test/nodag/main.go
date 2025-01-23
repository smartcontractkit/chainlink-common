//go:build wasip1

package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictarget"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
)

func InitWorkflow(config []byte) *sdk.WorkflowSpecFactory {
	workflow := sdk.NewWorkflowSpecFactory()
	// add triggers
	triggerCfg := basictrigger.TriggerConfig{Name: "trigger", Number: 100}
	_ = triggerCfg.New(workflow) // TODO: let AddRunFunctionForTrigger() add this to workflow spec
	sdk.AddRunFunctionForTrigger(workflow, "trigger", triggerCfg, RunWorkflow)
	return workflow
}

func RunWorkflow(runtime sdk.RuntimeV2, triggerOutputs basictrigger.TriggerOutputs) error {
	// two action calls "futures"
	actionCall1, _ := wasm.NewCapabilityCall[basicaction.ActionInputs, basicaction.ActionConfig, basicaction.ActionOutputs](
		"ref_action1", "basicaction@1.0.0", basicaction.ActionInputs{}, basicaction.ActionConfig{},
	)
	actionCall2, _ := wasm.NewCapabilityCall[basicaction.ActionInputs, basicaction.ActionConfig, basicaction.ActionOutputs](
		"ref_action2", "basicaction@1.0.0", basicaction.ActionInputs{}, basicaction.ActionConfig{},
	)

	// blocking "await" on multiple calls at once
	err := runtime.CallCapabilities(actionCall1, actionCall2)
	if err != nil {
		return err
	}

	// some compute before calling a target
	actionOutputs1, _ := actionCall1.Result()
	actionOutputs2, _ := actionCall2.Result()
	if len(actionOutputs1.AdaptedThing) <= len(actionOutputs2.AdaptedThing) {
		// a single target call
		inputStr := "abcd"
		targetCall, _ := wasm.NewCapabilityCall[basictarget.TargetInputs, basictarget.TargetConfig, any](
			"ref_target1", "basictarget@1.0.0", basictarget.TargetInputs{CoolInput: &inputStr}, basictarget.TargetConfig{},
		)
		return runtime.CallCapabilities(targetCall)
	}
	return nil
}

func main() {
	runner := wasm.NewRunnerV2()
	workflow := InitWorkflow(runner.Config())
	runner.Run(workflow)
}
