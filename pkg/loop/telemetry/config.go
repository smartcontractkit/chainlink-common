package telemetry

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

const (
	envTracingEnabled                     = "CL_TRACING_ENABLED"
	envTracingCollectorTarget             = "CL_TRACING_COLLECTOR_TARGET"
	envTracingSamplingRatio               = "CL_TRACING_SAMPLING_RATIO"
	envTracingAttribute                   = "CL_TRACING_ATTRIBUTE_"
	envTracingTLSCertPath                 = "CL_TRACING_TLS_CERT_PATH"
	envTelemetryEnabled                   = "CL_TELEMETRY_ENABLED"
	envTelemetryEndpoint                  = "CL_TELEMETRY_ENDPOINT"
	envTelemetryInsecureConn              = "CL_TELEMETRY_INSECURE_CONNECTION"
	envTelemetryCACertFile                = "CL_TELEMETRY_CA_CERT_FILE"
	envTelemetryAttribute                 = "CL_TELEMETRY_ATTRIBUTE_"
	envTelemetryTraceSampleRatio          = "CL_TELEMETRY_TRACE_SAMPLE_RATIO"
	envTelemetryAuthHeader                = "CL_TELEMETRY_AUTH_HEADER"
	envTelemetryAuthPubKeyHex             = "CL_TELEMETRY_AUTH_PUB_KEY_HEX"
	envTelemetryAuthHeadersTTL            = "CL_TELEMETRY_AUTH_HEADERS_TTL"
	envTelemetryEmitterBatchProcessor     = "CL_TELEMETRY_EMITTER_BATCH_PROCESSOR"
	envTelemetryEmitterExportTimeout      = "CL_TELEMETRY_EMITTER_EXPORT_TIMEOUT"
	envTelemetryEmitterExportInterval     = "CL_TELEMETRY_EMITTER_EXPORT_INTERVAL"
	envTelemetryEmitterExportMaxBatchSize = "CL_TELEMETRY_EMITTER_EXPORT_MAX_BATCH_SIZE"
	envTelemetryEmitterMaxQueueSize       = "CL_TELEMETRY_EMITTER_MAX_QUEUE_SIZE"
	envTelemetryLogStreamingEnabled       = "CL_TELEMETRY_LOG_STREAMING_ENABLED"
	envTelemetryLogLevel                  = "CL_TELEMETRY_LOG_LEVEL"
	envTelemetryLogBatchProcessor         = "CL_TELEMETRY_LOG_BATCH_PROCESSOR"
	envTelemetryLogExportTimeout          = "CL_TELEMETRY_LOG_EXPORT_TIMEOUT"
	envTelemetryLogExportMaxBatchSize     = "CL_TELEMETRY_LOG_EXPORT_MAX_BATCH_SIZE"
	envTelemetryLogExportInterval         = "CL_TELEMETRY_LOG_EXPORT_INTERVAL"
	envTelemetryLogMaxQueueSize           = "CL_TELEMETRY_LOG_MAX_QUEUE_SIZE"
	envTelemetryTraceCompressor           = "CL_TELEMETRY_TRACE_COMPRESSOR"
	envTelemetryMetricCompressor          = "CL_TELEMETRY_METRIC_COMPRESSOR"
	envTelemetryMetricExportBatchSize     = "CL_TELEMETRY_METRIC_EXPORT_BATCH_SIZE"
	envTelemetryMetricCardinalityLimit    = "CL_TELEMETRY_METRIC_CARDINALITY_LIMIT"
	envTelemetryMetricViewsDenyAttributes = "CL_TELEMETRY_METRIC_VIEWS_DENY_ATTRIBUTES"
	envTelemetryPrometheusBridgeEnabled   = "CL_TELEMETRY_PROMETHEUS_BRIDGE_ENABLED"
	envTelemetryPrometheusBridgePrefixes  = "CL_TELEMETRY_PROMETHEUS_BRIDGE_PREFIXES"
	envTelemetryLogCompressor             = "CL_TELEMETRY_LOG_COMPRESSOR"
	envChipIngressEndpoint                = "CL_CHIP_INGRESS_ENDPOINT"
	envChipIngressInsecureConnection      = "CL_CHIP_INGRESS_INSECURE_CONNECTION"
	envChipIngressBatchEmitterEnabled     = "CL_CHIP_INGRESS_BATCH_EMITTER_ENABLED"
	envChipIngressDurableEmitterEnabled   = "CL_CHIP_INGRESS_DURABLE_EMITTER_ENABLED"
	envChipIngressBufferSize              = "CL_CHIP_INGRESS_BUFFER_SIZE"
	envChipIngressMaxBatchSize            = "CL_CHIP_INGRESS_MAX_BATCH_SIZE"
	envChipIngressMaxConcurrentSends      = "CL_CHIP_INGRESS_MAX_CONCURRENT_SENDS"
	envChipIngressSendInterval            = "CL_CHIP_INGRESS_SEND_INTERVAL"
	envChipIngressSendTimeout             = "CL_CHIP_INGRESS_SEND_TIMEOUT"
	envChipIngressDrainTimeout            = "CL_CHIP_INGRESS_DRAIN_TIMEOUT"
	envChipIngressMaxGRPCRequestSize      = "CL_CHIP_INGRESS_MAX_GRPC_REQUEST_SIZE"
)

