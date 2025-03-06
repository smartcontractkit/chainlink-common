package sdk

import (
	"google.golang.org/protobuf/types/known/anypb"
)

type DonRunner interface {
	// SubscribeToTrigger is meant to be called by generated code, prefer to use the generated code
	SubscribeToTrigger(id string, triggerCfg *anypb.Any, handler func(runtime DonRuntime, triggerOutputs *anypb.Any) ([]byte, error)) error
}

type NodeRunner interface {
	// SubscribeToTrigger is meant to be called by generated code, prefer to use the generated code
	SubscribeToTrigger(id string, triggerCfg *anypb.Any, handler func(runtime NodeRuntime, triggerOutputs *anypb.Any) ([]byte, error)) error
}
