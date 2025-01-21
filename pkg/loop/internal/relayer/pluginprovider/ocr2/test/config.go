package pluginprovider

import (
	"context"
	"testing"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// ConfigProviderTester is a helper interface for testing ConfigProviders
type ConfigProviderTester interface {
	types.ConfigProvider
	// AssertEqual checks that the sub-components of the other ConfigProvider are equal to this one
	AssertEqual(ctx context.Context, t *testing.T, other types.ConfigProvider)
}

type staticConfigProviderConfig struct {
	offchainDigester      testtypes.OffchainConfigDigesterEvaluator
	contractConfigTracker testtypes.ContractConfigTrackerEvaluator
}

var _ ConfigProviderTester = staticConfigProvider{}

// staticConfigProvider is a static implementation of ConfigProviderTester
type staticConfigProvider struct {
	services.Service
	staticConfigProviderConfig
}

func newStaticConfigProvider(lggr logger.Logger, cfg staticConfigProviderConfig) staticConfigProvider {
	lggr = logger.Named(lggr, "staticConfigProvider")
	return staticConfigProvider{
		Service:                    test.NewStaticService(lggr),
		staticConfigProviderConfig: cfg,
	}
}

func (s staticConfigProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return s.offchainDigester
}

func (s staticConfigProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return s.contractConfigTracker
}

func (s staticConfigProvider) AssertEqual(ctx context.Context, t *testing.T, cp types.ConfigProvider) {
	t.Run("OffchainConfigDigester", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.offchainDigester.Evaluate(ctx, cp.OffchainConfigDigester()))
	})
	t.Run("ContractConfigTracker", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.contractConfigTracker.Evaluate(ctx, cp.ContractConfigTracker()))
	})
}
