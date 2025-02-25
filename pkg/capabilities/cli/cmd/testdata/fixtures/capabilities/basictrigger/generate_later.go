package basictrigger

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func Subscribe[T any](runner sdk.DonRunner, config *TriggerConfig, handler func(sdk.DonRuntime, *TriggerOutputs) (T, error)) error {
	wrappedCfg, err := values.CreateMapFromStruct(*config)
	if err != nil {
		return err
	}

	wrappedHandler := func(runtime sdk.DonRuntime, triggerOutputs *values.Map) (*values.Map, error) {
		var typedTriggerOutputs TriggerOutputs
		err := triggerOutputs.UnwrapTo(&typedTriggerOutputs)
		if err != nil {
			return nil, err
		}
		results, err := handler(runtime, &typedTriggerOutputs)
		fmt.Println("calls awaited")
		if err != nil {
			return nil, err
		}

		return values.CreateMapFromStruct(results)
	}

	return runner.SubscribeToTrigger("basic-trigger@1.0.0", wrappedCfg, wrappedHandler)
}