// TelemetryConfig is the tracing, beholder and chip ingress part of loop.EnvConfig. Beholder is only
// set up when TelemetryEndpoint is set; otherwise tracing, if enabled, is set up on its own.
type TelemetryConfig struct {
	TracingEnabled         bool
	TracingCollectorTarget string
	TracingSamplingRatio   float64
	TracingTLSCertPath     string

	// The `toml` tag is singular here to match what AsCmdEnv emits.
	// This is for compatibility with x/config/cli
	TracingAttributes map[string]string `toml:"TracingAttribute"`

	TelemetryEnabled            bool
	TelemetryEndpoint           string
	TelemetryInsecureConnection bool
	TelemetryCACertFile         string
	// The `toml` tag is singular here to match what AsCmdEnv emits.
	// This is for compatibility with x/config/cli
	TelemetryAttributes       beholder.OtelAttributes `toml:"TelemetryAttribute"`
	TelemetryTraceSampleRatio float64
	// The `toml` tag is singular here to match what AsCmdEnv emits.
	// This is for compatibility with x/config/cli
	TelemetryAuthHeaders    map[string]string `toml:"TelemetryAuthHeader"`
	TelemetryAuthPubKeyHex  string
	TelemetryAuthHeadersTTL time.Duration
	// TelemetryEmitterBatchProcessor maps to beholder Config.EmitterBatchProcessor
	// (batched async custom-message export vs immediate per-record export).
	TelemetryEmitterBatchProcessor     bool
	TelemetryEmitterExportTimeout      time.Duration
	TelemetryEmitterExportInterval     time.Duration
	TelemetryEmitterExportMaxBatchSize int
	TelemetryEmitterMaxQueueSize       int
	TelemetryLogStreamingEnabled       bool
	TelemetryLogLevel                  zapcore.Level
	TelemetryLogBatchProcessor         bool
	TelemetryLogExportTimeout          time.Duration
	TelemetryLogExportMaxBatchSize     int
	TelemetryLogExportInterval         time.Duration
	TelemetryLogMaxQueueSize           int
	TelemetryTraceCompressor           string
	TelemetryMetricCompressor          string
	// TelemetryMetricExportBatchSize limits each exporter call to that many
	// data points. Zero (the default) disables batching.
	TelemetryMetricExportBatchSize int
	// TelemetryMetricCardinalityLimit is nil when unset, so AsCmdEnv can
	// distinguish "no opinion" (child applies its own default) from an
	// explicit 0, which disables the limit.
	TelemetryMetricCardinalityLimit *int
	// TelemetryMetricViewsDenyAttributes lists attribute keys dropped by the
	// default global deny view (e.g. event_id). Empty skips Default().
	TelemetryMetricViewsDenyAttributes []string
	TelemetryPrometheusBridgeEnabled   bool
	TelemetryPrometheusBridgePrefixes  []string
	TelemetryLogCompressor             string

	ChipIngressEndpoint              string
	ChipIngressInsecureConnection    bool
	ChipIngressBatchEmitterEnabled   bool
	ChipIngressDurableEmitterEnabled bool

	ChipIngressBufferSize         uint
	ChipIngressMaxBatchSize       uint
	ChipIngressMaxConcurrentSends int
	ChipIngressSendInterval       time.Duration
	ChipIngressSendTimeout        time.Duration
	ChipIngressDrainTimeout       time.Duration
	ChipIngressMaxGRPCRequestSize int
}

