package pb

import (
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

func toStepDefinition(sd workflows.StepDefinition) (*StepDefinition, error) {
	var inputs *values.Map
	if sd.Inputs.Mapping != nil {
		i, err := values.WrapMap(sd.Inputs.Mapping)
		if err != nil {
			return nil, fmt.Errorf("could not translate config to map: %w", err)
		}
		inputs = i
	}

	wrappedConfig, err := values.WrapMap(sd.Config)
	if err != nil {
		return nil, fmt.Errorf("could not wrap config into map: %w", err)
	}

	return &StepDefinition{
		Id:  sd.ID,
		Ref: sd.Ref,
		Inputs: &StepInputs{
			OutputRef: sd.Inputs.OutputRef,
			Mapping:   values.ProtoMap(inputs),
		},
		Config:         values.ProtoMap(wrappedConfig),
		CapabilityType: string(sd.CapabilityType),
	}, nil
}

func WorkflowSpecToProto(spec *workflows.WorkflowSpec) (*WorkflowSpec, error) {
	ws := &WorkflowSpec{
		Name:      spec.Name,
		Owner:     spec.Owner,
		Triggers:  []*StepDefinition{},
		Actions:   []*StepDefinition{},
		Consensus: []*StepDefinition{},
		Targets:   []*StepDefinition{},
	}

	for _, t := range spec.Triggers {
		tt, err := toStepDefinition(t)
		if err != nil {
			return nil, fmt.Errorf("error translating step definition to proto: %w", err)
		}
		ws.Triggers = append(ws.Triggers, tt)
	}

	for _, a := range spec.Actions {
		ta, err := toStepDefinition(a)
		if err != nil {
			return nil, fmt.Errorf("error translating step definition to proto: %w", err)
		}
		ws.Actions = append(ws.Actions, ta)
	}

	for _, c := range spec.Consensus {
		tc, err := toStepDefinition(c)
		if err != nil {
			return nil, fmt.Errorf("error translating step definition to proto: %w", err)
		}
		ws.Consensus = append(ws.Consensus, tc)
	}

	for _, t := range spec.Targets {
		tt, err := toStepDefinition(t)
		if err != nil {
			return nil, fmt.Errorf("error translating step definition to proto: %w", err)
		}
		ws.Targets = append(ws.Consensus, tt)
	}

	return ws, nil
}

func fromStepDefinition(sd *StepDefinition) (workflows.StepDefinition, error) {
	if sd.Inputs == nil {
		return workflows.StepDefinition{}, errors.New("invalid step definition: inputs cannot be nil")
	}

	var mapping map[string]any
	if sd.Inputs.Mapping != nil {
		v, err := values.FromMapValueProto(sd.Inputs.Mapping)
		if err != nil {
			return workflows.StepDefinition{}, fmt.Errorf("invalid step definition: could not convert inputs mapping to value: %w", err)
		}

		err = v.UnwrapTo(&mapping)
		if err != nil {
			return workflows.StepDefinition{}, fmt.Errorf("invalid step definition: could not unwrap inputs mapping: %w", err)
		}
	}

	mvConfig, err := values.FromMapValueProto(sd.Config)
	if err != nil {
		return workflows.StepDefinition{}, fmt.Errorf("invalid step definition: could not unwrap config: %w", err)
	}

	cmapping := map[string]any{}
	if mvConfig != nil {
		err := mvConfig.UnwrapTo(&cmapping)
		if err != nil {
			return workflows.StepDefinition{}, fmt.Errorf("invalid step definition: could not unwrap config to map: %w", err)
		}
	}

	return workflows.StepDefinition{
		ID:  sd.Id,
		Ref: sd.Ref,
		Inputs: workflows.StepInputs{
			OutputRef: sd.Inputs.OutputRef,
			Mapping:   mapping,
		},
		Config:         cmapping,
		CapabilityType: capabilities.CapabilityType(sd.CapabilityType),
	}, nil
}

func ProtoToWorkflowSpec(spec *WorkflowSpec) (*workflows.WorkflowSpec, error) {
	ws := &workflows.WorkflowSpec{
		Name:      spec.Name,
		Owner:     spec.Owner,
		Triggers:  []workflows.StepDefinition{},
		Actions:   []workflows.StepDefinition{},
		Consensus: []workflows.StepDefinition{},
		Targets:   []workflows.StepDefinition{},
	}

	for _, t := range spec.Triggers {
		tt, err := fromStepDefinition(t)
		if err != nil {
			return nil, fmt.Errorf("error translating step definition to proto: %w", err)
		}
		ws.Triggers = append(ws.Triggers, tt)
	}

	for _, a := range spec.Actions {
		ta, err := fromStepDefinition(a)
		if err != nil {
			return nil, fmt.Errorf("error translating step definition to proto: %w", err)
		}
		ws.Actions = append(ws.Actions, ta)
	}

	for _, c := range spec.Consensus {
		tc, err := fromStepDefinition(c)
		if err != nil {
			return nil, fmt.Errorf("error translating step definition to proto: %w", err)
		}
		ws.Consensus = append(ws.Consensus, tc)
	}

	for _, t := range spec.Targets {
		tt, err := fromStepDefinition(t)
		if err != nil {
			return nil, fmt.Errorf("error translating step definition to proto: %w", err)
		}
		ws.Targets = append(ws.Consensus, tt)
	}

	return ws, nil
}
