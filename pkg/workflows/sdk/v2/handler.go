package sdk

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Handler[T any] interface {
	Id() string
	Method() string
	TriggerCfg() *anypb.Any
	Callback() func(runtime T, triggerOutputs *anypb.Any) (any, error)
}

func NewDonHandler[T proto.Message, O any](trigger DonTrigger[T], callback func(runtime DonRuntime, triggerOutputs T) (O, error)) Handler[DonRuntime] {
	return newHandler[DonRuntime, T, DonTrigger[T]](trigger, callback)
}

func NewEmptyDonHandler[T proto.Message](trigger DonTrigger[T], callback func(runtime DonRuntime, triggerOutputs T) error) Handler[DonRuntime] {
	return newEmptyDonHandler[DonRuntime, T, DonTrigger[T]](trigger, callback)
}

func NewNodeHandler[T proto.Message, O any](trigger NodeTrigger[T], callback func(runtime NodeRuntime, triggerOutputs T) (O, error)) Handler[NodeRuntime] {
	return newHandler[NodeRuntime, T, NodeTrigger[T]](trigger, callback)
}

func NewEmptyNodeHandler[T proto.Message](trigger NodeTrigger[T], callback func(runtime NodeRuntime, triggerOutputs T) error) Handler[NodeRuntime] {
	return newEmptyDonHandler[NodeRuntime, T, NodeTrigger[T]](trigger, callback)
}

func newHandler[R any, M proto.Message, T Trigger[M], O any](trigger T, callback func(runtime R, triggerOutputs M) (O, error)) Handler[R] {
	wrapped := func(runtime R, triggerOutputs *anypb.Any) (any, error) {
		unwrappedTrigger := trigger.NewT()
		if err := triggerOutputs.UnmarshalTo(unwrappedTrigger); err != nil {
			return nil, err
		}

		return callback(runtime, unwrappedTrigger)
	}

	return &handler[R, M]{Trigger: trigger, fn: wrapped}
}

func newEmptyDonHandler[R any, M proto.Message, T Trigger[M]](trigger T, callback func(runtime R, triggerOutputs M) error) Handler[R] {
	wrapped := func(runtime R, triggerOutputs *anypb.Any) (any, error) {
		unwrappedTrigger := trigger.NewT()
		if err := triggerOutputs.UnmarshalTo(unwrappedTrigger); err != nil {
			return nil, err
		}

		return nil, callback(runtime, unwrappedTrigger)
	}

	return &handler[R, M]{Trigger: trigger, fn: wrapped}
}

type handler[R any, T proto.Message] struct {
	Trigger[T]
	fn func(runtime R, triggerOutputs *anypb.Any) (any, error)
}

func (h *handler[R, T]) TriggerCfg() *anypb.Any {
	return h.Trigger.ConfigAsAny()
}

func (h *handler[R, T]) Callback() func(runtime R, triggerOutputs *anypb.Any) (any, error) {
	return h.fn
}

var _ Handler[any] = (*handler[any, proto.Message])(nil)
