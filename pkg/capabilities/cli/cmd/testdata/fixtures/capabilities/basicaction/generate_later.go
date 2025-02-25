package basicaction

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type Basic struct {
	Config ActionConfig
}

func (b *Basic) Call(runtime sdk.DonRuntime, input *ActionInput) sdk.Promise[*ActionOutputs] {
	config, _ := values.CreateMapFromStruct(b.Config)
	wrappedInput, _ := values.CreateMapFromStruct(input)
	result := runtime.CallCapability("basicaction@1.0.0", capabilities.CapabilityRequest{
		Config: config,
		Inputs: wrappedInput,
	})
	return sdk.Then(result, func(response values.Value) (*ActionOutputs, error) {
		output := &ActionOutputs{}
		err := response.UnwrapTo(output)
		return output, err
	})
}
