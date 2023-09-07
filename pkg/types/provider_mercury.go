package types

import (
	"github.com/smartcontractkit/chainlink-relay/pkg/reportingplugins/mercury"
)

// MercuryProvider provides components needed for a mercury OCR2 plugin.
// Mercury requires config tracking but does not transmit on-chain.
type MercuryProvider interface {
	ConfigProvider
	OnchainConfigCodec() mercury.OnchainConfigCodec
	ContractTransmitter() mercury.Transmitter
}
