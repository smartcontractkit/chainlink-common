package host

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"google.golang.org/protobuf/types/known/emptypb"

	legacySdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/legacy"
	legacywasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/legacy/pb"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func GetTriggersSpec(ctx context.Context, modCfg *ModuleConfig, binary []byte, config []byte) ([]*wasmpb.TriggerSubscriptionRequest, error) {

}

func GetWorkflowSpec(ctx context.Context, modCfg *ModuleConfig, binary []byte, config []byte) (*legacySdk.WorkflowSpec, error) {
	m, err := NewModule(modCfg, binary, WithDeterminism())
	if err != nil {
		return nil, fmt.Errorf("could not instantiate module: %w", err)
	}

	if !m.isLegacyDAG {

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
