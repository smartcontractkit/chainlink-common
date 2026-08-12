// Package registry is the contract for reading a CapabilitiesRegistry: what a registry says, and
// the interface something that can read one implements.
//
// Only the contract lives here, so that a reader - chainlink-evm reads the on-chain registry with a
// client and the generated bindings - implements it without depending on whatever keeps it in sync
// or serves it. Keeping the snapshot fresh, resolving this node within it and answering metadata
// calls are the caller's business, and live with the binary that does them.
package registry

import (
	"context"

	ragetypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

// DonID is a DON's registry ID.
type DonID uint32

// DON pairs a DON's identity with the raw capability configuration blobs the registry stores for it.
type DON struct {
	capabilities.DON
	// CapabilityConfigurations maps capability ID to the wire-encoded
	// capabilities/pb.CapabilityConfig the registry holds for this DON.
	//
	// The bytes are deliberately left undecoded. A process that only serves them has no reason to
	// interpret them: it hands them over verbatim and the caller unmarshals. That keeps the
	// config-decoding logic, and its drift against the contract, out of the read path entirely.
	CapabilityConfigurations map[string][]byte
}

// NodeInfo is a node's registry record.
type NodeInfo struct {
	NodeOperatorID      uint32
	ConfigCount         uint32
	WorkflowDONID       uint32
	Signer              [32]byte
	P2pID               [32]byte
	EncryptionPublicKey [32]byte
	CsaKey              [32]byte
	CapabilityIDs       []string
}

// Capability is a capability's registry record.
type Capability struct {
	ID             string
	CapabilityType capabilities.CapabilityType
}

// Snapshot is the registry as one read found it. It says nothing about which node is asking: that
// is a property of the process, not of the registry, so a Reader needs no identity - only a way to
// reach wherever the registry is kept.
type Snapshot struct {
	DONs         map[DonID]DON
	Nodes        map[ragetypes.PeerID]NodeInfo
	Capabilities map[string]Capability
}

// Reader reads the registry as it currently stands, wherever it is kept.
//
// This is the only part of using a registry that knows where one lives, which is why it is the only
// part implemented per chain: everything else - when to read, what to do with the result, who may
// see it - is the same whatever the registry is written on.
type Reader interface {
	// Read returns the whole registry as one snapshot. Reading the world rather than following
	// changes is deliberate: a snapshot cannot desync from a reorg or a dropped subscription, so
	// there is no missed-update failure mode to reason about. The cost is a propagation delay
	// bounded by however often the caller reads.
	Read(ctx context.Context) (*Snapshot, error)
}
