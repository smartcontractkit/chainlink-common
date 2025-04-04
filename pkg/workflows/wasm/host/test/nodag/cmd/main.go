/*//go:build wasip1*/

package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
)

func main() {
	runner := wasm.NewDonRunner()
	basic := &basictrigger.Basic{}
	sdk.SubscribeToDonTrigger(runner, basic.Trigger(&basictrigger.Config{
		Name:   "name",
		Number: 100,
	}), onTrigger)
}

func onTrigger(runtime sdk.DonRuntime, outputs *basictrigger.Outputs) (string, error) {
	action := basicaction.BasicAction{ /* TODO config */ }
	first := action.PerformAction(runtime, &basicaction.Inputs{InputThing: false})
	firstResult, err := first.Await()
	if err != nil {
		return "", err
	}

	second := action.PerformAction(runtime, &basicaction.Inputs{InputThing: true})
	secondResult, err := second.Await()
	if err != nil {
		return "", err
	}

	return outputs.CoolOutput + firstResult.AdaptedThing + secondResult.AdaptedThing, nil
}
