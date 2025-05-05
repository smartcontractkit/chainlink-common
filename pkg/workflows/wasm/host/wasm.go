package host

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"google.golang.org/protobuf/types/known/emptypb"

	legacySdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	legacywasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

func GetWorkflowSpec(ctx context.Context, modCfg *ModuleConfig, binary []byte, config []byte) (*legacySdk.WorkflowSpec, error) {
	m, err := NewModule(modCfg, binary, WithDeterminism())
	if err != nil {
		return nil, fmt.Errorf("could not instantiate module: %w", err)
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

func GetTriggersSpec(ctx context.Context, modCfg *ModuleConfig, binary []byte, config []byte) (*sdkpb.TriggerSubscriptionRequest, error) {
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
