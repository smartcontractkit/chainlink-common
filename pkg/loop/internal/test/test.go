package test

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/services"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

const ConfigTOML = `[Foo]
Bar = "Baz"
`

var (

	//CapabilitiesRegistry
	GetID          = "get-id"
	GetTriggerID   = "get-trigger-id"
	GetActionID    = "get-action-id"
	GetConsensusID = "get-consensus-id"
	GetTargetID    = "get-target-id"
	CapabilityInfo = capabilities.CapabilityInfo{
		ID:             "capability-info-id@1.0.0",
		CapabilityType: capabilities.CapabilityTypeAction,
		Description:    "capability-info-description",
	}
)

var _ capabilities.BaseCapability = (*baseCapability)(nil)

type baseCapability struct {
}

func (e baseCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	return CapabilityInfo, nil
}

type staticService struct {
	lggr logger.Logger
}

func NewStaticService(lggr logger.Logger) services.Service {
	return staticService{lggr: lggr}
}
func (s staticService) Name() string { return s.lggr.Name() }

func (s staticService) Start(ctx context.Context) error {
	s.lggr.Info("Started")
	return nil
}

func (s staticService) Close() error {
	s.lggr.Info("Closed")
	return nil
}

func (s staticService) Ready() error {
	s.lggr.Info("Ready")
	return nil
}

// HealthReport reports only for this single service. Override to include sub-services.
func (s staticService) HealthReport() map[string]error {
	return map[string]error{s.Name(): s.Ready()}
}
