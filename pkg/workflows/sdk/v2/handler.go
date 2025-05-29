package sdk

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Handler methods are meant to be used by the internals of the CRE.
//
//	The interface represents what trigger to listen to and how to handle its invocations
type Handler[T any] interface {
	CapabilityID() string
	Method() string
	TriggerCfg() *anypb.Any
	Callback() func(runtime T, payload *anypb.Any) (any, error)
	Spend() *SpendLimits
}

// NewDonHandler creates a new Handler for a DonTrigger with a return value for the workflow
func NewDonHandler[T proto.Message, O any](trigger DonTrigger[T], callback func(runtime DonRuntime, payload T) (O, error)) Handler[DonRuntime] {
	return newHandler[DonRuntime, T, DonTrigger[T]](trigger, callback)
}

// NewEmptyDonHandler creates a new Handler for a DonTrigger without a return value for the workflow
func NewEmptyDonHandler[T proto.Message](trigger DonTrigger[T], callback func(runtime DonRuntime, payload T) error) Handler[DonRuntime] {
	return newEmptyDonHandler[DonRuntime, T, DonTrigger[T]](trigger, callback)
}

// NewNodeHandler creates a new Handler for a NodeTrigger with a return value for the workflow
func NewNodeHandler[T proto.Message, O any](trigger NodeTrigger[T], callback func(runtime NodeRuntime, payload T) (O, error)) Handler[NodeRuntime] {
	return newHandler[NodeRuntime, T, NodeTrigger[T]](trigger, callback)
}

// NewEmptyNodeHandler creates a new Handler for a NodeTrigger without a return value for the workflow
func NewEmptyNodeHandler[T proto.Message](trigger NodeTrigger[T], callback func(runtime NodeRuntime, payload T) error) Handler[NodeRuntime] {
	return newEmptyDonHandler[NodeRuntime, T, NodeTrigger[T]](trigger, callback)
}

func newHandler[R any, M proto.Message, T Trigger[M], O any](trigger T, callback func(runtime R, payload M) (O, error)) Handler[R] {
	wrapped := func(runtime R, payload *anypb.Any) (any, error) {
		unwrappedTrigger := trigger.NewT()
		if err := payload.UnmarshalTo(unwrappedTrigger); err != nil {
			return nil, err
		}

		return callback(runtime, unwrappedTrigger)
	}

	return &handler[R, M]{Trigger: trigger, fn: wrapped}
}

func newEmptyDonHandler[R any, M proto.Message, T Trigger[M]](trigger T, callback func(runtime R, payload M) error) Handler[R] {
	wrapped := func(runtime R, payload *anypb.Any) (any, error) {
		unwrappedTrigger := trigger.NewT()
		if err := payload.UnmarshalTo(unwrappedTrigger); err != nil {
			return nil, err
		}

		return nil, callback(runtime, unwrappedTrigger)
	}

	return &handler[R, M]{Trigger: trigger, fn: wrapped}
}

type handler[R any, T proto.Message] struct {
	Trigger[T]
	fn    func(runtime R, trigger *anypb.Any) (any, error)
	spend *SpendLimits
}

func (h *handler[R, T]) TriggerCfg() *anypb.Any {
	return h.Trigger.ConfigAsAny()
}

func (h *handler[R, T]) Callback() func(runtime R, payload *anypb.Any) (any, error) {
	return h.fn
}

func (h *handler[R, T]) Spend() *SpendLimits {
	return h.spend
}

// WithMaxSpend sets the spend limits for this handler
func (h *handler[R, T]) WithMaxSpend(limits *SpendLimits) *handler[R, T] {
	h.spend = limits
	return h
}

var _ Handler[any] = (*handler[any, proto.Message])(nil)
