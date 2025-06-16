//go:build wasip1

package main

import (
	"strconv"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testhelpers/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2"
)

func main() {
	runner := wasm.NewRunner(func(configBytes []byte) (string, error) {
		return string(configBytes), nil
	})

	basic := &basictrigger.Basic{}

	runner.Run(func(_ *sdk.WorkflowContext[string]) (sdk.Workflow[string], error) {
		return sdk.Workflow[string]{
			sdk.On(
				basic.Trigger(testhelpers.TestWorkflowTriggerConfig()),
				func(wcx *sdk.WorkflowContext[string], runtime sdk.Runtime, payload *basictrigger.Outputs) (uint64, error) {
					r, err := runtime.Rand()
					if err != nil {
						return 0, err
					}
					total := r.Uint64()
					sdk.RunInNodeMode(wcx, runtime, func(wcx *sdk.WorkflowContext[string], nrt sdk.NodeRuntime) (uint64, error) {
						node, err := (&nodeaction.BasicAction{}).PerformAction(nrt, &nodeaction.NodeInputs{
							InputThing: false,
						}).Await()

						if err != nil {
							return 0, err
						}

						// Conditionally generate a random number based on the node output.
						// This ensures it doesn't impact the next DON mode number.
						if node.OutputThing < 100 {
							nr, err := nrt.Rand()
							if err != nil {
								return 0, err
							}
							_, _ = wcx.LogWriter.Write([]byte(strconv.FormatUint(nr.Uint64(), 10)))
						}
						return 0, nil
					}, sdk.ConsensusIdenticalAggregation[uint64]())
					total += r.Uint64()
					return total, nil
				},
			),
		}, nil
	})
}
