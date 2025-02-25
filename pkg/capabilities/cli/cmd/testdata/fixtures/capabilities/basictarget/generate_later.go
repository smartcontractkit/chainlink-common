package basictarget

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type BasicTarget struct {
	Config TargetConfig
}

func (b *BasicTarget) Write(runtime sdk.DonRuntime, input *TargetInputs) sdk.EmptyPromise {
	config, _ := values.CreateMapFromStruct(b.Config)
	wrappedInput, _ := values.CreateMapFromStruct(input)
	result := runtime.CallCapability("basictarget@1.0.0", capabilities.CapabilityRequest{
		Config: config,
		Inputs: wrappedInput,
	})
	return sdk.ToEmptyPromise(sdk.Then(result, func(response values.Value) (struct{}, error) {
		var s struct{}
		return s, nil
	}))
}
