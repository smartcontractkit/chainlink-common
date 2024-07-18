package workflows

import (
	"strconv"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

// 1. Capability defines JSON schema for inputs and outputs of a capability.
// Trigger: triggerOutputType := workflowBuilder.addTrigger(DataStreamsTrigger.ModifiedConfig{})
// Adds metadata to the builder. Returns output type.
// 2. Consensus: consensusOutputType := workflowBuilder.addConsensus(ConsensusConfig{
// 	Inputs: triggerOutputType,
// })

type Workflow struct {
	spec *WorkflowSpec
}

type CapabilityDefinition[O any] interface {
	Ref() any
	self() CapabilityDefinition[O]
}

type CapabilityListDefinition[O any] interface {
	CapabilityDefinition[[]O]
	Index(i int) CapabilityDefinition[O]
}

func ListOf[O any](capabilities ...CapabilityDefinition[O]) CapabilityListDefinition[O] {
	impl := multiCapabilityList[O]{refs: make([]any, len(capabilities))}
	for i, c := range capabilities {
		impl.refs[i] = c.Ref()
	}
	return &impl
}

func ConstantDefinition[O any](o O) CapabilityDefinition[O] {
	return &capabilityDefinitionImpl[O]{ref: o}
}

// ToListDefinition TODO think if this is actually broken, what if the definitions were built up, would this still work?
// also what about hard-coded?
func ToListDefinition[O any](c CapabilityDefinition[[]O]) CapabilityListDefinition[O] {
	return &singleCapabilityList[O]{CapabilityDefinition: c}
}

type multiCapabilityList[O any] struct {
	refs []any
}

func (c *multiCapabilityList[O]) Index(i int) CapabilityDefinition[O] {
	return &capabilityDefinitionImpl[O]{ref: c.refs[i]}
}

func (c *multiCapabilityList[O]) Ref() any {
	return c.refs
}

func (c *multiCapabilityList[O]) self() CapabilityDefinition[[]O] {
	return c
}

type singleCapabilityList[O any] struct {
	CapabilityDefinition[[]O]
}

func (s singleCapabilityList[O]) Index(i int) CapabilityDefinition[O] {
	return &capabilityDefinitionImpl[O]{ref: s.CapabilityDefinition.Ref().(string) + "." + strconv.FormatInt(int64(i), 10)}
}

type Step[O any] struct {
	Definition StepDefinition
}

type capabilityDefinitionImpl[O any] struct {
	ref any
}

func (c *capabilityDefinitionImpl[O]) Ref() any {
	return c.ref
}

func (c *capabilityDefinitionImpl[O]) self() CapabilityDefinition[O] {
	return c
}

type NewWorkflowParams struct {
	Owner string
	Name  string
}

func NewWorkflow(
	params NewWorkflowParams,
) *Workflow {
	return &Workflow{
		spec: &WorkflowSpec{
			Owner: params.Owner,
			Name:  params.Name,
		},
	}
}

// AddStep is meant to be called by generated code
func AddStep[O any](w *Workflow, step Step[O]) (CapabilityDefinition[O], error) {
	// TODO should return error if the name is already used
	stepDefinition := step.Definition

	switch stepDefinition.CapabilityType {
	case capabilities.CapabilityTypeTrigger:
		w.spec.Triggers = append(w.spec.Triggers, stepDefinition)
	case capabilities.CapabilityTypeAction:
		w.spec.Actions = append(w.spec.Actions, stepDefinition)
	case capabilities.CapabilityTypeConsensus:
		w.spec.Consensus = append(w.spec.Consensus, stepDefinition)
	case capabilities.CapabilityTypeTarget:
		w.spec.Targets = append(w.spec.Targets, stepDefinition)
	}

	return &capabilityDefinitionImpl[O]{ref: step.Definition.Ref}, nil
}

// AccessField is meant to be used by generated code
func AccessField[I, O any](c CapabilityDefinition[I], fieldName string) CapabilityDefinition[O] {
	return &capabilityDefinitionImpl[O]{ref: c.Ref().(string) + "." + fieldName}
}

func (w Workflow) Spec() WorkflowSpec {
	return *w.spec
}

// ComponentCapabilityDefinition is meant to be used by generated code
type ComponentCapabilityDefinition[O any] map[string]any

func (c ComponentCapabilityDefinition[O]) Ref() any {
	return map[string]any(c)
}

func (c ComponentCapabilityDefinition[O]) self() CapabilityDefinition[O] {
	return c
}
