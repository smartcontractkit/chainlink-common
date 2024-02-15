package test

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.CapabilitiesRegistry = (*StaticCapabilitiesRegistry)(nil)

type StaticCapabilitiesRegistry struct{}

func (s *StaticCapabilitiesRegistry) Get(ctx context.Context, ID string) (capabilities.BaseCapability, error) {
	panic("not implement")
}

func (s *StaticCapabilitiesRegistry) GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error) {
	panic("not implement")
}

func (s *StaticCapabilitiesRegistry) GetAction(ctx context.Context, ID string) (capabilities.ActionCapability, error) {
	panic("not implement")
}

func (s *StaticCapabilitiesRegistry) GetConsensus(ctx context.Context, ID string) (capabilities.ConsensusCapability, error) {
	panic("not implement")
}

func (s *StaticCapabilitiesRegistry) GetTarget(ctx context.Context, ID string) (capabilities.TargetCapability, error) {
	panic("not implement")
}

func (s *StaticCapabilitiesRegistry) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	panic("not implement")
}

func (s *StaticCapabilitiesRegistry) Add(ctx context.Context, c capabilities.BaseCapability) error {
	panic("not implement")
}
