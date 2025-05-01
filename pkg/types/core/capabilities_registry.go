package core

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

type CapabilitiesRegistry interface {
	LocalNode(ctx context.Context) (capabilities.Node, error)
	ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error)
	CapabilitiesRegistryBase
}

type CapabilitiesRegistryBase interface {
	GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error)
	Get(ctx context.Context, ID string) (capabilities.BaseCapability, error)
	GetExecutable(ctx context.Context, ID string) (capabilities.ExecutableCapability, error)
	List(ctx context.Context) ([]capabilities.BaseCapability, error)
	Add(ctx context.Context, c capabilities.BaseCapability) error
	Remove(ctx context.Context, ID string) error
}