// AsCmdEnv returns c as environment variable key/value pairs for an exec.Cmd.
func (c *TelemetryConfig) AsCmdEnv() []string {
	env := []string{}
	add := func(k, v string) {
		env = append(env, k+"="+v)
	}

	add(envTracingEnabled, strconv.FormatBool(c.TracingEnabled))
	add(envTracingCollectorTarget, c.TracingCollectorTarget)
	add(envTracingSamplingRatio, strconv.FormatFloat(c.TracingSamplingRatio, 'f', -1, 64))
	add(envTracingTLSCertPath, c.TracingTLSCertPath)

	for k, v := range c.TracingAttributes {
		add(envTracingAttribute+k, v)
	}

	add(envTelemetryEnabled, strconv.FormatBool(c.TelemetryEnabled))
	add(envTelemetryEndpoint, c.TelemetryEndpoint)
	add(envTelemetryInsecureConn, strconv.FormatBool(c.TelemetryInsecureConnection))
	add(envTelemetryCACertFile, c.TelemetryCACertFile)
	add(envTelemetryTraceSampleRatio, strconv.FormatFloat(c.TelemetryTraceSampleRatio, 'f', -1, 64))
	for k, v := range c.TelemetryAttributes {
		add(envTelemetryAttribute+k, v)
	}

	for k, v := range c.TelemetryAuthHeaders {
		add(envTelemetryAuthHeader+k, v)
	}
	add(envTelemetryAuthPubKeyHex, c.TelemetryAuthPubKeyHex)
	add(envTelemetryAuthHeadersTTL, c.TelemetryAuthHeadersTTL.String())
	add(envTelemetryEmitterBatchProcessor, strconv.FormatBool(c.TelemetryEmitterBatchProcessor))
	add(envTelemetryEmitterExportTimeout, c.TelemetryEmitterExportTimeout.String())
	add(envTelemetryEmitterExportInterval, c.TelemetryEmitterExportInterval.String())
	add(envTelemetryEmitterExportMaxBatchSize, strconv.Itoa(c.TelemetryEmitterExportMaxBatchSize))
	add(envTelemetryEmitterMaxQueueSize, strconv.Itoa(c.TelemetryEmitterMaxQueueSize))
	add(envTelemetryLogStreamingEnabled, strconv.FormatBool(c.TelemetryLogStreamingEnabled))
	add(envTelemetryLogLevel, c.TelemetryLogLevel.String())
	add(envTelemetryLogBatchProcessor, strconv.FormatBool(c.TelemetryLogBatchProcessor))
	add(envTelemetryLogExportTimeout, c.TelemetryLogExportTimeout.String())
	add(envTelemetryLogExportMaxBatchSize, strconv.Itoa(c.TelemetryLogExportMaxBatchSize))
	add(envTelemetryLogExportInterval, c.TelemetryLogExportInterval.String())
	add(envTelemetryLogMaxQueueSize, strconv.Itoa(c.TelemetryLogMaxQueueSize))
	add(envTelemetryTraceCompressor, c.TelemetryTraceCompressor)
	add(envTelemetryMetricCompressor, c.TelemetryMetricCompressor)
	add(envTelemetryMetricExportBatchSize, strconv.Itoa(c.TelemetryMetricExportBatchSize))
	if c.TelemetryMetricCardinalityLimit != nil {
		add(envTelemetryMetricCardinalityLimit, strconv.Itoa(*c.TelemetryMetricCardinalityLimit))
	}
	if len(c.TelemetryMetricViewsDenyAttributes) > 0 {
		add(envTelemetryMetricViewsDenyAttributes, strings.Join(c.TelemetryMetricViewsDenyAttributes, ","))
	}
	add(envTelemetryPrometheusBridgeEnabled, strconv.FormatBool(c.TelemetryPrometheusBridgeEnabled))
	add(envTelemetryPrometheusBridgePrefixes, strings.Join(c.TelemetryPrometheusBridgePrefixes, ","))
	add(envTelemetryLogCompressor, c.TelemetryLogCompressor)

	add(envChipIngressEndpoint, c.ChipIngressEndpoint)
	add(envChipIngressInsecureConnection, strconv.FormatBool(c.ChipIngressInsecureConnection))
	add(envChipIngressBatchEmitterEnabled, strconv.FormatBool(c.ChipIngressBatchEmitterEnabled))
	add(envChipIngressDurableEmitterEnabled, strconv.FormatBool(c.ChipIngressDurableEmitterEnabled))
	add(envChipIngressBufferSize, strconv.FormatUint(uint64(c.ChipIngressBufferSize), 10))
	add(envChipIngressMaxBatchSize, strconv.FormatUint(uint64(c.ChipIngressMaxBatchSize), 10))
	add(envChipIngressMaxConcurrentSends, strconv.Itoa(c.ChipIngressMaxConcurrentSends))
	add(envChipIngressSendInterval, c.ChipIngressSendInterval.String())
	add(envChipIngressSendTimeout, c.ChipIngressSendTimeout.String())
	add(envChipIngressDrainTimeout, c.ChipIngressDrainTimeout.String())
	add(envChipIngressMaxGRPCRequestSize, strconv.Itoa(c.ChipIngressMaxGRPCRequestSize))

	return env
}

