//go:build wasip1

package main

import (
	"net/http"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func BuildWorkflow(config []byte) *sdk.WorkflowSpecFactory {
	workflow := sdk.NewWorkflowSpecFactory()

	triggerCfg := basictrigger.TriggerConfig{Name: "trigger", Number: 100}
	trigger := triggerCfg.New(workflow)

	sdk.Compute1[basictrigger.TriggerOutputs, bool](
		workflow,
		"transform",
		sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
		func(rsdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {

			for i := 0; i < 6; i++ {
				_, err := rsdk.Fetch(sdk.FetchRequest{
					Method: http.MethodGet,
					URL:    "https://min-api.cryptocompare.com/data/pricemultifull?fsyms=ETH&tsyms=BTC",
				})
				if err != nil {
					return false, err
				}
			}

			return true, nil
		})

	return workflow
}
func main() {
	runner := wasm.NewRunner()
	workflow := BuildWorkflow(runner.Config())
	runner.Run(workflow)
}
