package testutils

import (
	"context"
	"encoding/json"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

func MockCapability[I, O any](id string, fn func(I) (O, error)) *Mock[I, O] {
	return &Mock[I, O]{mockBase: mockCapabilityBase[I, O](id, fn)}
}

func mockCapabilityBase[I, O any](id string, fn func(I) (O, error)) *mockBase[I, O] {
	return &mockBase[I, O]{
		id:      id,
		inputs:  map[string]I{},
		outputs: map[string]O{},
		errors:  map[string]error{},
		fn:      fn,
	}
}

type mockBase[I, O any] struct {
	id      string
	inputs  map[string]I
	outputs map[string]O
	errors  map[string]error
	fn      func(I) (O, error)
}

var _ capabilities.ExecutableCapability = &Mock[any, any]{}

func (m *mockBase[I, O]) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	return capabilities.CapabilityInfo{ID: m.id, IsLocal: true}, nil
}

func (m *mockBase[I, O]) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	return nil
}

func (m *mockBase[I, O]) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	return nil
}

func (m *mockBase[I, O]) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	var i I

	if err := request.Inputs.UnwrapTo(&i); err != nil {
		m.errors[request.Metadata.ReferenceID] = err
		return capabilities.CapabilityResponse{}, err
	}

	m.inputs[request.Metadata.ReferenceID] = i

	// validate against schema
	var tmp I
	b, err := json.Marshal(i)
	if err != nil {
		m.errors[request.Metadata.ReferenceID] = err
		return capabilities.CapabilityResponse{}, err
	}

	if err = json.Unmarshal(b, &tmp); err != nil {
		m.errors[request.Metadata.ReferenceID] = err
		return capabilities.CapabilityResponse{}, err
	}

	result, err := m.fn(i)
	if err != nil {
		m.errors[request.Metadata.ReferenceID] = err
		return capabilities.CapabilityResponse{}, err
	}

	m.outputs[request.Metadata.ReferenceID] = result

	wrapped, err := values.CreateMapFromStruct(result)
	if err != nil {
		m.errors[request.Metadata.ReferenceID] = err
		return capabilities.CapabilityResponse{}, err
	}

	return capabilities.CapabilityResponse{Value: wrapped}, nil
}

func (m *mockBase[I, O]) ID() string {
	return m.id
}

func (m *mockBase[I, O]) GetStep(ref string) StepResults[I, O] {
	input, ran := m.inputs[ref]
	output := m.outputs[ref]
	err := m.errors[ref]
	return StepResults[I, O]{WasRun: ran, Input: input, Output: output, Error: err}
}

type Mock[I, O any] struct {
	*mockBase[I, O]
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
		mockBase: mockCapabilityBase[struct{}, O](id, func(struct{}) (O, error) {
			return fn()
		}),
	}
}

type TriggerMock[O any] struct {
	*mockBase[struct{}, O]
}

var _ capabilities.TriggerCapability = &TriggerMock[any]{}

func (t *TriggerMock[O]) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	result, err := t.mockBase.fn(struct{}{})

	wrapped, wErr := values.CreateMapFromStruct(result)
	if wErr != nil {
		return nil, wErr
	}

	response := capabilities.TriggerResponse{
		Event: capabilities.TriggerEvent{
			TriggerType: "Mock " + t.ID(),
			ID:          t.ID(),
			Outputs:     wrapped,
		},
		Err: err,
	}
	ch := make(chan capabilities.TriggerResponse, 1)
	ch <- response
	close(ch)
	return ch, nil
}

func (t *TriggerMock[O]) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	return nil
}

func (t *TriggerMock[O]) AckEvent(ctx context.Context, triggerId string, eventId string, method string) error {
	return nil
}

func (t *TriggerMock[O]) GetStep() TriggerResults[O] {
	step := t.mockBase.GetStep("trigger")
	return TriggerResults[O]{Output: step.Output, Error: step.Error}
}

type TargetMock[I any] struct {
	*mockBase[I, struct{}]
}

func MockTarget[I any](id string, fn func(I) error) *TargetMock[I] {
	return &TargetMock[I]{
		mockBase: mockCapabilityBase[I, struct{}](id, func(i I) (struct{}, error) {
			return struct{}{}, fn(i)
		}),
	}
}

var _ capabilities.ExecutableCapability = &TargetMock[any]{}

func (t *TargetMock[I]) GetAllWrites() TargetResults[I] {
	targetResults := TargetResults[I]{}
	for ref := range t.mockBase.inputs {
		targetResults.NumRuns++
		step := t.mockBase.GetStep(ref)
		targetResults.Inputs = append(targetResults.Inputs, step.Input)
		if step.Error != nil {
			targetResults.Errors = append(targetResults.Errors, step.Error)
		}
	}
	return targetResults
}
