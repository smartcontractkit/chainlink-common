package test

import (
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type staticPluginProvider struct {
	staticService
}

func NewStaticPluginProvider(lggr logger.Logger) staticPluginProvider {
	return staticPluginProvider{staticService{lggr: logger.Named(lggr, "staticPluginProvider")}}
}

func (s staticPluginProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return staticOffchainConfigDigester{}
}

func (s staticPluginProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return staticContractConfigTracker{}
}

func (s staticPluginProvider) ContractTransmitter() libocr.ContractTransmitter {
	return staticContractTransmitter{}
}

func (s staticPluginProvider) ChainReader() types.ChainReader {
	return staticChainReader{}
}

func (s StaticPluginProvider) Codec() types.Codec {
	return staticCodec{}
}
