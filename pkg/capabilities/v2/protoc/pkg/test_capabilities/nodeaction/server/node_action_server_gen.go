// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc, DO NOT EDIT.

package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// Avoid unused imports if there is configuration type
var _ = emptypb.Empty{}

type BasicActionCapability interface {
	PerformAction(ctx context.Context, metadata capabilities.RequestMetadata, input *nodeaction.NodeInputs) (*nodeaction.NodeOutputs, error)

	Start(ctx context.Context) error
	Close() error
	HealthReport() map[string]error
	Name() string
	Description() string
	Ready() error
	Initialise(ctx context.Context, config string, telemetryService core.TelemetryService, store core.KeyValueStore, errorLog core.ErrorLog, pipelineRunner core.PipelineRunnerService, relayerSet core.RelayerSet, oracleFactory core.OracleFactory) error
}

func NewBasicActionServer(capability BasicActionCapability) *BasicActionServer {
	stopCh := make(chan struct{})
	return &BasicActionServer{
		basicActionCapability: basicActionCapability{BasicActionCapability: capability, stopCh: stopCh},
		stopCh:                stopCh,
	}
}

type BasicActionServer struct {
	basicActionCapability
	capabilityRegistry core.CapabilitiesRegistry
	stopCh             chan struct{}
}

func (cs *BasicActionServer) Initialise(ctx context.Context, config string, telemetryService core.TelemetryService, store core.KeyValueStore, capabilityRegistry core.CapabilitiesRegistry, errorLog core.ErrorLog, pipelineRunner core.PipelineRunnerService, relayerSet core.RelayerSet, oracleFactory core.OracleFactory) error {
	if err := cs.BasicActionCapability.Initialise(ctx, config, telemetryService, store, errorLog, pipelineRunner, relayerSet, oracleFactory); err != nil {
		return fmt.Errorf("error when initializing capability: %w", err)
	}

	cs.capabilityRegistry = capabilityRegistry

	if err := capabilityRegistry.Add(ctx, &basicActionCapability{
		BasicActionCapability: cs.BasicActionCapability,
	}); err != nil {
		return fmt.Errorf("error when adding kv store action to the registry: %w", err)
	}

	return nil
}

func (cs *BasicActionServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if cs.capabilityRegistry != nil {
		if err := cs.capabilityRegistry.Remove(ctx, "basic-test-node-action@1.0.0"); err != nil {
			return err
		}
	}

	if cs.stopCh != nil {
		close(cs.stopCh)
	}

	return cs.basicActionCapability.Close()
}

func (cs *BasicActionServer) Infos(ctx context.Context) ([]capabilities.CapabilityInfo, error) {
	info, err := cs.basicActionCapability.Info(ctx)
	if err != nil {
		return nil, err
	}
	return []capabilities.CapabilityInfo{info}, nil
}

type basicActionCapability struct {
	BasicActionCapability
	stopCh chan struct{}
}

func (c *basicActionCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	// Maybe we do need to split it out, even if the user doesn't see it
	return capabilities.NewCapabilityInfo("basic-test-node-action@1.0.0", capabilities.CapabilityTypeCombined, c.BasicActionCapability.Description())
}

var _ capabilities.ExecutableAndTriggerCapability = (*basicActionCapability)(nil)

func (c *basicActionCapability) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	return nil, fmt.Errorf("trigger %s not found", request.Method)
}

func (c *basicActionCapability) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	return fmt.Errorf("trigger %s not found", request.Method)
}

func (c *basicActionCapability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	return nil
}

func (c *basicActionCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	return nil
}

func (c *basicActionCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	response := capabilities.CapabilityResponse{}
	switch request.Method {
	case "PerformAction":
		input := &nodeaction.NodeInputs{}
		config := &emptypb.Empty{}
		wrapped := func(ctx context.Context, metadata capabilities.RequestMetadata, input *nodeaction.NodeInputs, _ *emptypb.Empty) (*nodeaction.NodeOutputs, error) {
			return c.BasicActionCapability.PerformAction(ctx, metadata, input)
		}
		return capabilities.Execute(ctx, request, input, config, wrapped)
	default:
		return response, fmt.Errorf("method %s not found", request.Method)
	}
}
