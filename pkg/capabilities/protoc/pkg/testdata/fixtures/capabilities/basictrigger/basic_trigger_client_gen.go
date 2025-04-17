// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc, DO NOT EDIT.

package basictrigger

import (
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

type Basic struct {
	// TODO config types (optional)
	// TODO capability interfaces.
}

func (c Basic) Trigger(config *Config) sdk.DonTrigger[*Outputs] {
	configAny, _ := anypb.New(config)
	return &basicTrigger{
		config: configAny,
	}
}

type basicTrigger struct {
	config *anypb.Any
}

func (*basicTrigger) IsDonTrigger() {}

func (*basicTrigger) NewT() *Outputs {
	return &Outputs{}
}

func (*basicTrigger) Id() string {
	return "basic-test-trigger@1.0.0"
}

func (*basicTrigger) Method() string {
	return "Trigger"
}

func (t *basicTrigger) ConfigAsAny() *anypb.Any {
	return t.config
}
