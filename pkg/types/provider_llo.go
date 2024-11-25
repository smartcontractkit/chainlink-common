package types

import (
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/types/llo"
)

type LLOConfigProvider interface {
	OffchainConfigDigester() ocrtypes.OffchainConfigDigester
	// One instance will be run per config tracker
	ContractConfigTrackers() []ocrtypes.ContractConfigTracker
}

type LLOProvider interface {
	Service
	LLOConfigProvider
	ShouldRetireCache() llo.ShouldRetireCache
	ContractTransmitter() llo.Transmitter
	ChannelDefinitionCache() llo.ChannelDefinitionCache
}
