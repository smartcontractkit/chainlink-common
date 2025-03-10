package basictrigger

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func Subscribe[T any](runner sdk.DonRunner, config *TriggerConfig, handler func(sdk.DonRuntime, *TriggerOutputs) (T, error)) error {
	panic("TODO")
	/*
		// TODO this shouldn't be a values.Value, but I didn't re-generate the classes with proto also yet...
		// maybe just do it now...
		vconfig, err := values.Wrap(config)
		if err != nil {
			return err
		}

		wrappedCfg, err := anypb.New(values.Proto(vconfig))
		if err != nil {
			return err
		}

		wrappedHandler := func(runtime sdk.DonRuntime, triggerOutputs *anypb.Any) ([]bytes, error) {
			// TODO
			var typedTriggerOutputs TriggerOutputs
			err := triggerOutputs.UnwrapTo(&typedTriggerOutputs)
			if err != nil {
				return nil, err
			}
			results, err := handler(runtime, &typedTriggerOutputs)
			if err != nil {
				return nil, err
			}

			return values.CreateMapFromStruct(results)
		}

		return runner.SubscribeToTrigger("basic-trigger@1.0.0", wrappedCfg, wrappedHandler)

	*/
}
