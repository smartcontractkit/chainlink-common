package sdk

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// ExecutionHandler is meant to be used internally by the SDK to define workflows.
type ExecutionHandler[C, R any] interface {
	CapabilityID() string
	Method() string
	TriggerCfg() *anypb.Any
	Callback() func(env *Environment[C], runtime R, payload *anypb.Any) (any, error)
}

func Handler[C any, M proto.Message, T any, O any](trigger Trigger[M, T], callback func(env *Environment[C], runtime Runtime, payload T) (O, error)) ExecutionHandler[C, Runtime] {
	return handler(trigger, callback)
}

func handler[R, C any, M proto.Message, T any, O any](trigger Trigger[M, T], callback func(env *Environment[C], runtime R, payload T) (O, error)) ExecutionHandler[C, R] {
	wrapped := func(env *Environment[C], runtime R, payload *anypb.Any) (any, error) {
		unwrappedTrigger := trigger.NewT()
		if err := payload.UnmarshalTo(unwrappedTrigger); err != nil {
			return nil, err
		}
		input, err := trigger.Adapt(unwrappedTrigger)
		if err != nil {
			return nil, err
		}
		return callback(env, runtime, input)
	}
	return &executionHandlerImpl[C, R, M, T]{
		Trigger: trigger,
		fn:      wrapped,
	}
}

type executionHandlerImpl[C, R any, M proto.Message, T any] struct {
	Trigger[M, T]
	fn func(env *Environment[C], runtime R, trigger *anypb.Any) (any, error)
}

var _ ExecutionHandler[int, any] = (*executionHandlerImpl[int, any, proto.Message, any])(nil)

func (h *executionHandlerImpl[C, R, M, T]) TriggerCfg() *anypb.Any {
	return h.Trigger.ConfigAsAny()
}

func (h *executionHandlerImpl[C, R, M, T]) Callback() func(env *Environment[C], runtime R, payload *anypb.Any) (any, error) {
	return h.fn
}
