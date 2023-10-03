package types

import ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

// The bootstrap jobs only watch config.
type ConfigProvider interface {
	Service
	OffchainConfigDigester() ocrtypes.OffchainConfigDigester
	ContractConfigTracker() ocrtypes.ContractConfigTracker
}

type ChainReader interface {
	RegisterEventFilter(filterName string, filter EventFilter, startingBlock BlockID) (string, error)
	UnregisterEventFilter(filterName string) error
	QueryEvents(query EventQuery) ([]Event, error)
	GetLatestValue(bc BoundContract, method string, params, returnVal any) error
}

// Plugin is an alias for PluginProvider, for compatibility.
// Deprecated
type Plugin = PluginProvider

// PluginProvider provides common components for any OCR2 plugin.
// It watches config and is able to transmit.
type PluginProvider interface {
	ConfigProvider
	ChainReader() ChainReader
	ContractTransmitter() ocrtypes.ContractTransmitter
}
