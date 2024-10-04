//go:build wasip1

package main

import (
	"bytes"
	"crypto/rand"
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
)

func BuildWorkflow(config []byte) *sdk.WorkflowSpecFactory {
	workflow := sdk.NewWorkflowSpecFactory(
		sdk.NewWorkflowParams{},
	)

	triggerCfg := basictrigger.TriggerConfig{Name: "trigger", Number: 100}
	trigger := triggerCfg.New(workflow)

	sdk.Compute1[basictrigger.TriggerOutputs, bool](
		workflow,
		"transform",
		sdk.Compute1Inputs[basictrigger.TriggerOutputs]{Arg0: trigger},
		func(sdk sdk.Runtime, outputs basictrigger.TriggerOutputs) (bool, error) {
			b := make([]byte, 5)
			_, err := rand.Read(b)
			if err != nil {
				return false, err
			}

			// Compare to first 5 bytes of rand source seeded with 42.
			deterministic := bytes.Compare(b, []byte{0x53, 0x8c, 0x7f, 0x96, 0xb1}) == 0

			if !deterministic {
				return false, errors.New("expected deterministic output")
			}

			return deterministic, nil
		})

	return workflow
}

func main() {

	runner := wasm.NewRunner()
	workflow := BuildWorkflow(runner.Config())
	runner.Run(workflow)

}
