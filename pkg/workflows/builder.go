package workflows

// The real implementations will come in a follow-up PR.
// these stubs allow the code from the generators to compile.
// A holistic view can be seen at https://github.com/smartcontractkit/chainlink-common/pull/695

type WorkflowSpecFactory struct{}

type CapDefinition[T any] interface {
	Ref() any
}

type Step[O any] struct {
	Definition StepDefinition
}

// AddTo is meant to be called by generated code
func (step *Step[O]) AddTo(_ *WorkflowSpecFactory) CapDefinition[O] {
	panic("TODO: implement")
}

// AccessField is meant to be used by generated code
func AccessField[I, O any](_ CapDefinition[I], _ string) CapDefinition[O] {
	panic("TODO: implement")
}

// ComponentCapDefinition is meant to be used by generated code
type ComponentCapDefinition[O any] map[string]any

func (ComponentCapDefinition[O]) Ref() any {
	panic("TODO: implement")
}
