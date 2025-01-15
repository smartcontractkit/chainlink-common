package test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	validationtest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/validation/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Factory(lggr logger.Logger) staticFactory {
	return newStaticFactory(lggr, StaticFactoryConfig)
}

var StaticFactoryConfig = staticFactoryConfig{
	ReportingPluginConfig: reportingPluginConfig,
	rpi:                   rpi,
	reportingPlugin:       ReportingPlugin,
}

type staticFactoryConfig struct {
	libocr.ReportingPluginConfig
	rpi             libocr.ReportingPluginInfo
	reportingPlugin testtypes.ReportingPluginTester
}

type staticFactory struct {
	services.Service
	staticFactoryConfig
}

func newStaticFactory(lggr logger.Logger, cfg staticFactoryConfig) staticFactory {
	lggr = logger.Named(lggr, "staticFactory")
	return staticFactory{
		Service:             test.NewStaticService(lggr),
		staticFactoryConfig: cfg,
	}
}

var _ types.ReportingPluginFactory = staticFactory{}

func (s staticFactory) NewReportingPlugin(ctx context.Context, config libocr.ReportingPluginConfig) (libocr.ReportingPlugin, libocr.ReportingPluginInfo, error) {
	err := s.equalConfig(config)
	if err != nil {
		return nil, libocr.ReportingPluginInfo{}, fmt.Errorf("config mismatch: %w", err)
	}
	return s.reportingPlugin, s.rpi, nil
}

func (s staticFactory) equalConfig(other libocr.ReportingPluginConfig) error {
	if other.ConfigDigest != s.ConfigDigest {
		return fmt.Errorf("expected ConfigDigest %x but got %x", s.ConfigDigest, other.ConfigDigest)
	}
	if other.OracleID != s.OracleID {
		return fmt.Errorf("expected OracleID %d but got %d", s.OracleID, other.OracleID)
	}
	if other.F != s.F {
		return fmt.Errorf("expected F %d but got %d", s.F, other.F)
	}
	if other.N != s.N {
		return fmt.Errorf("expected N %d but got %d", s.N, other.N)
	}
	if !bytes.Equal(other.OnchainConfig, s.OnchainConfig) {
		return fmt.Errorf("expected OnchainConfig %x but got %x", s.OnchainConfig, other.OnchainConfig)
	}
	if !bytes.Equal(other.OffchainConfig, s.OffchainConfig) {
		return fmt.Errorf("expected OffchainConfig %x but got %x", s.OffchainConfig, other.OffchainConfig)
	}
	if other.EstimatedRoundInterval != s.EstimatedRoundInterval {
		return fmt.Errorf("expected EstimatedRoundInterval %d but got %d", s.EstimatedRoundInterval, other.EstimatedRoundInterval)
	}
	if other.MaxDurationQuery != s.MaxDurationQuery {
		return fmt.Errorf("expected MaxDurationQuery %d but got %d", s.MaxDurationQuery, other.MaxDurationQuery)
	}
	if other.MaxDurationObservation != s.MaxDurationObservation {
		return fmt.Errorf("expected MaxDurationObservation %d but got %d", s.MaxDurationObservation, other.MaxDurationObservation)
	}
	if other.MaxDurationReport != s.MaxDurationReport {
		return fmt.Errorf("expected MaxDurationReport %d but got %d", s.MaxDurationReport, other.MaxDurationReport)
	}
	if other.MaxDurationShouldAcceptFinalizedReport != s.MaxDurationShouldAcceptFinalizedReport {
		return fmt.Errorf("expected MaxDurationShouldAcceptAttestedReport %d but got %d", s.MaxDurationShouldAcceptFinalizedReport, other.MaxDurationShouldAcceptFinalizedReport)
	}
	if other.MaxDurationShouldTransmitAcceptedReport != s.MaxDurationShouldTransmitAcceptedReport {
		return fmt.Errorf("expected MaxDurationShouldTransmitAcceptedReport %d but got %d", s.MaxDurationShouldTransmitAcceptedReport, other.MaxDurationShouldTransmitAcceptedReport)
	}
	return nil
}

func RunFactory(t *testing.T, factory libocr.ReportingPluginFactory) {
	expectedFactory := Factory(logger.Test(t))
	t.Run("ReportingPluginFactory", func(t *testing.T) {
		ctx := tests.Context(t)
		rp, gotRPI, err := factory.NewReportingPlugin(ctx, expectedFactory.ReportingPluginConfig)
		require.NoError(t, err)
		assert.Equal(t, expectedFactory.rpi, gotRPI)
		t.Cleanup(func() { assert.NoError(t, rp.Close()) })
		t.Run("ReportingPlugin", func(t *testing.T) {
			expectedFactory.reportingPlugin.AssertEqual(ctx, t, rp)
		})
	})
}

func RunValidation(t *testing.T, validationService core.ValidationService) {
	ctx := tests.Context(t)
	t.Run("ValidationService", func(t *testing.T) {
		err := validationService.ValidateConfig(ctx, validationtest.GoodPluginConfig)
		require.NoError(t, err)
		err = validationService.ValidateConfig(ctx, nil)
		require.Error(t, err)
	})
}
