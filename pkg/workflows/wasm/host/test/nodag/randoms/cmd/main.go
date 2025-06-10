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
	runner := wasm.NewDonRunner()
	basic := &basictrigger.Basic{}

	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basic.Trigger(testhelpers.TestWorkflowTriggerConfig()),
				func(runtime sdk.DonRuntime, trigger *basictrigger.Outputs) (uint64, error) {
					r, err := runtime.Rand()
					if err != nil {
						return 0, err
					}
					total := r.Uint64()
					sdk.RunInNodeMode[uint64](runtime, func(nrt sdk.NodeRuntime) (uint64, error) {
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
							runtime.LogWriter().Write([]byte(strconv.FormatUint(nr.Uint64(), 10)))
						}
						return 0, nil
					}, sdk.ConsensusIdenticalAggregation[uint64]())
					total += r.Uint64()
					return total, nil
				}),
		},
	})
}