// Parse reads c from the environment variables AsCmdEnv writes.
func (c *TelemetryConfig) Parse() error {
	var err error
	c.TracingEnabled, err = getBool(envTracingEnabled)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envTracingEnabled, err)
	}

	if c.TracingEnabled {
		c.TracingCollectorTarget, err = getValidCollectorTarget()
		if err != nil {
			return err
		}
		c.TracingAttributes = getMap(envTracingAttribute)
		c.TracingSamplingRatio = getFloat64OrZero(envTracingSamplingRatio)
		c.TracingTLSCertPath = os.Getenv(envTracingTLSCertPath)
	}

	c.TelemetryEnabled, err = getBool(envTelemetryEnabled)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envTelemetryEnabled, err)
	}

	if c.TelemetryEnabled {
		c.TelemetryEndpoint = os.Getenv(envTelemetryEndpoint)
		c.TelemetryInsecureConnection, err = getBool(envTelemetryInsecureConn)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEndpoint, err)
		}
		c.TelemetryCACertFile = os.Getenv(envTelemetryCACertFile)
		c.TelemetryAttributes = getMap(envTelemetryAttribute)
		c.TelemetryTraceSampleRatio = getFloat64OrZero(envTelemetryTraceSampleRatio)
		c.TelemetryAuthHeaders = getMap(envTelemetryAuthHeader)
		c.TelemetryAuthPubKeyHex = os.Getenv(envTelemetryAuthPubKeyHex)
		c.TelemetryAuthHeadersTTL, err = getDuration(envTelemetryAuthHeadersTTL)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryAuthHeadersTTL, err)
		}
		c.TelemetryEmitterBatchProcessor, err = getBool(envTelemetryEmitterBatchProcessor)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterBatchProcessor, err)
		}
		c.TelemetryEmitterExportTimeout, err = getDuration(envTelemetryEmitterExportTimeout)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterExportTimeout, err)
		}
		c.TelemetryEmitterExportInterval, err = getDuration(envTelemetryEmitterExportInterval)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterExportInterval, err)
		}
		c.TelemetryEmitterExportMaxBatchSize, err = getInt(envTelemetryEmitterExportMaxBatchSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterExportMaxBatchSize, err)
		}
		c.TelemetryEmitterMaxQueueSize, err = getInt(envTelemetryEmitterMaxQueueSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterMaxQueueSize, err)
		}
		c.TelemetryLogStreamingEnabled, err = getBool(envTelemetryLogStreamingEnabled)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogStreamingEnabled, err)
		}
		logLevelStr := os.Getenv(envTelemetryLogLevel)
		if logLevelStr == "" {
			logLevelStr = "info" // Default log level
		}
		var logLevel zapcore.Level
		if err := logLevel.Set(logLevelStr); err != nil {
			logLevel = zapcore.InfoLevel // Fallback to info level on invalid input
		}
		c.TelemetryLogLevel = logLevel
		c.TelemetryLogBatchProcessor, err = getBool(envTelemetryLogBatchProcessor)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogBatchProcessor, err)
		}
		c.TelemetryLogExportTimeout, err = getDuration(envTelemetryLogExportTimeout)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogExportTimeout, err)
		}
		c.TelemetryLogExportMaxBatchSize, err = getInt(envTelemetryLogExportMaxBatchSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogExportMaxBatchSize, err)
		}
		c.TelemetryLogExportInterval, err = getDuration(envTelemetryLogExportInterval)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogExportInterval, err)
		}
		c.TelemetryLogMaxQueueSize, err = getInt(envTelemetryLogMaxQueueSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogMaxQueueSize, err)
		}
		c.TelemetryTraceCompressor = os.Getenv(envTelemetryTraceCompressor)
		c.TelemetryMetricCompressor = os.Getenv(envTelemetryMetricCompressor)
		c.TelemetryMetricExportBatchSize, err = getInt(envTelemetryMetricExportBatchSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryMetricExportBatchSize, err)
		}
		if v, ok := os.LookupEnv(envTelemetryMetricCardinalityLimit); ok {
			limit, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", envTelemetryMetricCardinalityLimit, err)
			}
			if limit < 0 {
				return fmt.Errorf("failed to parse %s: value %d must not be negative (0 disables the limit)", envTelemetryMetricCardinalityLimit, limit)
			}
			c.TelemetryMetricCardinalityLimit = &limit
		} else {
			defaultLimit := beholder.DefaultMetricCardinalityLimit
			c.TelemetryMetricCardinalityLimit = &defaultLimit
		}
		if v, ok := os.LookupEnv(envTelemetryMetricViewsDenyAttributes); ok && v != "" {
			c.TelemetryMetricViewsDenyAttributes = splitCommaTrimmed(v)
		}
		c.TelemetryPrometheusBridgeEnabled, err = getBool(envTelemetryPrometheusBridgeEnabled)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryPrometheusBridgeEnabled, err)
		}
		c.TelemetryPrometheusBridgePrefixes = strings.Split(os.Getenv(envTelemetryPrometheusBridgePrefixes), ",")
		c.TelemetryLogCompressor = os.Getenv(envTelemetryLogCompressor)
		// Optional
		c.ChipIngressEndpoint = os.Getenv(envChipIngressEndpoint)
		c.ChipIngressInsecureConnection, err = getBool(envChipIngressInsecureConnection)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressInsecureConnection, err)
		}
		c.ChipIngressBatchEmitterEnabled, err = getBool(envChipIngressBatchEmitterEnabled)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressBatchEmitterEnabled, err)
		}
		c.ChipIngressDurableEmitterEnabled, err = getBool(envChipIngressDurableEmitterEnabled)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressDurableEmitterEnabled, err)
		}
		c.ChipIngressBufferSize, err = getUint(envChipIngressBufferSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressBufferSize, err)
		}
		c.ChipIngressMaxBatchSize, err = getUint(envChipIngressMaxBatchSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressMaxBatchSize, err)
		}
		c.ChipIngressMaxConcurrentSends, err = getInt(envChipIngressMaxConcurrentSends)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressMaxConcurrentSends, err)
		}
		c.ChipIngressSendInterval, err = getDuration(envChipIngressSendInterval)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressSendInterval, err)
		}
		c.ChipIngressSendTimeout, err = getDuration(envChipIngressSendTimeout)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressSendTimeout, err)
		}
		c.ChipIngressDrainTimeout, err = getDuration(envChipIngressDrainTimeout)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressDrainTimeout, err)
		}
		c.ChipIngressMaxGRPCRequestSize, err = getInt(envChipIngressMaxGRPCRequestSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressMaxGRPCRequestSize, err)
		}
		if c.ChipIngressMaxGRPCRequestSize < 0 {
			return fmt.Errorf("failed to parse %s: value %d must not be negative", envChipIngressMaxGRPCRequestSize, c.ChipIngressMaxGRPCRequestSize)
		}
	}

	return nil
}

