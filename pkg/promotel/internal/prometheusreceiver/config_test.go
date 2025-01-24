package prometheusreceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal/prometheusreceiver/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	sub, err := cm.Sub(component.NewIDWithName(metadata.Type, "").String())
	require.NoError(t, err)
	require.NoError(t, sub.Unmarshal(cfg))

	r0 := cfg.(*Config)
	assert.Equal(t, r0, factory.CreateDefaultConfig())

	sub, err = cm.Sub(component.NewIDWithName(metadata.Type, "customname").String())
	require.NoError(t, err)
	require.NoError(t, sub.Unmarshal(cfg))

	r1 := cfg.(*Config)
	assert.Equal(t, "demo", r1.PrometheusConfig.ScrapeConfigs[0].JobName)
	assert.Equal(t, 5*time.Second, time.Duration(r1.PrometheusConfig.ScrapeConfigs[0].ScrapeInterval))
	assert.True(t, r1.UseStartTimeMetric)
	assert.True(t, r1.TrimMetricSuffixes)
	assert.Equal(t, "^(.+_)*process_start_time_seconds$", r1.StartTimeMetricRegex)
	assert.True(t, r1.ReportExtraScrapeMetrics)
}

func TestValidateConfigWithScrapeConfigFiles(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config_scrape_config_files.yaml"))
	require.NoError(t, err)
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	sub, err := cm.Sub(component.NewIDWithName(metadata.Type, "").String())
	require.NoError(t, err)
	require.NoError(t, sub.Unmarshal(cfg))

	require.NoError(t, component.ValidateConfig(cfg))
}

func TestLoadConfigFailsOnUnknownSection(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "invalid-config-section.yaml"))
	require.NoError(t, err)
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	sub, err := cm.Sub(component.NewIDWithName(metadata.Type, "").String())
	require.NoError(t, err)
	require.Error(t, sub.Unmarshal(cfg))
}

// As one of the config parameters is consuming prometheus
// configuration as a subkey, ensure that invalid configuration
// within the subkey will also raise an error.
func TestLoadConfigFailsOnUnknownPrometheusSection(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "invalid-config-prometheus-section.yaml"))
	require.NoError(t, err)
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	sub, err := cm.Sub(component.NewIDWithName(metadata.Type, "").String())
	require.NoError(t, err)
	require.Error(t, sub.Unmarshal(cfg))
}
