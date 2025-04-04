// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc, DO NOT EDIT.

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type BasicCapability interface {
	Trigger(ctx context.Context, input *basictrigger.Config /* TODO config */) (*basictrigger.Outputs, error)
	Start(ctx context.Context) error
	Close() error
	HealthReport() map[string]error
	Name() string
	Description() string
	Ready() error
	Initialise(ctx context.Context, config string, telemetryService core.TelemetryService, store core.KeyValueStore, errorLog core.ErrorLog, pipelineRunner core.PipelineRunnerService, relayerSet core.RelayerSet, oracleFactory core.OracleFactory) error
}

func NewBasicServer(capability BasicCapability) loop.StandardCapabilities {
	return &basicServer{
		basicCapability: basicCapability{BasicCapability: capability},
	}
}

type basicServer struct {
	basicCapability
	capabilityRegistry core.CapabilitiesRegistry
}

func (cs *basicServer) Initialise(ctx context.Context, config string, telemetryService core.TelemetryService, store core.KeyValueStore, capabilityRegistry core.CapabilitiesRegistry, errorLog core.ErrorLog, pipelineRunner core.PipelineRunnerService, relayerSet core.RelayerSet, oracleFactory core.OracleFactory) error {
	if err := cs.BasicCapability.Initialise(ctx, config, telemetryService, store, errorLog, pipelineRunner, relayerSet, oracleFactory); err != nil {
		return fmt.Errorf("error when initializing capability: %w", err)
	}

	cs.capabilityRegistry = capabilityRegistry

	if err := capabilityRegistry.Add(ctx, &basicCapability{
		BasicCapability: cs.BasicCapability,
	}); err != nil {
		return fmt.Errorf("error when adding kv store action to the registry: %w", err)
	}

	return nil
}

func (cs *basicServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := cs.capabilityRegistry.Remove(ctx, "basic-test-trigger@1.0.0"); err != nil {
		return err
	}

	return cs.basicCapability.Close()
}

func (cs *basicServer) Infos(ctx context.Context) ([]capabilities.CapabilityInfo, error) {
	// TODO if we get rid of targets in favour of actions that return empty proto, do we need Consensus stil?
	info, err := cs.basicCapability.Info(ctx)
	if err != nil {
		return nil, err
	}
	return []capabilities.CapabilityInfo{info}, nil
}

type basicCapability struct {
	BasicCapability
}

var _ capabilities.TriggerCapability = (*basicCapability)(nil)

func (c *basicCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	return capabilities.NewCapabilityInfo("basic-test-trigger@1.0.0", capabilities.CapabilityTypeAction, c.BasicCapability.Description())
}

func (c *basicCapability) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *basicCapability) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	//TODO implement me
	panic("implement me")
}