// getValidCollectorTarget validates TRACING_COLLECTOR_TARGET as a URL.
func getValidCollectorTarget() (string, error) {
	tracingCollectorTarget := os.Getenv(envTracingCollectorTarget)
	_, err := url.ParseRequestURI(tracingCollectorTarget)
	if err != nil {
		return "", fmt.Errorf("invalid %s: %w", envTracingCollectorTarget, err)
	}
	return tracingCollectorTarget, nil
}

func splitCommaTrimmed(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func getMap(envKeyPrefix string) map[string]string {
	m := make(map[string]string)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, envKeyPrefix) {
			key, value, found := strings.Cut(env, "=")
			if found {
				key = strings.TrimPrefix(key, envKeyPrefix)
				m[key] = value
			}
		}
	}
	return m
}

// Any errors in parsing result in a sampling ratio of 0.0.
func getFloat64OrZero(envKey string) float64 {
	s := os.Getenv(envKey)
	if s == "" {
		return 0.0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return f
}

func getUint(envKey string) (uint, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return 0, nil
	}
	u, err := strconv.ParseUint(s, 10, strconv.IntSize)
	if err != nil {
		return 0, err
	}
	return uint(u), nil
}

func getBool(envKey string) (bool, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return false, nil
	}
	return strconv.ParseBool(s)
}

func getDuration(envKey string) (time.Duration, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}

func getInt(envKey string) (int, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return 0, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return i, nil
}
