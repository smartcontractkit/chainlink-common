//go:build wasip1

package main

import (
	"net/http"

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
	resp := runner.SDK.Fetch(wasm.TargetRequestPayload{
		Method: http.MethodGet,
		URL:    "https://min-api.cryptocompare.com/data/pricemultifull?fsyms=ETH&tsyms=BTC",
	})

	runner.SDK.Logger.Infow("fetch response", "body", string(resp.Body))
	workflow := BuildWorkflow(runner.Config())
	runner.Run(workflow)
}
