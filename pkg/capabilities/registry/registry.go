package registry

import (
	"context"

	"github.com/smartcontractkit/libocr/ragep2p/types"

	"errors"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type CapabilitiesRegistry interface {
	CapabilitiesRegistryBase
	CapabilitiesRegistryMetadata
}

type CapabilitiesRegistryMetadata interface {
	LocalNode(ctx context.Context) (capabilities.Node, error)
	NodeByPeerID(ctx context.Context, peerID types.PeerID) (capabilities.Node, error)
	ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error)
	DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error)
	// DONByID resolves a DON by its registry ID. Unlike DONsForCapability, this
	// resolves any DON known to the registry (including caller/workflow DONs that
	// do not host a given capability), which is required to authoritatively read
	// a caller DON's Families (e.g. zone membership) from its WorkflowDonID.
	DONByID(ctx context.Context, donID uint32) (capabilities.DON, error)
}

type CapabilitiesRegistryBase interface {
	GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error)
	Get(ctx context.Context, ID string) (capabilities.BaseCapability, error)
	GetExecutable(ctx context.Context, ID string) (capabilities.ExecutableCapability, error)
	List(ctx context.Context) ([]capabilities.BaseCapability, error)
	Add(ctx context.Context, c capabilities.BaseCapability) error
	Remove(ctx context.Context, ID string) error
}

var _ CapabilitiesRegistry = (*Registry)(nil)

// Registry is a struct for the registry of capabilities.
// Registry is safe for concurrent use.
type Registry struct {
	CapabilitiesRegistryBase

	metadataRegistry CapabilitiesRegistryMetadata
	lggr             logger.Logger
	mu               sync.RWMutex
}

func (r *Registry) LocalNode(ctx context.Context) (capabilities.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.metadataRegistry == nil {
		return capabilities.Node{}, errors.New("metadataRegistry information not available")
	}

	return r.metadataRegistry.LocalNode(ctx)
}

func (r *Registry) NodeByPeerID(ctx context.Context, peerID types.PeerID) (capabilities.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.metadataRegistry == nil {
		return capabilities.Node{}, errors.New("metadataRegistry information not available")
	}
	return r.metadataRegistry.NodeByPeerID(ctx, peerID)
}

func (r *Registry) ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.metadataRegistry == nil {
		return capabilities.CapabilityConfiguration{}, errors.New("metadataRegistry information not available")
	}

	return r.metadataRegistry.ConfigForCapability(ctx, capabilityID, donID)
}

func (r *Registry) DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.metadataRegistry == nil {
		return nil, errors.New("metadataRegistry information not available")
	}

	return r.metadataRegistry.DONsForCapability(ctx, capabilityID)
}

func (r *Registry) DONByID(ctx context.Context, donID uint32) (capabilities.DON, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.metadataRegistry == nil {
		return capabilities.DON{}, errors.New("metadataRegistry information not available")
	}

	return r.metadataRegistry.DONByID(ctx, donID)
}

// SetMetadataRegistry sets a local copy of the offchain registry for the registry to use.
// This is only public for testing purposes; the only production use should be from the CapabilitiesLauncher.
func (r *Registry) SetMetadataRegistry(lr CapabilitiesRegistryMetadata) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metadataRegistry = lr
}

// NewRegistry returns a new Registry.
func NewRegistry(lggr logger.Logger) *Registry {
	return &Registry{
		CapabilitiesRegistryBase: NewBaseRegistry(lggr),
		lggr:                     logger.Named(lggr, "CapabilitiesRegistry"),
	}
}

var _ CapabilitiesRegistryMetadata = (*TestMetadataRegistry)(nil)

// TestMetadataRegistry is a test implementation of the metadataRegistry
// interface. It is used when ExternalCapabilitiesRegistry is not available.
type TestMetadataRegistry struct {
	// WorkflowDONF allows local CRE to override the synthetic workflow DON fault
	// tolerance for compatibility paths that still expect a multi-signer shape.
	WorkflowDONF uint8
}

const (
	testWorkflowDONID            = 1
	testWorkflowDONConfigVersion = 1
)

func (t *TestMetadataRegistry) LocalNode(ctx context.Context) (capabilities.Node, error) {
	peerID := types.PeerID{}
	return capabilities.Node{
		PeerID:         &peerID,
		WorkflowDON:    newTestWorkflowDON(peerID, t.WorkflowDONF),
		CapabilityDONs: []capabilities.DON{},
	}, nil
}

func newTestWorkflowDON(peerID types.PeerID, faultTolerance uint8) capabilities.DON {
	return capabilities.DON{
		ID:            testWorkflowDONID,
		ConfigVersion: testWorkflowDONConfigVersion,
		Members: []types.PeerID{
			peerID,
		},
		F:                faultTolerance,
		IsPublic:         false,
		AcceptsWorkflows: true,
	}
}

func (t *TestMetadataRegistry) NodeByPeerID(ctx context.Context, _ types.PeerID) (capabilities.Node, error) {
	return t.LocalNode(ctx)
}

func (t *TestMetadataRegistry) ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error) {
	return capabilities.CapabilityConfiguration{}, nil
}

func (t *TestMetadataRegistry) DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	return []capabilities.DONWithNodes{}, nil
}

func (t *TestMetadataRegistry) DONByID(ctx context.Context, donID uint32) (capabilities.DON, error) {
	return capabilities.DON{}, nil
}
