package promotel

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/prometheus/prometheus/discovery"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"gopkg.in/yaml.v3"

	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal/prometheusreceiver"
)

type ReceiverConfig = component.Config
type ExporterConfig = component.Config

func NewReceiverConfig(rawConf map[string]any) (ReceiverConfig, error) {
	factory := prometheusreceiver.NewFactory()

	cfg := confmap.NewFromStringMap(rawConf)
	// Creates a default configuration for the receiver
	config := factory.CreateDefaultConfig()
	// Merges the configuration into the default config
	if err := cfg.Unmarshal(config); err != nil {
		return nil, err
	}
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

func NewDefaultReceiverConfig() (ReceiverConfig, error) {
	return NewReceiverConfig(map[string]any{
		"config": map[string]any{
			"scrape_configs": []map[string]any{{
				"job_name":        "promotel",
				"scrape_interval": "1s",
				"static_configs":  []map[string]any{{"targets": []string{"127.0.0.1:8888"}}},
				"metric_relabel_configs": []map[string]any{{
					"action": "labeldrop",
					"regex":  "service_instance_id|service_name",
				}},
			}},
		},
	})
}

func NewExporterConfig(rawConf map[string]any) (ExporterConfig, error) {
	factory := otlpexporter.NewFactory()

	cfg := confmap.NewFromStringMap(rawConf)
	// Creates a default configuration for the receiver
	config := factory.CreateDefaultConfig()
	// Merges the configuration into the default config
	if err := cfg.Unmarshal(config); err != nil {
		return nil, err
	}
	if err := component.ValidateConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

func NewDefaultExporterConfig() (ExporterConfig, error) {
	return NewExporterConfig(map[string]any{
		"endpoint": "localhost:4317",
		"tls": map[string]any{
			"insecure": true,
		},
	})
}

// Used for tests
func LoadTestConfig(fileName string, configName string) (ReceiverConfig, error) {
	content, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return nil, fmt.Errorf("unable to read the file %v: %w", fileName, err)
	}
	var rawConf map[string]any
	if err = yaml.Unmarshal(content, &rawConf); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	cm := confmap.NewFromStringMap(rawConf)
	componentType := component.MustNewType("prometheus")
	sub, err := cm.Sub(component.NewIDWithName(componentType, configName).String())
	if err != nil {
		return nil, err
	}
	return NewReceiverConfig(sub.ToStringMap())
}

func validateConfig(config component.Config) error {
	if err := component.ValidateConfig(config); err != nil {
		return err
	}
	cfg, ok := config.(*prometheusreceiver.Config)
	if !ok {
		return fmt.Errorf("expected config to be of type *prometheusreceiver.Config, got %T", config)
	}
	if cfg.PrometheusConfig == nil {
		return errors.New("PrometheusConfig is nil")
	}
	for _, scrapeConfig := range cfg.PrometheusConfig.ScrapeConfigs {
		if scrapeConfig.JobName == "" {
			return fmt.Errorf("unexpected job_name: %s", scrapeConfig.JobName)
		}
		if scrapeConfig.ScrapeInterval == 0 {
			return fmt.Errorf("unexpected scrape_interval: %s", scrapeConfig.ScrapeInterval)
		}
		if scrapeConfig.MetricsPath == "" {
			return errors.New("metrics_path is empty")
		}
		for _, cfg := range scrapeConfig.ServiceDiscoveryConfigs {
			staticConfig, ok := cfg.(discovery.StaticConfig)
			if !ok {
				return fmt.Errorf("expected static config, got %T", cfg)
			}
			for _, c := range staticConfig {
				if c.Targets == nil {
					return errors.New("targets is nil")
				}
				if len(c.Targets) == 0 {
					return errors.New("targets is empty")
				}
			}
			if len(staticConfig) == 0 || len(staticConfig[0].Targets) == 0 || staticConfig[0].Targets[0].String() == "" {
				return fmt.Errorf("unexpected targets: %v", staticConfig[0].Targets[0].String())
			}
		}
		for _, relabelConfig := range scrapeConfig.MetricRelabelConfigs {
			if relabelConfig.Action == "" {
				return fmt.Errorf("unexpected action: %s", relabelConfig.Action)
			}
			if relabelConfig.Regex.String() == "" {
				return fmt.Errorf("unexpected regex: %s", relabelConfig.Regex.String())
			}
		}
	}

	return nil
}
