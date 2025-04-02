package host

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func GetWorkflowSpec(ctx context.Context, modCfg *ModuleConfig, binary []byte, config []byte) (*sdk.WorkflowSpec, error) {
	m, err := NewModule(modCfg, binary, WithDeterminism())
	if err != nil {
		return nil, fmt.Errorf("could not instantiate module: %w", err)
	}

	m.Start()

	rid := uuid.New().String()
	req := &wasmpb.Request{
		Id:     rid,
		Config: config,
		Message: &wasmpb.Request_SpecRequest{
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

	return wasmpb.ProtoToWorkflowSpec(sr)
}
