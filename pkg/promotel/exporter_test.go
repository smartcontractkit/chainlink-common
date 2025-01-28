package promotel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := otlpexporter.NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	require.NoError(t, componenttest.CheckConfigStruct(cfg))
	ocfg, ok := factory.CreateDefaultConfig().(*otlpexporter.Config)
	assert.True(t, ok)
	assert.Equal(t, configretry.NewDefaultBackOffConfig(), ocfg.RetryConfig)
	assert.Equal(t, exporterhelper.NewDefaultQueueConfig(), ocfg.QueueConfig)
	assert.Equal(t, exporterhelper.NewDefaultTimeoutConfig(), ocfg.TimeoutConfig)
	assert.Equal(t, configcompression.TypeGzip, ocfg.Compression)
}

func TestCreateMetrics(t *testing.T) {
	factory := otlpexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*otlpexporter.Config)
	cfg.ClientConfig.Endpoint = "localhost:4317"

	set := exportertest.NewNopSettings()
	oexp, err := factory.CreateMetrics(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NotNil(t, oexp)
}

func TestMetricExporter(t *testing.T) {
	factory := otlpexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*otlpexporter.Config)
	cfg.ClientConfig.Endpoint = "localhost:4317"

	exporter, err := NewMetricExporter(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, exporter)

	require.NoError(t, exporter.Start(context.Background()))
	require.NoError(t, exporter.Close())
	require.NotNil(t, exporter.Consumer())
}
