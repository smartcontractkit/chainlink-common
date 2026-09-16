package core

import (
	"context"
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	"github.com/smartcontractkit/libocr/ragep2p/types"
)

// Deprecated
type CapabilitiesRegistry = registry.CapabilitiesRegistry

// Deprecated
type CapabilitiesRegistryMetadata = registry.CapabilitiesRegistryMetadata

// Deprecated
type CapabilitiesRegistryBase = registry.CapabilitiesRegistryBase

type UnimplementedCapabilitiesRegistry struct {
	UnimplementedCapabilitiesRegistryMetadata
	UnimplementedCapabilitiesRegistryBase
}

type UnimplementedCapabilitiesRegistryMetadata struct{}

func (UnimplementedCapabilitiesRegistryMetadata) LocalNode(ctx context.Context) (capabilities.Node, error) {
	return capabilities.Node{}, errors.New("LocalNode not implemented")
}

func (UnimplementedCapabilitiesRegistryMetadata) NodeByPeerID(ctx context.Context, peerID types.PeerID) (capabilities.Node, error) {
	return capabilities.Node{}, errors.New("NodeByPeerID not implemented")
}

func (UnimplementedCapabilitiesRegistryMetadata) ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error) {
	return capabilities.CapabilityConfiguration{}, errors.New("ConfigForCapability not implemented")
}

func (UnimplementedCapabilitiesRegistryMetadata) DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	return nil, errors.New("DONsForCapability not implemented")
}

func (UnimplementedCapabilitiesRegistryMetadata) DONByID(ctx context.Context, donID uint32) (capabilities.DON, error) {
	return capabilities.DON{}, errors.New("DONByID not implemented")
}

type UnimplementedCapabilitiesRegistryBase struct {
}

func (UnimplementedCapabilitiesRegistryBase) GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error) {
	return nil, errors.New("GetTrigger not implemented")
}

func (UnimplementedCapabilitiesRegistryBase) Get(ctx context.Context, ID string) (capabilities.BaseCapability, error) {
	return nil, errors.New("Get not implemented")
}

func (UnimplementedCapabilitiesRegistryBase) GetExecutable(ctx context.Context, ID string) (capabilities.ExecutableCapability, error) {
	return nil, errors.New("GetExecutable not implemented")
}

func (UnimplementedCapabilitiesRegistryBase) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	return nil, errors.New("List not implemented")
}

func (UnimplementedCapabilitiesRegistryBase) Add(ctx context.Context, c capabilities.BaseCapability) error {
	return errors.New("Add not implemented")
}

func (UnimplementedCapabilitiesRegistryBase) Remove(ctx context.Context, ID string) error {
	return errors.New("Remove not implemented")
}
