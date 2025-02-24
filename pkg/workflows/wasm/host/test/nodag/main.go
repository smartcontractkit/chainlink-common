package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictarget"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
)

func main() {
	runner := wasm.NewRunnerV2()
	triggerCfg := basictrigger.TriggerConfig{Number: 100}
	_ = wasm.SubscribeToTrigger(runner, "basic-trigger@1.0.0", triggerCfg, OnBasicTriggerEvent)
	runner.Run()
}

func OnBasicTriggerEvent(runtime sdk.RuntimeV2, triggerOutputs basictrigger.TriggerOutputs) error {
	// two async capability calls
	actionCall1, _ := wasm.CallCapability[basicaction.ActionInputs, basicaction.ActionConfig, basicaction.ActionOutputs](
		runtime, "basicaction@1.0.0", basicaction.ActionInputs{}, basicaction.ActionConfig{},
	)
	actionCall2, _ := wasm.CallCapability[basicaction.ActionInputs, basicaction.ActionConfig, basicaction.ActionOutputs](
		runtime, "basicaction@1.0.0", basicaction.ActionInputs{}, basicaction.ActionConfig{},
	)

	// blocking "await" on multiple calls at once
	err := runtime.AwaitCapabilities(actionCall1, actionCall2)
	if err != nil {
		return err
	}

	// some compute before calling a target
	actionOutputs1, _ := actionCall1.Result()
	actionOutputs2, _ := actionCall2.Result()
	if len(actionOutputs1.AdaptedThing) <= len(actionOutputs2.AdaptedThing) {
		// a single target call
		inputStr := "abcd"
		targetCall, _ := wasm.CallCapability[basictarget.TargetInputs, basictarget.TargetConfig, any](
			runtime, "basictarget@1.0.0", basictarget.TargetInputs{CoolInput: &inputStr}, basictarget.TargetConfig{},
		)
		return runtime.AwaitCapabilities(targetCall)
	}
	return nil
}
