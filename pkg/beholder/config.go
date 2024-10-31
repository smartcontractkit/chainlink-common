package beholder

import (
	"time"

	otelattr "go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	InsecureConnection       bool
	CACertFile               string
	OtelExporterGRPCEndpoint string
	OtelExporterHTTPEndpoint string

	// OTel Resource
	ResourceAttributes []otelattr.KeyValue
	// Message Emitter
	EmitterExportTimeout time.Duration
	// Batch processing is enabled by default
	// Disable it only for testing
	EmitterBatchProcessor bool
	EmitterRetryConfig    *RetryConfig
	// OTel Trace
	TraceSampleRatio  float64
	TraceBatchTimeout time.Duration
	TraceSpanExporter sdktrace.SpanExporter // optional additional exporter
	TraceRetryConfig  *RetryConfig
	// OTel Metric
	MetricReaderInterval time.Duration
	MetricRetryConfig    *RetryConfig
	// OTel Log
	LogExportTimeout time.Duration
	// Batch processing is enabled by default
	// Disable it only for testing
	LogBatchProcessor bool
}

type RetryConfig struct {
	// InitialInterval the time to wait after the first failure before
	// retrying.
	InitialInterval time.Duration
	// MaxInterval is the upper bound on backoff interval. Once this value is
	// reached the delay between consecutive retries will always be
	// `MaxInterval`.
	MaxInterval time.Duration
	// MaxElapsedTime is the maximum amount of time (including retries) spent
	// trying to send a request/batch.  Once this value is reached, the data
	// is discarded.
	// Set to zero to disable retry
	MaxElapsedTime time.Duration
}

// Same defaults as used by the OTel SDK
var defaultRetryConfig = RetryConfig{
	InitialInterval: 5 * time.Second,
	MaxInterval:     30 * time.Second,
	MaxElapsedTime:  1 * time.Minute, // Retry is enabled
}

const (
	defaultPackageName = "beholder"
)

var defaultOtelAttributes = []otelattr.KeyValue{
	otelattr.String("package_name", "beholder"),
}

func DefaultConfig() Config {
	return Config{
		InsecureConnection:       true,
		CACertFile:               "",
		OtelExporterGRPCEndpoint: "localhost:4317",
		// Resource
		ResourceAttributes: defaultOtelAttributes,
		// Message Emitter
		EmitterExportTimeout:  1 * time.Second,
		EmitterBatchProcessor: true,
		// OTel message log exporter retry config
		EmitterRetryConfig: defaultRetryConfig.Copy(),
		// Trace
		TraceSampleRatio:  1,
		TraceBatchTimeout: 1 * time.Second,
		// OTel trace exporter retry config
		TraceRetryConfig: defaultRetryConfig.Copy(),
		// Metric
		MetricReaderInterval: 1 * time.Second,
		// OTel metric exporter retry config
		MetricRetryConfig: defaultRetryConfig.Copy(),
		// Log
		LogExportTimeout:  1 * time.Second,
		LogBatchProcessor: true,
	}
}

func TestDefaultConfig() Config {
	config := DefaultConfig()
	// Should be only disabled for testing
	config.EmitterBatchProcessor = false
	config.LogBatchProcessor = false
	// Retries are disabled for testing
	config.EmitterRetryConfig.MaxElapsedTime = 0 // Retry is disabled
	config.TraceRetryConfig.MaxElapsedTime = 0   // Retry is disabled
	config.MetricRetryConfig.MaxElapsedTime = 0  // Retry is disabled
	return config
}

func TestDefaultConfigHTTPClient() Config {
	config := DefaultConfig()
	// Should be only disabled for testing
	config.EmitterBatchProcessor = false
	config.LogBatchProcessor = false
	config.OtelExporterGRPCEndpoint = ""
	config.OtelExporterHTTPEndpoint = "localhost:4318"
	return config
}

func (c *RetryConfig) Copy() *RetryConfig {
	newConfig := *c
	return &newConfig
}

// Calculate if retry is enabled
func (c *RetryConfig) Enabled() bool {
	if c == nil {
		return false
	}
	return c.InitialInterval > 0 && c.MaxInterval > 0 && c.MaxElapsedTime > 0
}

// Implement getters for fields to avoid nil pointer dereference in case the config is not set
func (c *RetryConfig) GetInitialInterval() time.Duration {
	if c == nil {
		return 0
	}
	return c.InitialInterval
}

func (c *RetryConfig) GetMaxInterval() time.Duration {
	if c == nil {
		return 0
	}
	return c.MaxInterval
}

func (c *RetryConfig) GetMaxElapsedTime() time.Duration {
	if c == nil {
		return 0
	}
	return c.MaxElapsedTime
}
