package testhelpers

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

func RunTestWorkflow(runner sdk.Runner[string]) {
	runner.Run(func(env *sdk.Environment[string]) (sdk.Workflow[string], error) {
		return sdk.Workflow[string]{
			sdk.Handler(
				basictrigger.Trigger(TestWorkflowTriggerConfig()),
				onTrigger),
		}, nil
	})
}

func RunIdenticalTriggersWorkflow(runner sdk.Runner[string]) {
	runner.Run(func(env *sdk.Environment[string]) (sdk.Workflow[string], error) {
		return sdk.Workflow[string]{
			sdk.Handler(
				basictrigger.Trigger(TestWorkflowTriggerConfig()),
				onTrigger,
			),
			sdk.Handler(
				basictrigger.Trigger(&basictrigger.Config{
					Name:   "second-trigger",
					Number: 200,
				}),
				func(env *sdk.Environment[string], rt sdk.Runtime, outputs *basictrigger.Outputs) (string, error) {
					res, err := onTrigger(env, rt, outputs)
					if err != nil {
						return "", err
					}
					return res + "true", nil
				},
			),
		}, nil
	})
}

func onTrigger(env *sdk.Environment[string], runtime sdk.Runtime, outputs *basictrigger.Outputs) (string, error) {
	env.Logger.Info("Hi")
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

func RunTestSecretsWorkflow(runner sdk.Runner[string]) {
	runner.Run(func(env *sdk.Environment[string]) (sdk.Workflow[string], error) {
		_, err := env.GetSecret(&pb.SecretRequest{Id: "Foo"}).Await()
		if err != nil {
			return nil, err
		}
		return sdk.Workflow[string]{
			sdk.Handler(
				basictrigger.Trigger(TestWorkflowTriggerConfig()),
				func(env *sdk.Environment[string], rt sdk.Runtime, outputs *basictrigger.Outputs) (string, error) {
					secret, err := env.GetSecret(&pb.SecretRequest{Id: "Foo"}).Await()
					if err != nil {
						return "", err
					}
					return secret.Value, nil
				}),
		}, nil
	})
}
