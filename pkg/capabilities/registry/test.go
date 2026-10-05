package registry

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/libocr/ragep2p/types"
)

var _ CapabilitiesRegistryMetadata = (*TestRegistryMetadata)(nil)

// TestRegistryMetadata is a test implementation of the metadataRegistry
// interface. It is used when ExternalCapabilitiesRegistry is not available.
type TestRegistryMetadata struct {
	// WorkflowDONF allows local CRE to override the synthetic workflow DON fault
	// tolerance for compatibility paths that still expect a multi-signer shape.
	WorkflowDONF uint8
}

const (
	testWorkflowDONID            = 1
	testWorkflowDONConfigVersion = 1
)

func (t *TestRegistryMetadata) LocalNode(ctx context.Context) (capabilities.Node, error) {
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

func (t *TestRegistryMetadata) NodeByPeerID(ctx context.Context, _ types.PeerID) (capabilities.Node, error) {
	return t.LocalNode(ctx)
}

func (t *TestRegistryMetadata) ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error) {
	return capabilities.CapabilityConfiguration{}, nil
}

func (t *TestRegistryMetadata) DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	return []capabilities.DONWithNodes{}, nil
}

func (t *TestRegistryMetadata) DONByID(ctx context.Context, donID uint32) (capabilities.DON, error) {
	return capabilities.DON{}, nil
}
