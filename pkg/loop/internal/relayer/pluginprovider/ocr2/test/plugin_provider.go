package pluginprovider

import (
	"context"
	"testing"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	chaincomponentstest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.PluginProvider = staticPluginProvider{}

// staticPluginProvider is a static implementation of PluginProviderTester
type staticPluginProvider struct {
	services.Service
	offchainConfigDigester staticOffchainConfigDigester
	contractConfigTracker  staticContractConfigTracker
	contractTransmitter    testtypes.ContractTransmitterEvaluator
	contractReader         testtypes.ContractReaderTester
	codec                  testtypes.CodecEvaluator
}

var _ testtypes.PluginProviderTester = staticPluginProvider{}

func newStaticPluginProvider(lggr logger.Logger) staticPluginProvider {
	lggr = logger.Named(lggr, "staticPluginProvider")
	return staticPluginProvider{
		Service:                test.NewStaticService(lggr),
		offchainConfigDigester: OffchainConfigDigester,
		contractConfigTracker:  ContractConfigTracker,
		contractTransmitter:    ContractTransmitter,
		contractReader:         chaincomponentstest.ContractReader,
		codec:                  chaincomponentstest.Codec,
	}
}

func (s staticPluginProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return s.offchainConfigDigester
}

func (s staticPluginProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return s.contractConfigTracker
}

func (s staticPluginProvider) ContractTransmitter() libocr.ContractTransmitter {
	return s.contractTransmitter
}

func (s staticPluginProvider) ContractReader() types.ContractReader {
	return s.contractReader
}

func (s staticPluginProvider) Codec() types.Codec {
	return s.codec
}

func (s staticPluginProvider) AssertEqual(ctx context.Context, t *testing.T, provider types.PluginProvider) {
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

	t.Run("ContractReader", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.contractReader.Evaluate(ctx, provider.ContractReader()))
	})
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

	err = s.contractReader.Evaluate(ctx, provider.ContractReader())
	if err != nil {
		return err
	}

	return nil
}
