package testutils

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// CapabilityMock allows for mocking of capabilities in a workflow
// they can be registered for a particular reference or entirely
// Note that registrations for a step are taken over registrations for a capability when there are both.
type CapabilityMock interface {
	Run(request capabilities.CapabilityRequest) capabilities.CapabilityResponse
	ID() string
}

func MockCapability[I, O any](id string, fn func(I) (O, error)) *Mock[I, O] {
	return &Mock[I, O]{
		id:      id,
		inputs:  map[string]I{},
		outputs: map[string]O{},
		errors:  map[string]error{},
		fn:      fn,
	}
}

type Mock[I, O any] struct {
	id      string
	inputs  map[string]I
	outputs map[string]O
	errors  map[string]error
	fn      func(I) (O, error)
}

var _ CapabilityMock = &Mock[any, any]{}

func (m *Mock[I, O]) Run(request capabilities.CapabilityRequest) capabilities.CapabilityResponse {
	var i I
	if err := request.Inputs.UnwrapTo(&i); err != nil {
		m.errors[request.Metadata.ReferenceID] = err
		return capabilities.CapabilityResponse{Err: err}
	}

	m.inputs[request.Metadata.ReferenceID] = i

	result, err := m.fn(i)
	if err != nil {
		m.errors[request.Metadata.ReferenceID] = err
		return capabilities.CapabilityResponse{Err: err}
	}

	m.outputs[request.Metadata.ReferenceID] = result

	wrapped, err := values.CreateMapFromStruct(result)
	if err != nil {
		m.errors[request.Metadata.ReferenceID] = err
		return capabilities.CapabilityResponse{Err: err}
	}

	return capabilities.CapabilityResponse{Value: wrapped}
}

func (m *Mock[I, O]) ID() string {
	return m.id
}

func (m *Mock[I, O]) GetStep(ref string) StepResults[I, O] {
	input, ran := m.inputs[ref]
	output := m.outputs[ref]
	err := m.errors[ref]
	return StepResults[I, O]{WasRun: ran, Input: input, Output: output, Error: err}
}

type StepResults[I, O any] struct {
	WasRun bool
	Input  I
	Output O
	Error  error
}

type TargetResults[I any] struct {
	NumRuns int
	Inputs  []I
	Errors  []error
}

type TriggerResults[O any] struct {
	Output O
	Error  error
}

func MockTrigger[O any](id string, fn func() (O, error)) *TriggerMock[O] {
	return &TriggerMock[O]{
		mock: MockCapability[struct{}, O](id, func(struct{}) (O, error) {
			return fn()
		}),
	}
}

type TriggerMock[O any] struct {
	mock *Mock[struct{}, O]
}

func (t *TriggerMock[O]) Run(request capabilities.CapabilityRequest) capabilities.CapabilityResponse {
	return t.mock.Run(request)
}

func (t *TriggerMock[O]) ID() string {
	return t.mock.ID()
}

func (t *TriggerMock[O]) GetStep() TriggerResults[O] {
	step := t.mock.GetStep("trigger")
	return TriggerResults[O]{Output: step.Output, Error: step.Error}
}

var _ CapabilityMock = &TriggerMock[any]{}

type TargetMock[I any] struct {
	mock *Mock[I, struct{}]
}

func MockTarget[I any](id string, fn func(I) error) *TargetMock[I] {
	return &TargetMock[I]{
		mock: MockCapability[I, struct{}](id, func(i I) (struct{}, error) {
			return struct{}{}, fn(i)
		}),
	}
}

func (t *TargetMock[I]) Run(request capabilities.CapabilityRequest) capabilities.CapabilityResponse {
	return t.mock.Run(request)
}

func (t *TargetMock[I]) ID() string {
	return t.mock.ID()
}

func (t *TargetMock[I]) GetAllWrites() TargetResults[I] {
	targetResults := TargetResults[I]{}
	for ref := range t.mock.inputs {
		targetResults.NumRuns++
		step := t.mock.GetStep(ref)
		targetResults.Inputs = append(targetResults.Inputs, step.Input)
		targetResults.Errors = append(targetResults.Errors, step.Error)
	}
	return targetResults
}

var _ CapabilityMock = &TargetMock[any]{}
