package test

import (
	"context"
	"testing"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	ocr2test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr2/test"
	ocr3test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr3/test"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func OCR3CapabilityProvider(lggr logger.Logger) staticPluginProvider {
	lggr = logger.Named(lggr, "staticPluginProvider")
	return staticPluginProvider{
		PluginProviderTester: ocr2test.AgnosticPluginProvider(lggr),
		contractTransmitter:  ocr3test.ContractTransmitter,
	}
}

var _ types.PluginProvider = staticPluginProvider{}
var _ testtypes.OCR3CapabilityProviderTester = staticPluginProvider{}

// staticPluginProvider is a static implementation of PluginProviderTester
type staticPluginProvider struct {
	testtypes.PluginProviderTester
	contractTransmitter testtypes.OCR3ContractTransmitterEvaluator
}

func (s staticPluginProvider) OCR3ContractTransmitter() ocr3types.ContractTransmitter[[]byte] {
	return s.contractTransmitter
}

func (s staticPluginProvider) AssertEqual(ctx context.Context, t *testing.T, provider types.OCR3CapabilityProvider) {
	s.PluginProviderTester.AssertEqual(ctx, t, provider)

	t.Run("OCR3ContractTransmitter", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.contractTransmitter.Evaluate(ctx, provider.OCR3ContractTransmitter()))
	})
}

func (s staticPluginProvider) Evaluate(ctx context.Context, provider types.OCR3CapabilityProvider) error {
	err := s.PluginProviderTester.Evaluate(ctx, provider)
	if err != nil {
		return err
	}

	err = s.contractTransmitter.Evaluate(ctx, provider.OCR3ContractTransmitter())
	if err != nil {
		return err
	}

	return nil
}
