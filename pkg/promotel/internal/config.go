package internal

import (
	"errors"
	"fmt"

	"github.com/pkcll/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

type (
	ReceiverConfig = prometheusreceiver.Config
	ExporterConfig = otlpexporter.Config
)

func NewReceiverConfig() (*ReceiverConfig, error) {
	factory := otlpexporter.NewFactory()
	cfg, ok := factory.CreateDefaultConfig().(*prometheusreceiver.Config)
	if !ok {
		return &prometheusreceiver.Config{}, errors.New("failed to cast config to prometheusreceiver.Config")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewDefaultExporterConfig() (*ExporterConfig, error) {
	factory := otlpexporter.NewFactory()
	cfg, ok := factory.CreateDefaultConfig().(*otlpexporter.Config)
	if !ok {
		return &otlpexporter.Config{}, errors.New("failed to cast config to otlpexporter.Config")
	}
	cfg.ClientConfig.Endpoint = "localhost:4317"
	cfg.ClientConfig.TLSSetting.Insecure = true
	return cfg, nil
}

func NewMetricExporterConfig(endpoint string, TLSInsecure bool, headers map[string]string) (*ExporterConfig, error) {
	cfg, err := NewDefaultExporterConfig()
	if err != nil {
		return cfg, err
	}
	cfg.ClientConfig.Endpoint = endpoint
	h := make(map[string]configopaque.String)
	for k, v := range headers {
		h[k] = configopaque.String(v)
	}
	cfg.ClientConfig.Headers = h
	cfg.ClientConfig.TLSSetting.Insecure = TLSInsecure
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func TestReceiverConfig(rawConf map[string]any) (*ReceiverConfig, error) {
	factory := prometheusreceiver.NewFactory()
	cfg := confmap.NewFromStringMap(rawConf)
	// Creates a default configuration for the receiver
	config := factory.CreateDefaultConfig()
	// Merges the configuration into the default config
	if err := cfg.Unmarshal(config); err != nil {
		return nil, err
	}
	c, ok := config.(*prometheusreceiver.Config)
	if !ok {
		return &prometheusreceiver.Config{}, fmt.Errorf("failed to cast config to otlpexporter.Config")
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func TestExporterConfig(rawConf map[string]any) (*ExporterConfig, error) {
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
	c, ok := config.(*otlpexporter.Config)
	if !ok {
		return &otlpexporter.Config{}, fmt.Errorf("failed to cast config to otlpexporter.Config")
	}
	return c, nil
}

func TestDefaultExporterConfig() (*ExporterConfig, error) {
	return TestExporterConfig(map[string]any{
		"endpoint": "localhost:4317",
		"tls": map[string]any{
			"insecure": true,
		},
	})
}
