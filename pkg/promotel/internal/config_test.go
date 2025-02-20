package internal_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	promModel "github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/exporter/exporterbatcher"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"gopkg.in/yaml.v2"

	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal"
)

func TestReceiverConfig(t *testing.T) {
	configFileName := filepath.Join("testdata", "promconfig.yaml")
	c0, err := LoadTestReceiverConfig(configFileName, "")
	require.NoError(t, err)

	assert.NotNil(t, c0.PrometheusConfig)
	assert.NotNil(t, c0.PrometheusConfig)

	c1, err := LoadTestReceiverConfig(configFileName, "withScrape")
	require.NoError(t, err)

	assert.NotNil(t, c0.PrometheusConfig)

	assert.Len(t, c1.PrometheusConfig.ScrapeConfigs, 1)
	assert.Equal(t, "demo", c1.PrometheusConfig.ScrapeConfigs[0].JobName)
	assert.Equal(t, promModel.Duration(5*time.Second), c1.PrometheusConfig.ScrapeConfigs[0].ScrapeInterval)

	c2, err := LoadTestReceiverConfig(configFileName, "withOnlyScrape")
	require.NoError(t, err)

	assert.Len(t, c2.PrometheusConfig.ScrapeConfigs, 1)
	assert.Equal(t, "demo", c2.PrometheusConfig.ScrapeConfigs[0].JobName)
	assert.Equal(t, promModel.Duration(5*time.Second), c2.PrometheusConfig.ScrapeConfigs[0].ScrapeInterval)
}

func TestUnmarshalDefaultConfig(t *testing.T) {
	factory := otlpexporter.NewFactory()
	cfg := factory.CreateDefaultConfig()
	require.NoError(t, confmap.New().Unmarshal(&cfg))
	assert.Equal(t, factory.CreateDefaultConfig(), cfg)

	cfg, err := internal.TestDefaultExporterConfig()
	require.NoError(t, err)
	assert.Equal(t, "localhost:4317", cfg.(*otlpexporter.Config).ClientConfig.Endpoint)
	assert.True(t, cfg.(*otlpexporter.Config).ClientConfig.TLSSetting.Insecure)

	cfg, err = internal.NewDefaultExporterConfig()
	require.NoError(t, err)
	assert.Equal(t, "localhost:4317", cfg.(*otlpexporter.Config).ClientConfig.Endpoint)
	assert.True(t, cfg.(*otlpexporter.Config).ClientConfig.TLSSetting.Insecure)
}

func TestUnmarshalConfig(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "exporter-config.yaml"))
	require.NoError(t, err)
	cfg, err := internal.TestExporterConfig(cm.ToStringMap())
	require.NoError(t, err)
	assert.Equal(t,
		&otlpexporter.Config{
			TimeoutConfig: exporterhelper.TimeoutConfig{
				Timeout: 10 * time.Second,
			},
			RetryConfig: configretry.BackOffConfig{
				Enabled:             true,
				InitialInterval:     10 * time.Second,
				RandomizationFactor: 0.7,
				Multiplier:          1.3,
				MaxInterval:         1 * time.Minute,
				MaxElapsedTime:      10 * time.Minute,
			},
			QueueConfig: exporterhelper.QueueConfig{
				Enabled:      true,
				NumConsumers: 2,
				QueueSize:    10,
			},
			BatcherConfig: exporterbatcher.Config{
				Enabled:      true,
				FlushTimeout: 200 * time.Millisecond,
				MinSizeConfig: exporterbatcher.MinSizeConfig{
					MinSizeItems: 1000,
				},
				MaxSizeConfig: exporterbatcher.MaxSizeConfig{
					MaxSizeItems: 10000,
				},
			},
			ClientConfig: configgrpc.ClientConfig{
				Headers: map[string]configopaque.String{
					"can you have a . here?": "F0000000-0000-0000-0000-000000000000",
					"header1":                "234",
					"another":                "somevalue",
				},
				Endpoint:    "1.2.3.4:1234",
				Compression: "gzip",
				TLSSetting: configtls.ClientConfig{
					Config: configtls.Config{
						CAFile: "/var/lib/mycert.pem",
					},
					Insecure: false,
				},
				Keepalive: &configgrpc.KeepaliveClientConfig{
					Time:                20 * time.Second,
					PermitWithoutStream: true,
					Timeout:             30 * time.Second,
				},
				WriteBufferSize: 512 * 1024,
				BalancerName:    "round_robin",
				Auth:            &configauth.Authentication{AuthenticatorID: component.MustNewID("nop")},
			},
		}, cfg)
}

// Used for tests
func LoadTestReceiverConfig(fileName string, configName string) (*internal.ReceiverConfig, error) {
	content, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return nil, fmt.Errorf("unable to read the file %v: %w", fileName, err)
	}
	var rawConf map[string]any
	if err = yaml.Unmarshal(content, &rawConf); err != nil {
		return nil, err
	}
	cm := confmap.NewFromStringMap(rawConf)
	componentType := component.MustNewType("prometheus")
	sub, err := cm.Sub(component.NewIDWithName(componentType, configName).String())
	if err != nil {
		return nil, err
	}
	return internal.TestReceiverConfig(sub.ToStringMap())
}
