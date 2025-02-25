package main

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictarget"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
)

func main() {
	runner := wasm.NewDonRunner()
	err := basictrigger.Subscribe(runner, &basictrigger.TriggerConfig{Number: 100}, OnBasicTriggerEvent)
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
	runner.Run()
}

func OnBasicTriggerEvent(runtime sdk.DonRuntime, triggerOutputs *basictrigger.TriggerOutputs) (struct{}, error) {
	// two async capability calls
	capability := &basicaction.Basic{Config: basicaction.ActionConfig{}}
	actionCall1 := capability.Call(runtime, &basicaction.ActionInput{})
	actionCall2 := capability.Call(runtime, &basicaction.ActionInput{})

	// TODO in examples use one at a time
	_ = runtime.AwaitCapabilities(actionCall1, actionCall2)

	actionOutputs1, _ := actionCall1.Await()
	actionOutputs2, _ := actionCall2.Await()
	if len(actionOutputs1.AdaptedThing) <= len(actionOutputs2.AdaptedThing) {
		fmt.Println("smaller")
		// a single target call
		inputStr := "abcd"
		target := basictarget.BasicTarget{Config: basictarget.TargetConfig{}}
		w, err := target.Write(runtime, &basictarget.TargetInputs{CoolInput: &inputStr}).Await()
		fmt.Printf("done waiting w: %v, err: %v\n", w, err)
		return w, err
	}
	fmt.Println("Done")

	return struct{}{}, nil
}
