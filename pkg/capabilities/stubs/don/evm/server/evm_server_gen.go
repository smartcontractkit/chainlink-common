// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc, DO NOT EDIT.

package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/crosschain"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type EvmCapabilityCapability interface {
	GetTxResult(ctx context.Context, metadata capabilities.RequestMetadata, input *evm.TxID /* TODO this isn't right */, config *evm.TxID) (*crosschain.TxResult, error)
	RegisterGetTxResult(ctx context.Context, metadata capabilities.RegistrationMetadata /* TODO config */) error
	UnregisterGetTxResult(ctx context.Context, metadata capabilities.RegistrationMetadata /* TODO config */) error

	ReadMethod(ctx context.Context, metadata capabilities.RequestMetadata, input *evm.ReadMethodRequest /* TODO this isn't right */, config *evm.ReadMethodRequest) (*crosschain.ByteArray, error)
	RegisterReadMethod(ctx context.Context, metadata capabilities.RegistrationMetadata /* TODO config */) error
	UnregisterReadMethod(ctx context.Context, metadata capabilities.RegistrationMetadata /* TODO config */) error

	QueryLogs(ctx context.Context, metadata capabilities.RequestMetadata, input *evm.QueryLogsRequest /* TODO this isn't right */, config *evm.QueryLogsRequest) (*evm.LogList, error)
	RegisterQueryLogs(ctx context.Context, metadata capabilities.RegistrationMetadata /* TODO config */) error
	UnregisterQueryLogs(ctx context.Context, metadata capabilities.RegistrationMetadata /* TODO config */) error

	SubmitTransaction(ctx context.Context, metadata capabilities.RequestMetadata, input *evm.SubmitTransactionRequest /* TODO this isn't right */, config *evm.SubmitTransactionRequest) (*evm.TxID, error)
	RegisterSubmitTransaction(ctx context.Context, metadata capabilities.RegistrationMetadata /* TODO config */) error
	UnregisterSubmitTransaction(ctx context.Context, metadata capabilities.RegistrationMetadata /* TODO config */) error

	RegisterOnFinalityViolation(ctx context.Context, metadata capabilities.RequestMetadata, input *emptypb.Empty) (<-chan capabilities.TriggerAndId[*crosschain.BlockRange], error)
	UnregisterOnFinalityViolation(ctx context.Context, metadata capabilities.RequestMetadata, input *emptypb.Empty) error
	Start(ctx context.Context) error
	Close() error
	HealthReport() map[string]error
	Name() string
	Description() string
	Ready() error
	Initialise(ctx context.Context, config string, telemetryService core.TelemetryService, store core.KeyValueStore, errorLog core.ErrorLog, pipelineRunner core.PipelineRunnerService, relayerSet core.RelayerSet, oracleFactory core.OracleFactory) error
}

func NewEvmCapabilityServer(capability EvmCapabilityCapability) loop.StandardCapabilities {
	return &evmCapabilityServer{
		evmCapabilityCapability: evmCapabilityCapability{EvmCapabilityCapability: capability},
	}
}

type evmCapabilityServer struct {
	evmCapabilityCapability
	capabilityRegistry core.CapabilitiesRegistry
}

func (cs *evmCapabilityServer) Initialise(ctx context.Context, config string, telemetryService core.TelemetryService, store core.KeyValueStore, capabilityRegistry core.CapabilitiesRegistry, errorLog core.ErrorLog, pipelineRunner core.PipelineRunnerService, relayerSet core.RelayerSet, oracleFactory core.OracleFactory) error {
	if err := cs.EvmCapabilityCapability.Initialise(ctx, config, telemetryService, store, errorLog, pipelineRunner, relayerSet, oracleFactory); err != nil {
		return fmt.Errorf("error when initializing capability: %w", err)
	}

	cs.capabilityRegistry = capabilityRegistry

	if err := capabilityRegistry.Add(ctx, &evmCapabilityCapability{
		EvmCapabilityCapability: cs.EvmCapabilityCapability,
	}); err != nil {
		return fmt.Errorf("error when adding kv store action to the registry: %w", err)
	}

	return nil
}

func (cs *evmCapabilityServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := cs.capabilityRegistry.Remove(ctx, "evm@1.0.0"); err != nil {
		return err
	}

	return cs.evmCapabilityCapability.Close()
}

func (cs *evmCapabilityServer) Infos(ctx context.Context) ([]capabilities.CapabilityInfo, error) {
	// TODO if we get rid of targets in favour of actions that return empty proto, do we need Consensus stil?
	info, err := cs.evmCapabilityCapability.Info(ctx)
	if err != nil {
		return nil, err
	}
	return []capabilities.CapabilityInfo{info}, nil
}

type evmCapabilityCapability struct {
	EvmCapabilityCapability
}

func (c *evmCapabilityCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	// TODO this is problematic right not because we can do both...?
	// Maybe we do need to split it out, even if the user doesn't see it
	return capabilities.NewCapabilityInfo("evm@1.0.0", capabilities.CapabilityTypeAction, c.EvmCapabilityCapability.Description())
}

var _ capabilities.TriggerCapability = (*evmCapabilityCapability)(nil)

func (c *evmCapabilityCapability) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	switch request.Method {
	case "OnFinalityViolation":
		input := &emptypb.Empty{}
		return capabilities.RegisterTrigger(ctx, "evm@1.0.0", request, input, c.EvmCapabilityCapability.RegisterOnFinalityViolation)
	default:
		return nil, fmt.Errorf("method %s not found", request.Method)
	}

}

func (c *evmCapabilityCapability) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	switch request.Method {
	case "OnFinalityViolation":
		input := &emptypb.Empty{}
		_, err := capabilities.FromValueOrAny(request.Config, request.Request, input)
		if err != nil {
			return err
		}
		return c.EvmCapabilityCapability.UnregisterOnFinalityViolation(ctx, request.Metadata, input)
	default:
		return fmt.Errorf("method %s not found", request.Method)
	}
}

var _ capabilities.ActionCapability = (*evmCapabilityCapability)(nil)

func (c *evmCapabilityCapability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	//TODO implement me
	panic("implement me")
}

func (c *evmCapabilityCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	//TODO implement me
	panic("implement me")
}

func (c *evmCapabilityCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	response := capabilities.CapabilityResponse{}
	switch request.Method {
	case "GetTxResult":
		input := &evm.TxID{}
		// TODO config
		config := &evm.TxID{}
		return capabilities.Execute(ctx, request, input, config, c.EvmCapabilityCapability.GetTxResult)
	case "ReadMethod":
		input := &evm.ReadMethodRequest{}
		// TODO config
		config := &evm.ReadMethodRequest{}
		return capabilities.Execute(ctx, request, input, config, c.EvmCapabilityCapability.ReadMethod)
	case "QueryLogs":
		input := &evm.QueryLogsRequest{}
		// TODO config
		config := &evm.QueryLogsRequest{}
		return capabilities.Execute(ctx, request, input, config, c.EvmCapabilityCapability.QueryLogs)
	case "SubmitTransaction":
		input := &evm.SubmitTransactionRequest{}
		// TODO config
		config := &evm.SubmitTransactionRequest{}
		return capabilities.Execute(ctx, request, input, config, c.EvmCapabilityCapability.SubmitTransaction)
	default:
		return response, fmt.Errorf("method %s not found", request.Method)
	}
}
