package host

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	legacySdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/legacy"
	legacywasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/legacy/pb"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func GetTriggersSpec(ctx context.Context, modCfg *ModuleConfig, binary []byte, config []byte) ([]*wasmpb.TriggerSubscriptionRequest, error) {
	m, err := NewModule(modCfg, binary, WithDeterminism())
	if err != nil {
		return nil, fmt.Errorf("could not instantiate module: %w", err)
	}
	m.Start()
	defer m.Close()

	execResult, err := m.Execute(ctx, &wasmpb.ExecuteRequest{
		Id:      uuid.New().String(),
		Config:  config,
		Request: &wasmpb.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}},
	})

	if err != nil {
		return nil, err
	}

	switch r := execResult.Result.(type) {
	case *wasmpb.ExecutionResult_Value:
		v, err := values.FromProto(r.Value)
		if err != nil {
			return nil, err
		}

		var unwrapped []*wasmpb.TriggerSubscriptionRequest
		/*
			ExecId  string     `protobuf:"bytes,1,opt,name=execId,proto3" json:"execId,omitempty"`
			Id      string     `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
			Payload *anypb.Any `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
		*/
		tmp, e2 := v.Unwrap()
		if e2 == nil {
			fmt.Println(tmp)
			tmp2 := &wasmpb.TriggerSubscriptionRequest{}
			if err = mapstructure.Decode(tmp.([]any)[0], &tmp2); err == nil {
				// This decodes correctly
				fmt.Printf("%+v\n", tmp2)
			} else {
				fmt.Println(err.Error())
			}
		}
		if err = v.UnwrapTo(&unwrapped); err != nil {
			// And obviously here
			return nil, err
		}

		return unwrapped, nil
	case *wasmpb.ExecutionResult_Error:
		return nil, errors.New(r.Error)
	default:
		return nil, errors.New("unexpected response from WASM binary: got nil spec response")
	}
}

func GetWorkflowSpec(ctx context.Context, modCfg *ModuleConfig, binary []byte, config []byte) (*legacySdk.WorkflowSpec, error) {
	m, err := NewModule(modCfg, binary, WithDeterminism())
	if err != nil {
		return nil, fmt.Errorf("could not instantiate module: %w", err)
	}

	if !m.isLegacyDAG {
		return wrapTriggersToWorkflowSpec(ctx, modCfg, binary, config)
	}

	m.Start()

	rid := uuid.New().String()
	req := &legacywasmpb.Request{
		Id:     rid,
		Config: config,
		Message: &legacywasmpb.Request_SpecRequest{
			SpecRequest: &emptypb.Empty{},
		},
	}
	resp, err := m.Run(ctx, req)
	if err != nil {
		return nil, err
	}

	sr := resp.GetSpecResponse()
	if sr == nil {
		return nil, errors.New("unexpected response from WASM binary: got nil spec response")
	}

	m.Close()

	return legacywasmpb.ProtoToWorkflowSpec(sr)
}

func wrapTriggersToWorkflowSpec(ctx context.Context, modCfg *ModuleConfig, binary []byte, config []byte) (*legacySdk.WorkflowSpec, error) {
	triggers, err := GetTriggersSpec(ctx, modCfg, binary, config)
	if err != nil {
		return nil, err
	}

	spec := &legacySdk.WorkflowSpec{Triggers: make([]legacySdk.StepDefinition, len(triggers))}

	for i, trigger := range triggers {
		spec.Triggers[i] = legacySdk.StepDefinition{
			ID:             trigger.Id,
			Ref:            fmt.Sprintf("trigger%d", i),
			ConfigProto:    trigger.Payload,
			CapabilityType: capabilities.CapabilityTypeTrigger,
		}
	}

	return spec, nil
}
