package sdk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

type WorkflowSpecFactory struct {
	spec           *WorkflowSpec
	names          map[string]bool
	duplicateNames map[string]bool
	emptyNames     bool
	badCapTypes    []string
	fns            map[string]func(sdk Runtime, request capabilities.CapabilityRequest) capabilities.CapabilityResponse
}

func (w *WorkflowSpecFactory) GetFn(name string) func(sdk Runtime, request capabilities.CapabilityRequest) capabilities.CapabilityResponse {
	return w.fns[name]
}

type CapDefinition[O any] interface {
	Ref() any
	self() CapDefinition[O]
}

type CapListDefinition[O any] interface {
	CapDefinition[[]O]
	Index(i int) CapDefinition[O]
}

func ListOf[O any](capabilities ...CapDefinition[O]) CapListDefinition[O] {
	impl := multiCapList[O]{refs: make([]any, len(capabilities))}
	for i, c := range capabilities {
		impl.refs[i] = c.Ref()
	}
	return &impl
}

func ConstantDefinition[O any](o O) CapDefinition[O] {
	return &capDefinitionImpl[O]{ref: o}
}

// ToListDefinition TODO think if this is actually broken, what if the definitions were built up, would this still work?
// also what about hard-coded?
func ToListDefinition[O any](c CapDefinition[[]O]) CapListDefinition[O] {
	return &singleCapList[O]{CapDefinition: c}
}

type multiCapList[O any] struct {
	refs []any
}

func (c *multiCapList[O]) Index(i int) CapDefinition[O] {
	return &capDefinitionImpl[O]{ref: c.refs[i]}
}

func (c *multiCapList[O]) Ref() any {
	return c.refs
}

// self is required to implement CapDefinition, complication fails without it, false positive.
// nolint
func (c *multiCapList[O]) self() CapDefinition[[]O] {
	return c
}

type singleCapList[O any] struct {
	CapDefinition[[]O]
}

func (s singleCapList[O]) Index(i int) CapDefinition[O] {
	return &capDefinitionImpl[O]{ref: s.CapDefinition.Ref().(string) + "." + strconv.FormatInt(int64(i), 10)}
}

type Step[O any] struct {
	Definition StepDefinition
}

type capDefinitionImpl[O any] struct {
	ref any
}

func (c *capDefinitionImpl[O]) Ref() any {
	return c.ref
}

// self is required to implement CapDefinition, complication fails without it, false positive.
// nolint
func (c *capDefinitionImpl[O]) self() CapDefinition[O] {
	return c
}

type NewWorkflowParams struct {
	Owner string
	Name  string
}

func NewWorkflowSpecFactory(
	params NewWorkflowParams,
) *WorkflowSpecFactory {
	return &WorkflowSpecFactory{
		spec: &WorkflowSpec{
			Owner:     params.Owner,
			Name:      params.Name,
			Triggers:  make([]StepDefinition, 0),
			Actions:   make([]StepDefinition, 0),
			Consensus: make([]StepDefinition, 0),
			Targets:   make([]StepDefinition, 0),
		},
		names:          map[string]bool{},
		duplicateNames: map[string]bool{},
		emptyNames:     false,
	}
}

// AddTo is meant to be called by generated code
func (step *Step[O]) AddTo(w *WorkflowSpecFactory) CapDefinition[O] {
	stepDefinition := step.Definition
	stepRef := stepDefinition.Ref
	if w.names[stepRef] && stepDefinition.CapabilityType != capabilities.CapabilityTypeTarget {
		w.duplicateNames[stepRef] = true
	}

	if stepRef == "" && stepDefinition.CapabilityType != capabilities.CapabilityTypeTarget {
		w.emptyNames = true
	}

	w.names[stepRef] = true

	switch stepDefinition.CapabilityType {
	case capabilities.CapabilityTypeTrigger:
		w.spec.Triggers = append(w.spec.Triggers, stepDefinition)
	case capabilities.CapabilityTypeAction:
		w.spec.Actions = append(w.spec.Actions, stepDefinition)
	case capabilities.CapabilityTypeConsensus:
		w.spec.Consensus = append(w.spec.Consensus, stepDefinition)
	case capabilities.CapabilityTypeTarget:
		w.spec.Targets = append(w.spec.Targets, stepDefinition)
	default:
		w.badCapTypes = append(w.badCapTypes, stepDefinition.ID)
	}

	return &capDefinitionImpl[O]{ref: fmt.Sprintf("$(%s.outputs)", step.Definition.Ref)}
}

// AccessField is meant to be used by generated code
func AccessField[I, O any](c CapDefinition[I], fieldName string) CapDefinition[O] {
	originalRef := c.Ref().(string)
	return &capDefinitionImpl[O]{ref: originalRef[:len(originalRef)-1] + "." + fieldName + ")"}
}

func (w *WorkflowSpecFactory) Spec() (WorkflowSpec, error) {
	if len(w.duplicateNames) > 0 {
		duplicates := make([]string, 0, len(w.duplicateNames))
		for k := range w.duplicateNames {
			duplicates = append(duplicates, k)
		}
		return WorkflowSpec{}, fmt.Errorf("duplicate step ids %v", strings.Join(duplicates, ", "))
	}

	if w.emptyNames {
		return WorkflowSpec{}, fmt.Errorf("empty step references are not allowed")
	}

	if len(w.badCapTypes) > 0 {
		return WorkflowSpec{}, fmt.Errorf("bad capability type for steps %v", strings.Join(w.badCapTypes, ", "))
	}

	return *w.spec, nil
}

// ComponentCapDefinition is meant to be used by generated code
type ComponentCapDefinition[O any] map[string]any

func (c ComponentCapDefinition[O]) Ref() any {
	return map[string]any(c)
}

// self is required to implement CapDefinition, complication fails without it, false positive.
// nolint
func (c ComponentCapDefinition[O]) self() CapDefinition[O] {
	return c
}
