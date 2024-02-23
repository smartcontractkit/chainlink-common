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

type PluginProviderTester interface {
	types.PluginProvider
	// AssertEqual checks that the sub-components of the other PluginProvider are equal to this one
	AssertEqual(t *testing.T, ctx context.Context, other types.PluginProvider)
}

// StaticPluginProvider is a static implementation of PluginProviderTester
type StaticPluginProvider struct {
	offchainConfigDigester staticOffchainConfigDigester
	contractConfigTracker  staticContractConfigTracker
	contractTransmitter    ContractTransmitterEvaluator
	chainReader            ChainReaderEvaluator
	codec                  staticCodec
}

var _ PluginProviderTester = StaticPluginProvider{}

func (s StaticPluginProvider) Start(ctx context.Context) error { return nil }

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

func (s StaticPluginProvider) AssertEqual(t *testing.T, ctx context.Context, provider types.PluginProvider) {
	t.Run("OffchainConfigDigester", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.offchainConfigDigester.Evaluate(ctx, provider.OffchainConfigDigester()))
	})

	t.Run("ContractConfigTracker", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.contractConfigTracker.Evaluate(ctx, provider.ContractConfigTracker()))
	})

	t.Run("ContractTransmitter", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.contractTransmitter.Evaluate(ctx, provider.ContractTransmitter()))
	})

	t.Run("ChainReader", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.chainReader.Evaluate(ctx, provider.ChainReader()))
	})

	/*
		t.Run("Codec", func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, s.codec.Evaluate(ctx, provider.OnchainConfigCodec()))
		})
	*/
}
