package core

import (
	"context"
	"errors"

	"github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

type CapabilitiesRegistry interface {
	CapabilitiesRegistryBase
	CapabilitiesRegistryMetadata
}

type CapabilitiesRegistryMetadata interface {
	LocalNode(ctx context.Context) (capabilities.Node, error)
	NodeByPeerID(ctx context.Context, peerID types.PeerID) (capabilities.Node, error)
	ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error)
	DONForCapability(ctx context.Context, capabilityID string) (capabilities.DON, []capabilities.Node, error)
}

type CapabilitiesRegistryBase interface {
	GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error)
	Get(ctx context.Context, ID string) (capabilities.BaseCapability, error)
	GetExecutable(ctx context.Context, ID string) (capabilities.ExecutableCapability, error)
	List(ctx context.Context) ([]capabilities.BaseCapability, error)
	Add(ctx context.Context, c capabilities.BaseCapability) error
	Remove(ctx context.Context, ID string) error
}

var _ CapabilitiesRegistry = UnimplementedCapabilitiesRegistry{}
var _ CapabilitiesRegistryBase = UnimplementedCapabilitiesRegistryBase{}

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

func (UnimplementedCapabilitiesRegistryMetadata) DONForCapability(ctx context.Context, capabilityID string) (capabilities.DON, []capabilities.Node, error) {
	return capabilities.DON{}, nil, errors.New("DONForCapability not implemented")
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
