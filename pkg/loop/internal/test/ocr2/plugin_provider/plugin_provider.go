package pluginprovider_test

import (
	"context"
	"testing"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.PluginProvider = StaticPluginProvider{}

var TestPluginProvider = StaticPluginProvider{
	offchainConfigDigester: TestOffchainConfigDigester,
	contractConfigTracker:  TestContractConfigTracker,
	contractTransmitter:    TestContractTransmitter,
	chainReader:            TestChainReader,
	codec:                  staticCodec{},
}

type StaticPluginProvider struct {
	offchainConfigDigester StaticOffchainConfigDigester //libocr.OffchainConfigDigester
	contractConfigTracker  StaticContractConfigTracker  //libocr.ContractConfigTracker
	contractTransmitter    StaticContractTransmitter    //libocr.ContractTransmitter
	chainReader            ChainReaderTester            //StaticChainReader            //types.ChainReader
	codec                  staticCodec                  //types.Codec
}

func (s StaticPluginProvider) Start(ctx context.Context) error {
	// todo lazy initialization?
	/*
		s.offchainConfigDigester = TestOffchainConfigDigester
		s.contractConfigTracker = TestContractConfigTracker
		s.contractTransmitter = TestContractTransmitter
		s.chainReader = TestChainReader
	*/
	return nil
}

func (s StaticPluginProvider) Close() error { return nil }

func (s StaticPluginProvider) Ready() error { panic("unimplemented") }

func (s StaticPluginProvider) Name() string { panic("unimplemented") }

func (s StaticPluginProvider) HealthReport() map[string]error { panic("unimplemented") }

func (s StaticPluginProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return s.offchainConfigDigester
}

func (s StaticPluginProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return s.contractConfigTracker
}

func (s StaticPluginProvider) ContractTransmitter() libocr.ContractTransmitter {
	return s.contractTransmitter
}

func (s StaticPluginProvider) ChainReader() types.ChainReader {
	return s.chainReader
}

func (s StaticPluginProvider) Codec() types.Codec {
	return staticCodec{}
}

func (s StaticPluginProvider) AssertEqual(t *testing.T, provider types.PluginProvider) {
	t.Run("OffchainConfigDigester", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.offchainConfigDigester.Equal(provider.OffchainConfigDigester()))
	})

	t.Run("ContractConfigTracker", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.contractConfigTracker.Equal(context.Background(), provider.ContractConfigTracker()))
	})

	t.Run("ContractTransmitter", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.contractTransmitter.Equal(context.Background(), provider.ContractTransmitter()))
	})

	t.Run("ChainReader", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.chainReader.Evaluate(context.Background(), provider.ChainReader()))
	})

	/*
		t.Run("Codec", func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, s.codec.Evaluate(context.Background(), provider.OnchainConfigCodec()))
		})
	*/
}
