package sdk

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type DonRunner interface {
	// SubscribeToTrigger is meant to be called by SubscribeToDONTrigger only
	SubscribeToTrigger(id string, triggerCfg *anypb.Any, handler func(runtime DonRuntime, triggerOutputs *anypb.Any) (any, error))
}

// NOTE that some triggers in node mode need callbacks into the WASM when setting up (like WebSocket trigger).
// This is not in this interface yet.

type NodeRunner interface {
	// SubscribeToTrigger is meant to be called by SubscribeToNodeTrigger only
	SubscribeToTrigger(id string, triggerCfg *anypb.Any, handler func(runtime NodeRuntime, triggerOutputs *anypb.Any) (any, error))
}

type Trigger[T proto.Message] interface {
	NewT() T
	Id() string
	Config() *anypb.Any
}

type DonTrigger[T proto.Message] interface {
	Trigger[T]
	IsDonTrigger()
}

type NodeTrigger[T proto.Message] interface {
	Trigger[T]
	IsNodeTrigger()
}

func SubscribeToDonTrigger[T proto.Message](runner DonRunner, trigger DonTrigger[T], callback func(runtime DonRuntime, triggerOutputs T) (any, error)) {
	runner.SubscribeToTrigger(trigger.Id(), trigger.Config(), func(runtime DonRuntime, triggerOutputs *anypb.Any) (any, error) {
		unwrappedTrigger := trigger.NewT()
		if err := triggerOutputs.UnmarshalTo(unwrappedTrigger); err != nil {
			return nil, err
		}

		return callback(runtime, unwrappedTrigger)
	})
}

func SubscribeToNodeTrigger[T proto.Message](runner NodeRunner, trigger Trigger[T], callback func(runtime NodeRuntime, triggerOutputs T) (any, error)) {
	runner.SubscribeToTrigger(trigger.Id(), trigger.Config(), func(runtime NodeRuntime, triggerOutputs *anypb.Any) (any, error) {
		unwrappedTrigger := trigger.NewT()
		if err := triggerOutputs.UnmarshalTo(unwrappedTrigger); err != nil {
			return nil, err
		}

		return callback(runtime, unwrappedTrigger)
	})
}
