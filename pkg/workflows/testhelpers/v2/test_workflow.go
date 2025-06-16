package testhelpers

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

func RunTestWorkflow(runner sdk.Runner[string]) {
	basic := &basictrigger.Basic{}
	runner.Run(func(wcx *sdk.WorkflowContext[string]) (sdk.Workflow[string], error) {
		return sdk.Workflow[string]{
			sdk.On(
				basic.Trigger(TestWorkflowTriggerConfig()),
				onTrigger),
		}, nil
	})
}

func RunIdenticalTriggersWorkflow(runner sdk.Runner[string]) {
	basic := &basictrigger.Basic{}
	runner.Run(func(wcx *sdk.WorkflowContext[string]) (sdk.Workflow[string], error) {
		return sdk.Workflow[string]{
			sdk.On(
				basic.Trigger(TestWorkflowTriggerConfig()),
				onTrigger,
			),
			sdk.On(
				basic.Trigger(&basictrigger.Config{
					Name:   "second-trigger",
					Number: 200,
				}),
				func(wcx *sdk.WorkflowContext[string], rt sdk.Runtime, outputs *basictrigger.Outputs) (string, error) {
					res, err := onTrigger(wcx, rt, outputs)
					if err != nil {
						return "", err
					}
					return res + "true", nil
				},
			),
		}, nil
	})
}

func onTrigger(wcx *sdk.WorkflowContext[string], runtime sdk.Runtime, outputs *basictrigger.Outputs) (string, error) {
	wcx.Logger.Info("Hi")
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
