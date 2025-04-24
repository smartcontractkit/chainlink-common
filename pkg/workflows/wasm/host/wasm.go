package host

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	legacySdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	legacywasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

func GetTriggersSpec(ctx context.Context, modCfg *ModuleConfig, binary []byte, config []byte) (*wasmpb.TriggerSubscriptionRequest, error) {
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
	case *wasmpb.ExecutionResult_TriggerSubscriptions:
		return r.TriggerSubscriptions, nil
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
	subscriptionRequest, err := GetTriggersSpec(ctx, modCfg, binary, config)
	if err != nil {
		return nil, err
	}

	spec := &legacySdk.WorkflowSpec{Triggers: make([]legacySdk.StepDefinition, len(subscriptionRequest.Subscriptions))}

	for i, trigger := range subscriptionRequest.Subscriptions {
		spec.Triggers[i] = legacySdk.StepDefinition{
			ID:             trigger.Id,
			Ref:            fmt.Sprintf("trigger%d", i),
			ConfigProto:    trigger.Payload,
			CapabilityType: capabilities.CapabilityTypeTrigger,
		}
	}

	return spec, nil
}
