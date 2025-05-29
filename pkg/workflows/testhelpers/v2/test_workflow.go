package testhelpers

import (
	"context"
	"log/slog"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

func RunTestWorkflow(runner sdk.DonRunner) {
	basic := &basictrigger.Basic{}
	slog.Log(context.Background(), slog.LevelInfo, "Hi")
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basic.Trigger(TestWorkflowTriggerConfig()),
				onTrigger),
		},
	})
}

func RunIdenticalTriggersWorkflow(runner sdk.DonRunner) {
	basic := &basictrigger.Basic{}
	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewDonHandler(
				basic.Trigger(TestWorkflowTriggerConfig()),
				onTrigger,
			),
			sdk.NewDonHandler(
				basic.Trigger(&basictrigger.Config{
					Name:   "second-trigger",
					Number: 200,
				}),
				func(rt sdk.DonRuntime, outputs *basictrigger.Outputs) (string, error) {
					res, err := onTrigger(rt, outputs)
					if err != nil {
						return "", err
					}
					return res + "true", nil
				},
			),
		},
	})
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
