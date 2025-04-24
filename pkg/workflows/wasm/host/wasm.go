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

func GetWorkflowSpec(ctx context.Context, modCfg *ModuleConfig, wasmStore WasmBinaryStore, workflowID string, config []byte) (*sdk.WorkflowSpec, error) {
	m, err := NewModule(ctx, modCfg, wasmStore, workflowID, WithDeterminism())
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

func NewSingleBinaryWasmBinaryStore(binary []byte) WasmBinaryStore {
	// Create a mock implementation of the wasmBinaryStore interface
	binaryStore := &SingleBinaryWasmBinaryStore{
		binary: binary,
	}
	return binaryStore
}

// SingleBinaryWasmBinaryStore is a mock implementation of the wasmBinaryStore interface
type SingleBinaryWasmBinaryStore struct {
	binary []byte
}

func (m *SingleBinaryWasmBinaryStore) GetSerialisedModulePath(workflowID string) (string, bool, error) {
	return "", false, nil
}

func (m *SingleBinaryWasmBinaryStore) StoreSerialisedModule(workflowID string, binaryID string, module []byte) error {
	//noop
	return nil
}

func (m *SingleBinaryWasmBinaryStore) GetWasmBinary(ctx context.Context, workflowID string) ([]byte, error) {
	return m.binary, nil
}
