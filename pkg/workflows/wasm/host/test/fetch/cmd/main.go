//go:build wasip1

package main

import (
	"net/http"

	wasm "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/legacy"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	sdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/legacy"
)

func BuildWorkflow(config []byte) *sdk.WorkflowSpecFactory {
	workflow := sdk.NewWorkflowSpecFactory()

	triggerCfg := basictrigger.TriggerConfig{Name: "trigger", Number: 100}
	trigger := triggerCfg.New(workflow)

	sdk.Compute1[basictrigger.TriggerOutputs, sdk.FetchResponse](
		workflow,
		"transform",
		sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
		func(rsdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (sdk.FetchResponse, error) {
			return rsdk.Fetch(sdk.FetchRequest{
				Method: http.MethodGet,
				URL:    "https://min-api.cryptocompare.com/data/pricemultifull?fsyms=ETH&tsyms=BTC",
			})
		})

	return workflow
}
func main() {
	runner := wasm.NewRunner()
	workflow := BuildWorkflow(runner.Config())
	runner.Run(workflow)
}
