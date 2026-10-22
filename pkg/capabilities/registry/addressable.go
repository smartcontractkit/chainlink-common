// This file holds the registry interfaces that split resolution from registration, and the one
// that answers what OCR configuration a capability runs under.
//
// They live here rather than in types/core for the same reason CapabilitiesRegistry does: types/core
// aliases what this package defines, so anything a registry implementation needs has to be declared
// on this side of that edge or the two packages import each other.

package registry

import (
	"context"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

// ReadOnlyBase is the half of the registry that resolves capabilities.
//
// It is separate from registration because how a capability is registered depends
// on where it lives, while resolving one never does. A registry that holds
// capability values registers them by value; a registry that holds addresses
// registers them by address. Both resolve identically, so callers that only look
// capabilities up should depend on this and stay indifferent to which they have.
type ReadOnlyBase interface {
	GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error)
	Get(ctx context.Context, ID string) (capabilities.BaseCapability, error)
	GetExecutable(ctx context.Context, ID string) (capabilities.ExecutableCapability, error)
	List(ctx context.Context) ([]capabilities.BaseCapability, error)
}

// AddressableRegistryBase is a registry that holds the addresses capabilities are
// served at rather than the capabilities themselves.
//
// It dials what it is told about, so registration names an address instead of
// passing a value. That is the only difference from CapabilitiesRegistryBase;
// resolution is the shared ReadOnlyBase.
type AddressableRegistryBase interface {
	ReadOnlyBase

	// AddAt registers a capability of type capType already served at addr, which
	// must be a grpc.NewClient target reachable from the registry.
	AddAt(ctx context.Context, ID string, capType capabilities.CapabilityType, addr string) error
	Remove(ctx context.Context, ID string) error
}

// AddressableCapabilitiesRegistry is the address-based counterpart of
// CapabilitiesRegistry.
type AddressableCapabilitiesRegistry interface {
	AddressableRegistryBase
	CapabilitiesRegistryMetadata
}

// OCRConfigRegistry is a registry that can say which OCR3 configuration a
// capability runs under.
//
// Kept apart from CapabilitiesRegistry rather than added to it: most registries
// hold capabilities and read no contract, so they have nothing to answer with,
// and requiring them to say so would break every implementation of an interface
// that has many.
//
// What comes back is complete, digest included. The registry stores an OCR3
// configuration without its digest, because the digest covers the configuration
// together with the chain and address it was read from - which is what stops a
// configuration being replayed against another registry - so only whoever read
// the contract can compute it. A capability is given the result rather than the
// contract, since the digest is all its oracle needs.
type OCRConfigRegistry interface {
	// OCRConfig returns the configuration of one OCR instance of a capability
	// on a DON. key selects the instance when a capability runs more than one;
	// empty means its only one.
	OCRConfig(ctx context.Context, capabilityID string, donID uint32, key string) (ocrtypes.ContractConfig, error)
}
