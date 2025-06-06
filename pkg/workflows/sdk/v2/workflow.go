package sdk

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Workflows[C any] []BaseWorkflow[C, Runtime]

func (w *Workflows[C]) Register(workflow BaseWorkflow[C, Runtime]) {
	*w = append(*w, workflow)
}

func On[C any, M proto.Message, T any, O any](trigger Trigger[M, T], callback func(wcx *WorkflowContext[C], runtime Runtime, payload T) (O, error)) BaseWorkflow[C, Runtime] {
	return on(trigger, callback)
}

// BaseWorkflow is meant to be used internally by the SDK to define workflows.
type BaseWorkflow[C, R any] interface {
	CapabilityID() string
	Method() string
	TriggerCfg() *anypb.Any
	Callback() func(wcx *WorkflowContext[C], runtime R, payload *anypb.Any) (any, error)
}

func on[R, C any, M proto.Message, T any, O any](trigger Trigger[M, T], callback func(wcx *WorkflowContext[C], runtime R, payload T) (O, error)) BaseWorkflow[C, R] {
	wrapped := func(wcx *WorkflowContext[C], runtime R, payload *anypb.Any) (any, error) {
		unwrappedTrigger := trigger.NewT()
		if err := payload.UnmarshalTo(unwrappedTrigger); err != nil {
			return nil, err
		}
		input, err := trigger.Adapt(unwrappedTrigger)
		if err != nil {
			return nil, err
		}
		return callback(wcx, runtime, input)
	}
	return &workflowImpl[C, R, M, T]{
		Trigger: trigger,
		fn:      wrapped,
	}
}

type workflowImpl[C, R any, M proto.Message, T any] struct {
	Trigger[M, T]
	fn func(wcx *WorkflowContext[C], runtime R, trigger *anypb.Any) (any, error)
}

var _ BaseWorkflow[int, any] = (*workflowImpl[int, any, proto.Message, any])(nil)

func (h *workflowImpl[C, R, M, T]) TriggerCfg() *anypb.Any {
	return h.Trigger.ConfigAsAny()
}

func (h *workflowImpl[C, R, M, T]) Callback() func(wcx *WorkflowContext[C], runtime R, payload *anypb.Any) (any, error) {
	return h.fn
}
