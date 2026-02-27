package beholder

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	InsecureConnection       bool
	CACertFile               string
	OtelExporterGRPCEndpoint string
	OtelExporterHTTPEndpoint string

	// OTel Resource
	ResourceAttributes []attribute.KeyValue
	// Message Emitter
	EmitterExportTimeout      time.Duration
	EmitterExportInterval     time.Duration
	EmitterExportMaxBatchSize int
	EmitterMaxQueueSize       int
	EmitterBatchProcessor     bool // Enabled by default. Disable only for testing.

	// OTel Trace
	TraceSampleRatio  float64
	TraceBatchTimeout time.Duration
	TraceSpanExporter trace.SpanExporter // optional additional exporter
	TraceRetryConfig  *RetryConfig
	// TraceCompressor sets the gRPC compressor for traces. Valid values: "gzip" (default), "none".
	TraceCompressor string

	// OTel Metric
	MetricReaderInterval time.Duration
	MetricRetryConfig    *RetryConfig
	MetricViews          []metric.View
	// MetricCompressor sets the gRPC compressor for metrics. Valid values: "gzip" (default), "none".
	MetricCompressor string

	// Custom Events via Chip Ingress Emitter
	ChipIngressEmitterEnabled      bool
	ChipIngressEmitterGRPCEndpoint string
	ChipIngressInsecureConnection  bool // Disables TLS for Chip Ingress Emitter

	// Chip Ingress Batch Emitter
	ChipIngressBufferSize   uint          // Per-worker channel buffer size (default 100)
	ChipIngressMaxBatchSize uint          // Max events per PublishBatch call (default 50)
	ChipIngressSendInterval time.Duration // Flush interval per worker (default 500ms when zero or unset)
	ChipIngressSendTimeout  time.Duration // Timeout per PublishBatch call (default 10s)

	// OTel Log
	LogExportTimeout      time.Duration
	LogExportInterval     time.Duration
	LogExportMaxBatchSize int
	LogMaxQueueSize       int
	LogBatchProcessor     bool // Enabled by default. Disable only for testing.
	// Retry config for shared log exporter, used by Emitter and Logger
	LogRetryConfig      *RetryConfig
	LogStreamingEnabled bool          // Enable logs streaming to the OTel log exporter
	LogLevel            zapcore.Level // Log level for telemetry streaming
	// LogCompressor sets the gRPC compressor for logs. Valid values: "gzip" (default), "none".
	LogCompressor string

	// Auth
	// AuthHeaders serves two purposes:
	// 1. Static mode: When AuthKeySigner is nil, these headers are used as-is and never change
	// 2. Rotating mode: When AuthKeySigner is set, these headers are used as initial headers
	//    until TTL expires, then the lazy signer generates new ones
	AuthHeaders      map[string]string
	AuthHeadersTTL   time.Duration
	AuthKeySigner    Signer
	AuthPublicKeyHex string
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

var defaultOtelAttributes = []attribute.KeyValue{
	attribute.String("package_name", "beholder"),
}

func DefaultConfig() Config {
	return Config{
		InsecureConnection:       true,
		CACertFile:               "",
		OtelExporterGRPCEndpoint: "localhost:4317",
		// Resource
		ResourceAttributes: defaultOtelAttributes,
		// Message Emitter
		EmitterExportTimeout:      30 * time.Second,
		EmitterExportMaxBatchSize: 512,
		EmitterExportInterval:     1 * time.Second,
		EmitterMaxQueueSize:       2048,
		EmitterBatchProcessor:     true,
		// OTel message log exporter retry config
		LogRetryConfig: defaultRetryConfig.Copy(),
		// Trace
		TraceSampleRatio:  1,
		TraceBatchTimeout: 1 * time.Second,
		TraceCompressor:   "gzip",
		// OTel trace exporter retry config
		TraceRetryConfig: defaultRetryConfig.Copy(),
		// Metric
		MetricReaderInterval: 1 * time.Second,
		MetricCompressor:     "gzip",
		// OTel metric exporter retry config
		MetricRetryConfig: defaultRetryConfig.Copy(),
		// Log
		LogExportTimeout:      30 * time.Second,
		LogExportMaxBatchSize: 512,
		LogExportInterval:     1 * time.Second,
		LogMaxQueueSize:       2048,
		LogBatchProcessor:     true,
		LogStreamingEnabled:   true, // Enable logs streaming by default
		LogLevel:      zapcore.InfoLevel,
		LogCompressor: "gzip",
		// Chip Ingress Batch Emitter
		ChipIngressBufferSize:   100,
		ChipIngressMaxBatchSize: 50,
		ChipIngressSendInterval: 500 * time.Millisecond,
		ChipIngressSendTimeout:  10 * time.Second,
		// Auth (defaults to static auth mode with TTL=0)
		AuthHeadersTTL: 0,
	}
}

func TestDefaultConfig() Config {
	config := DefaultConfig()
	// Should be only disabled for testing
	config.EmitterBatchProcessor = false
	config.LogBatchProcessor = false
	// Retries are disabled for testing
	config.LogRetryConfig.MaxElapsedTime = 0    // Retry is disabled
	config.TraceRetryConfig.MaxElapsedTime = 0  // Retry is disabled
	config.MetricRetryConfig.MaxElapsedTime = 0 // Retry is disabled
	// Auth disabled for testing (TTL=0 means static auth mode)
	config.AuthHeadersTTL = 0
	return config
}

func TestDefaultConfigHTTPClient() Config {
	config := DefaultConfig()
	// Should be only disabled for testing
	config.EmitterBatchProcessor = false
	config.LogBatchProcessor = false
	config.OtelExporterGRPCEndpoint = ""
	config.OtelExporterHTTPEndpoint = "localhost:4318"
	// Auth disabled for testing (TTL=0 means static auth mode)
	config.AuthHeadersTTL = 0
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

type OtelAttributes map[string]string

func (a OtelAttributes) AsStringAttributes() (attributes []attribute.KeyValue) {
	for k, v := range a {
		attributes = append(attributes, attribute.String(k, v))
	}
	return attributes
}
