package pluginprovider_test

import (
	"context"
	"testing"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.PluginProvider = staticPluginProvider{}

var AgnosticProviderImpl = staticPluginProvider{
	offchainConfigDigester: OffchainConfigDigesterImpl,
	contractConfigTracker:  ContractConfigTrackerImpl,
	contractTransmitter:    ContractTransmitterImpl,
	chainReader:            ChainReaderImpl,
	codec:                  staticCodec{},
}

type PluginProviderTester interface {
	types.PluginProvider
	// AssertEqual tests equality of sub-components of the other PluginProvider in parallel
	AssertEqual(ctx context.Context, t *testing.T other types.PluginProvider)
	// Evaluate runs all the method of the other PluginProvider and checks for equality with the embedded PluginProvider
	// it returns the first error encountered
	Evaluate(ctx context.Context, other types.PluginProvider) error
}

// staticPluginProvider is a static implementation of PluginProviderTester
type staticPluginProvider struct {
	offchainConfigDigester staticOffchainConfigDigester
	contractConfigTracker  staticContractConfigTracker
	contractTransmitter    ContractTransmitterEvaluator
	chainReader            ChainReaderEvaluator
	codec                  staticCodec
}

var _ PluginProviderTester = staticPluginProvider{}

func (s staticPluginProvider) Start(ctx context.Context) error { return nil }

func (s staticPluginProvider) Close() error { return nil }

func (s staticPluginProvider) Ready() error { panic("unimplemented") }

func (s staticPluginProvider) Name() string { panic("unimplemented") }

func (s staticPluginProvider) HealthReport() map[string]error { panic("unimplemented") }

func (s staticPluginProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return s.offchainConfigDigester
}

func (s staticPluginProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return s.contractConfigTracker
}

func (s staticPluginProvider) ContractTransmitter() libocr.ContractTransmitter {
	return s.contractTransmitter
}

func (s staticPluginProvider) ChainReader() types.ChainReader {
	return s.chainReader
}

func (s staticPluginProvider) Codec() types.Codec {
	return staticCodec{}
}

func (s staticPluginProvider) AssertEqual(ctx context.Context, t *testing.T provider types.PluginProvider) {
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

func (s staticPluginProvider) Evaluate(ctx context.Context, provider types.PluginProvider) error {
	err := s.offchainConfigDigester.Evaluate(ctx, provider.OffchainConfigDigester())
	if err != nil {
		return err
	}

	err = s.contractConfigTracker.Evaluate(ctx, provider.ContractConfigTracker())
	if err != nil {
		return err
	}

	err = s.contractTransmitter.Evaluate(ctx, provider.ContractTransmitter())
	if err != nil {
		return err
	}

	err = s.chainReader.Evaluate(ctx, provider.ChainReader())
	if err != nil {
		return err
	}

	/*
		err = s.codec.Evaluate(ctx, provider.OnchainConfigCodec())
		if err != nil {
			return err
		}
	*/
	return nil
}
