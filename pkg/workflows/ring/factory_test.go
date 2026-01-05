package ring

import (
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/stretchr/testify/require"
)

func TestFactory_NewFactory(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore()
	arbiter := &mockArbiter{}

	t.Run("with_nil_config", func(t *testing.T) {
		f, err := NewFactory(store, arbiter, lggr, nil)
		require.NoError(t, err)
		require.NotNil(t, f)
	})

	t.Run("with_custom_config", func(t *testing.T) {
		cfg := &ConsensusConfig{BatchSize: 50}
		f, err := NewFactory(store, arbiter, lggr, cfg)
		require.NoError(t, err)
		require.NotNil(t, f)
	})

	t.Run("nil_arbiter_returns_error", func(t *testing.T) {
		_, err := NewFactory(store, nil, lggr, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "arbiterScaler is required")
	})
}

func TestFactory_NewReportingPlugin(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore()
	f, err := NewFactory(store, &mockArbiter{}, lggr, nil)
	require.NoError(t, err)

	config := ocr3types.ReportingPluginConfig{N: 4, F: 1}
	plugin, info, err := f.NewReportingPlugin(context.Background(), config)
	require.NoError(t, err)
	require.NotNil(t, plugin)
	require.NotEmpty(t, info.Name)
	require.Equal(t, "RingPlugin", info.Name)
	require.Equal(t, defaultMaxReportCount, info.Limits.MaxReportCount)
}

func TestFactory_Lifecycle(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore()
	f, err := NewFactory(store, &mockArbiter{}, lggr, nil)
	require.NoError(t, err)

	err = f.Start(context.Background())
	require.NoError(t, err)

	name := f.Name()
	require.NotEmpty(t, name)

	report := f.HealthReport()
	require.NotNil(t, report)
	require.Contains(t, report, name)

	err = f.Close()
	require.NoError(t, err)
}
