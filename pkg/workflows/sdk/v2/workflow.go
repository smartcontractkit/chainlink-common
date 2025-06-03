package sdk

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Workflows[C any] []BaseWorkflow[C, Runtime]

func (w *Workflows[C]) Register(workflow BaseWorkflow[C, Runtime]) {
	*w = append(*w, workflow)
}

func OnValue[C any, M proto.Message, T Trigger[M], O any](trigger T, callback func(wcx *WorkflowContext[C], runtime Runtime, payload M) (O, error)) BaseWorkflow[C, Runtime] {
	return onValue(trigger, callback)
}

func On[C any, M proto.Message, T Trigger[M]](trigger T, callback func(wcx *WorkflowContext[C], runtime Runtime, payload M) error) BaseWorkflow[C, Runtime] {
	return onValue(trigger, func(wcx *WorkflowContext[C], runtime Runtime, payload M) (struct{}, error) {
		return struct{}{}, callback(wcx, runtime, payload)
	})
}

// BaseWorkflow is meant to be used internally by the SDK to define workflows.
type BaseWorkflow[C, R any] interface {
	CapabilityID() string
	Method() string
	TriggerCfg() *anypb.Any
	Callback() func(wcx *WorkflowContext[C], runtime R, payload *anypb.Any) (any, error)
}

func onValue[R, C any, M proto.Message, T Trigger[M], O any](trigger T, callback func(wcx *WorkflowContext[C], runtime R, payload M) (O, error)) BaseWorkflow[C, R] {
	wrapped := func(wcx *WorkflowContext[C], runtime R, payload *anypb.Any) (any, error) {
		unwrappedTrigger := trigger.NewT()
		if err := payload.UnmarshalTo(unwrappedTrigger); err != nil {
			return nil, err
		}
		return callback(wcx, runtime, unwrappedTrigger)
	}
	return &workflowImpl[C, R, M]{
		Trigger: trigger,
		fn:      wrapped,
	}
}

type workflowImpl[C, R any, M proto.Message] struct {
	Trigger[M]
	fn func(wcx *WorkflowContext[C], runtime R, trigger *anypb.Any) (any, error)
}

var _ BaseWorkflow[int, any] = (*workflowImpl[int, any, proto.Message])(nil)

func (h *workflowImpl[C, R, M]) TriggerCfg() *anypb.Any {
	return h.Trigger.ConfigAsAny()
}

func (h *workflowImpl[C, R, M]) Callback() func(wcx *WorkflowContext[C], runtime R, payload *anypb.Any) (any, error) {
	return h.fn
}
