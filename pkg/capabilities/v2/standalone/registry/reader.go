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
	"fmt"

	"google.golang.org/protobuf/proto"

	ragetypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
)

// DonID is a DON's registry ID.
type DonID uint32

// DON pairs a DON's identity with the capability configurations the registry stores for it.
type DON struct {
	capabilities.DON
	// CapabilityConfigurations maps capability ID to the configuration the registry holds for that
	// capability on this DON.
	CapabilityConfigurations map[string]CapabilityConfiguration
}

// CapabilityConfiguration is what the registry stores for one capability on one DON: a wire-encoded
// capabilities/pb.CapabilityConfig.
//
// The bytes are deliberately left undecoded by the read path. A process that only serves them has
// no reason to interpret them: it hands them over verbatim and the caller unmarshals. Decoding is
// therefore offered rather than performed - see Unmarshal.
type CapabilityConfiguration struct {
	Config []byte
}

// Unmarshal decodes the stored bytes into the full capability configuration.
//
// This is a method on the contract rather than something each caller writes because the encoding is
// part of what the contract says these bytes are: pb.CapabilityConfigFromProto is the single
// decoder for that message, and routing every caller through it is what keeps a snapshot's idea of
// a config and the wire's from drifting apart.
func (c CapabilityConfiguration) Unmarshal() (capabilities.CapabilityConfiguration, error) {
	cfg := &pb.CapabilityConfig{}
	if err := proto.Unmarshal(c.Config, cfg); err != nil {
		return capabilities.CapabilityConfiguration{}, fmt.Errorf("failed to unmarshal capability configuration: %w", err)
	}
	return pb.CapabilityConfigFromProto(cfg)
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

// Contract identifies the registry a snapshot was read from: the chain it is on, and its address
// in the form that chain writes them.
//
// It travels with the snapshot because some of what the registry stores cannot be interpreted
// without it. An OCR3 configuration is stored without its config digest, since the digest covers
// the configuration together with the chain and address it came from - which is what stops a
// configuration being replayed against another registry - so whoever computes that digest has to
// be told both, and only the reader knows them.
type Contract struct {
	ChainID uint64
	Address string
}

// Snapshot is the registry as one read found it. It says nothing about which node is asking: that
// is a property of the process, not of the registry, so a Reader needs no identity - only a way to
// reach wherever the registry is kept.
type Snapshot struct {
	DONs         map[DonID]DON
	Nodes        map[ragetypes.PeerID]NodeInfo
	Capabilities map[string]Capability

	// Contract is where this snapshot was read from. A Reader that has no contract behind it -
	// one serving a fixed registry in a test, say - leaves it zero.
	Contract Contract
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
