package pluginprovider_test

import (
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/stretchr/testify/assert"
)

// ConfigProviderTester is a helper interface for testing ConfigProviders
type ConfigProviderTester interface {
	types.ConfigProvider
	// AssertEqual checks that the sub-components of the other ConfigProvider are equal to this one
	AssertEqual(t *testing.T, ctx context.Context, other types.ConfigProvider)
}

type staticConfigProviderConfig struct {
	offchainDigester      OffchainConfigDigesterEvaluator
	contractConfigTracker ContractConfigTrackerEvaluator
}

// staticConfigProvider is a static implementation of ConfigProviderTester
type staticConfigProvider struct {
	staticConfigProviderConfig
}

var _ ConfigProviderTester = staticConfigProvider{}

// TODO validate start/Close calls?
func (s staticConfigProvider) Start(ctx context.Context) error { return nil }

func (s staticConfigProvider) Close() error { return nil }

func (s staticConfigProvider) Ready() error { panic("unimplemented") }

func (s staticConfigProvider) Name() string { panic("unimplemented") }

func (s staticConfigProvider) HealthReport() map[string]error { panic("unimplemented") }

func (s staticConfigProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return s.offchainDigester
}

func (s staticConfigProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return s.contractConfigTracker
}

func (s staticConfigProvider) AssertEqual(t *testing.T, ctx context.Context, cp types.ConfigProvider) {
	t.Run("OffchainConfigDigester", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.offchainDigester.Evaluate(ctx, cp.OffchainConfigDigester()))
	})
	t.Run("ContractConfigTracker", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.contractConfigTracker.Evaluate(context.Background(), cp.ContractConfigTracker()))
	})
}
