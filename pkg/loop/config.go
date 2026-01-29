package loop

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
)

const (
	envAppID = "CL_APP_ID"

	envDatabaseURL                          = "CL_DATABASE_URL"
	envDatabaseIdleInTxSessionTimeout       = "CL_DATABASE_IDLE_IN_TX_SESSION_TIMEOUT"
	envDatabaseLockTimeout                  = "CL_DATABASE_LOCK_TIMEOUT"
	envDatabaseQueryTimeout                 = "CL_DATABASE_QUERY_TIMEOUT"
	envDatabaseListenerFallbackPollInterval = "CL_DATABASE_LISTNER_FALLBACK_POLL_INTERVAL"
	envDatabaseLogSQL                       = "CL_DATABASE_LOG_SQL"
	envDatabaseMaxOpenConns                 = "CL_DATABASE_MAX_OPEN_CONNS"
	envDatabaseMaxIdleConns                 = "CL_DATABASE_MAX_IDLE_CONNS"
	envDatabaseTracingEnabled               = "CL_DATABASE_TRACING_ENABLED"

	envFeatureLogPoller = "CL_FEATURE_LOG_POLLER"

	envMercuryCacheLatestReportDeadline = "CL_MERCURY_CACHE_LATEST_REPORT_DEADLINE"
	envMercuryCacheLatestReportTTL      = "CL_MERCURY_CACHE_LATEST_REPORT_TTL"
	envMercuryCacheMaxStaleAge          = "CL_MERCURY_CACHE_MAX_STALE_AGE"

	envMercuryTransmitterProtocol             = "CL_MERCURY_TRANSMITTER_PROTOCOL"
	envMercuryTransmitterTransmitQueueMaxSize = "CL_MERCURY_TRANSMITTER_TRANSMIT__QUEUE_MAX_SIZE"
	envMercuryTransmitterTransmitTimeout      = "CL_MERCURY_TRANSMITTER_TRANSMIT_TIMEOUT"
	envMercuryTransmitterTransmitConcurrency  = "CL_MERCURY_TRANSMITTER_TRANSMIT_CONCURRENCY"
	envMercuryTransmitterReaperFrequency      = "CL_MERCURY_TRANSMITTER_REAPER_FREQUENCY"
	envMercuryTransmitterReaperMaxAge         = "CL_MERCURY_TRANSMITTER_REAPER_MAX_AGE"
	envMercuryVerboseLogging                  = "CL_MERCURY_VERBOSE_LOGGING"

	envPromPort = "CL_PROMETHEUS_PORT"

	envTracingEnabled         = "CL_TRACING_ENABLED"
	envTracingCollectorTarget = "CL_TRACING_COLLECTOR_TARGET"
	envTracingSamplingRatio   = "CL_TRACING_SAMPLING_RATIO"
	envTracingAttribute       = "CL_TRACING_ATTRIBUTE_"
	envTracingTLSCertPath     = "CL_TRACING_TLS_CERT_PATH"

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
	envTelemetryLogCompressor             = "CL_TELEMETRY_LOG_COMPRESSOR"

	envChipIngressEndpoint           = "CL_CHIP_INGRESS_ENDPOINT"
	envChipIngressInsecureConnection = "CL_CHIP_INGRESS_INSECURE_CONNECTION"

	envCRESettings        = cresettings.EnvNameSettings
	envCRESettingsDefault = cresettings.EnvNameSettingsDefault
)

// EnvConfig is the configuration between the application and the LOOP executable. The values
// are fully resolved and static and passed via the environment.
type EnvConfig struct {
	AppID string

	DatabaseURL                          *config.SecretURL
	DatabaseIdleInTxSessionTimeout       time.Duration
	DatabaseLockTimeout                  time.Duration
	DatabaseQueryTimeout                 time.Duration
	DatabaseListenerFallbackPollInterval time.Duration
	DatabaseLogSQL                       bool
	DatabaseMaxOpenConns                 int
	DatabaseMaxIdleConns                 int
	DatabaseTracingEnabled               bool

	FeatureLogPoller bool

	MercuryCacheLatestReportDeadline time.Duration
	MercuryCacheLatestReportTTL      time.Duration
	MercuryCacheMaxStaleAge          time.Duration

	MercuryTransmitterProtocol             string
	MercuryTransmitterTransmitQueueMaxSize uint32
	MercuryTransmitterTransmitTimeout      time.Duration
	MercuryTransmitterTransmitConcurrency  uint32
	MercuryTransmitterReaperFrequency      time.Duration
	MercuryTransmitterReaperMaxAge         time.Duration
	MercuryVerboseLogging                  bool

	PrometheusPort int

	TracingEnabled         bool
	TracingCollectorTarget string
	TracingSamplingRatio   float64
	TracingTLSCertPath     string
	TracingAttributes      map[string]string

	TelemetryEnabled                   bool
	TelemetryEndpoint                  string
	TelemetryInsecureConnection        bool
	TelemetryCACertFile                string
	TelemetryAttributes                OtelAttributes
	TelemetryTraceSampleRatio          float64
	TelemetryAuthHeaders               map[string]string
	TelemetryAuthPubKeyHex             string
	TelemetryAuthHeadersTTL            time.Duration
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
	TelemetryLogCompressor             string

	ChipIngressEndpoint           string
	ChipIngressInsecureConnection bool

	CRESettings        string
	CRESettingsDefault string
}

// AsCmdEnv returns a slice of environment variable key/value pairs for an exec.Cmd.
func (e *EnvConfig) AsCmdEnv() (env []string) {
	add := func(k, v string) {
		env = append(env, k+"="+v)
	}

	add(envAppID, e.AppID)

	if e.DatabaseURL != nil { // optional
		add(envDatabaseURL, e.DatabaseURL.URL().String())
		add(envDatabaseIdleInTxSessionTimeout, e.DatabaseIdleInTxSessionTimeout.String())
		add(envDatabaseLockTimeout, e.DatabaseLockTimeout.String())
		add(envDatabaseQueryTimeout, e.DatabaseQueryTimeout.String())
		add(envDatabaseListenerFallbackPollInterval, e.DatabaseListenerFallbackPollInterval.String())
		add(envDatabaseLogSQL, strconv.FormatBool(e.DatabaseLogSQL))
		add(envDatabaseMaxOpenConns, strconv.Itoa(e.DatabaseMaxOpenConns))
		add(envDatabaseMaxIdleConns, strconv.Itoa(e.DatabaseMaxIdleConns))
		add(envDatabaseTracingEnabled, strconv.FormatBool(e.DatabaseTracingEnabled))
	}

	add(envFeatureLogPoller, strconv.FormatBool(e.FeatureLogPoller))

	add(envMercuryCacheLatestReportDeadline, e.MercuryCacheLatestReportDeadline.String())
	add(envMercuryCacheLatestReportTTL, e.MercuryCacheLatestReportTTL.String())
	add(envMercuryCacheMaxStaleAge, e.MercuryCacheMaxStaleAge.String())

	add(envMercuryTransmitterProtocol, e.MercuryTransmitterProtocol)
	add(envMercuryTransmitterTransmitQueueMaxSize, strconv.FormatUint(uint64(e.MercuryTransmitterTransmitQueueMaxSize), 10))
	add(envMercuryTransmitterTransmitTimeout, e.MercuryTransmitterTransmitTimeout.String())
	add(envMercuryTransmitterTransmitConcurrency, strconv.FormatUint(uint64(e.MercuryTransmitterTransmitConcurrency), 10))
	add(envMercuryTransmitterReaperFrequency, e.MercuryTransmitterReaperFrequency.String())
	add(envMercuryTransmitterReaperMaxAge, e.MercuryTransmitterReaperMaxAge.String())
	add(envMercuryVerboseLogging, strconv.FormatBool(e.MercuryVerboseLogging))

	add(envPromPort, strconv.Itoa(e.PrometheusPort))

	add(envTracingEnabled, strconv.FormatBool(e.TracingEnabled))
	add(envTracingCollectorTarget, e.TracingCollectorTarget)
	add(envTracingSamplingRatio, strconv.FormatFloat(e.TracingSamplingRatio, 'f', -1, 64))
	add(envTracingTLSCertPath, e.TracingTLSCertPath)

	for k, v := range e.TracingAttributes {
		add(envTracingAttribute+k, v)
	}

	add(envTelemetryEnabled, strconv.FormatBool(e.TelemetryEnabled))
	add(envTelemetryEndpoint, e.TelemetryEndpoint)
	add(envTelemetryInsecureConn, strconv.FormatBool(e.TelemetryInsecureConnection))
	add(envTelemetryCACertFile, e.TelemetryCACertFile)
	add(envTelemetryTraceSampleRatio, strconv.FormatFloat(e.TelemetryTraceSampleRatio, 'f', -1, 64))
	for k, v := range e.TelemetryAttributes {
		add(envTelemetryAttribute+k, v)
	}

	for k, v := range e.TelemetryAuthHeaders {
		add(envTelemetryAuthHeader+k, v)
	}
	add(envTelemetryAuthPubKeyHex, e.TelemetryAuthPubKeyHex)
	add(envTelemetryAuthHeadersTTL, e.TelemetryAuthHeadersTTL.String())
	add(envTelemetryEmitterBatchProcessor, strconv.FormatBool(e.TelemetryEmitterBatchProcessor))
	add(envTelemetryEmitterExportTimeout, e.TelemetryEmitterExportTimeout.String())
	add(envTelemetryEmitterExportInterval, e.TelemetryEmitterExportInterval.String())
	add(envTelemetryEmitterExportMaxBatchSize, strconv.Itoa(e.TelemetryEmitterExportMaxBatchSize))
	add(envTelemetryEmitterMaxQueueSize, strconv.Itoa(e.TelemetryEmitterMaxQueueSize))
	add(envTelemetryLogStreamingEnabled, strconv.FormatBool(e.TelemetryLogStreamingEnabled))
	add(envTelemetryLogLevel, e.TelemetryLogLevel.String())
	add(envTelemetryLogBatchProcessor, strconv.FormatBool(e.TelemetryLogBatchProcessor))
	add(envTelemetryLogExportTimeout, e.TelemetryLogExportTimeout.String())
	add(envTelemetryLogExportMaxBatchSize, strconv.Itoa(e.TelemetryLogExportMaxBatchSize))
	add(envTelemetryLogExportInterval, e.TelemetryLogExportInterval.String())
	add(envTelemetryLogMaxQueueSize, strconv.Itoa(e.TelemetryLogMaxQueueSize))
	add(envTelemetryTraceCompressor, e.TelemetryTraceCompressor)
	add(envTelemetryMetricCompressor, e.TelemetryMetricCompressor)
	add(envTelemetryLogCompressor, e.TelemetryLogCompressor)

	add(envChipIngressEndpoint, e.ChipIngressEndpoint)
	add(envChipIngressInsecureConnection, strconv.FormatBool(e.ChipIngressInsecureConnection))

	if e.CRESettings != "" {
		add(envCRESettings, e.CRESettings)
	}
	if e.CRESettingsDefault != "" {
		add(envCRESettingsDefault, e.CRESettingsDefault)
	}

	return
}

// parse deserializes environment variables
func (e *EnvConfig) parse() error {
	e.AppID = os.Getenv(envAppID)
	var err error
	e.DatabaseURL, err = getEnv(envDatabaseURL, func(s string) (*config.SecretURL, error) {
		if s == "" { // DatabaseURL is optional
			return nil, nil
		}
		u, err2 := url.Parse(s)
		if err2 != nil {
			return nil, err2
		}
		return (*config.SecretURL)(u), nil
	})
	if err != nil {
		return err
	}
	if e.DatabaseURL != nil {
		e.DatabaseIdleInTxSessionTimeout, err = getDuration(envDatabaseIdleInTxSessionTimeout)
		if err != nil {
			return err
		}
		e.DatabaseLockTimeout, err = getDuration(envDatabaseLockTimeout)
		if err != nil {
			return err
		}
		e.DatabaseQueryTimeout, err = getDuration(envDatabaseQueryTimeout)
		if err != nil {
			return err
		}
		e.DatabaseListenerFallbackPollInterval, err = getDuration(envDatabaseListenerFallbackPollInterval)
		if err != nil {
			return err
		}
		e.DatabaseLogSQL, err = getBool(envDatabaseLogSQL)
		if err != nil {
			return err
		}
		e.DatabaseMaxOpenConns, err = getInt(envDatabaseMaxOpenConns)
		if err != nil {
			return err
		}
		e.DatabaseMaxIdleConns, err = getInt(envDatabaseMaxIdleConns)
		if err != nil {
			return err
		}
		e.DatabaseTracingEnabled, err = getBool(envDatabaseTracingEnabled)
		if err != nil {
			return err
		}
	}

	e.FeatureLogPoller, err = getBool(envFeatureLogPoller)
	if err != nil {
		return err
	}

	e.MercuryCacheLatestReportDeadline, err = getDuration(envMercuryCacheLatestReportDeadline)
	if err != nil {
		return err
	}
	e.MercuryCacheLatestReportTTL, err = getDuration(envMercuryCacheLatestReportTTL)
	if err != nil {
		return err
	}
	e.MercuryCacheMaxStaleAge, err = getDuration(envMercuryCacheMaxStaleAge)
	if err != nil {
		return err
	}

	e.MercuryTransmitterProtocol = os.Getenv(envMercuryTransmitterProtocol)
	e.MercuryTransmitterTransmitQueueMaxSize, err = getUint32(envMercuryTransmitterTransmitQueueMaxSize)
	if err != nil {
		return err
	}
	e.MercuryTransmitterTransmitTimeout, err = getDuration(envMercuryTransmitterTransmitTimeout)
	if err != nil {
		return err
	}
	e.MercuryTransmitterTransmitConcurrency, err = getUint32(envMercuryTransmitterTransmitConcurrency)
	if err != nil {
		return err
	}
	e.MercuryTransmitterReaperFrequency, err = getDuration(envMercuryTransmitterReaperFrequency)
	if err != nil {
		return err
	}
	e.MercuryTransmitterReaperMaxAge, err = getDuration(envMercuryTransmitterReaperMaxAge)
	if err != nil {
		return err
	}
	e.MercuryVerboseLogging, err = getBool(envMercuryVerboseLogging)
	if err != nil {
		return err
	}

	promPortStr := os.Getenv(envPromPort)
	e.PrometheusPort, err = strconv.Atoi(promPortStr)
	if err != nil {
		return fmt.Errorf("failed to parse %s = %q: %w", envPromPort, promPortStr, err)
	}

	e.TracingEnabled, err = getBool(envTracingEnabled)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envTracingEnabled, err)
	}

	if e.TracingEnabled {
		e.TracingCollectorTarget, err = getValidCollectorTarget()
		if err != nil {
			return err
		}
		e.TracingAttributes = getMap(envTracingAttribute)
		e.TracingSamplingRatio = getFloat64OrZero(envTracingSamplingRatio)
		e.TracingTLSCertPath = os.Getenv(envTracingTLSCertPath)
	}

	e.TelemetryEnabled, err = getBool(envTelemetryEnabled)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envTelemetryEnabled, err)
	}

	if e.TelemetryEnabled {
		e.TelemetryEndpoint = os.Getenv(envTelemetryEndpoint)
		e.TelemetryInsecureConnection, err = getBool(envTelemetryInsecureConn)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEndpoint, err)
		}
		e.TelemetryCACertFile = os.Getenv(envTelemetryCACertFile)
		e.TelemetryAttributes = getMap(envTelemetryAttribute)
		e.TelemetryTraceSampleRatio = getFloat64OrZero(envTelemetryTraceSampleRatio)
		e.TelemetryAuthHeaders = getMap(envTelemetryAuthHeader)
		e.TelemetryAuthPubKeyHex = os.Getenv(envTelemetryAuthPubKeyHex)
		e.TelemetryAuthHeadersTTL, err = getDuration(envTelemetryAuthHeadersTTL)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryAuthHeadersTTL, err)
		}
		e.TelemetryEmitterBatchProcessor, err = getBool(envTelemetryEmitterBatchProcessor)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterBatchProcessor, err)
		}
		e.TelemetryEmitterExportTimeout, err = getDuration(envTelemetryEmitterExportTimeout)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterExportTimeout, err)
		}
		e.TelemetryEmitterExportInterval, err = getDuration(envTelemetryEmitterExportInterval)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterExportInterval, err)
		}
		e.TelemetryEmitterExportMaxBatchSize, err = getInt(envTelemetryEmitterExportMaxBatchSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterExportMaxBatchSize, err)
		}
		e.TelemetryEmitterMaxQueueSize, err = getInt(envTelemetryEmitterMaxQueueSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterMaxQueueSize, err)
		}
		e.TelemetryLogStreamingEnabled, err = getBool(envTelemetryLogStreamingEnabled)
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
		e.TelemetryLogLevel = logLevel
		e.TelemetryLogBatchProcessor, err = getBool(envTelemetryLogBatchProcessor)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogBatchProcessor, err)
		}
		e.TelemetryLogExportTimeout, err = getDuration(envTelemetryLogExportTimeout)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogExportTimeout, err)
		}
		e.TelemetryLogExportMaxBatchSize, err = getInt(envTelemetryLogExportMaxBatchSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogExportMaxBatchSize, err)
		}
		e.TelemetryLogExportInterval, err = getDuration(envTelemetryLogExportInterval)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogExportInterval, err)
		}
		e.TelemetryLogMaxQueueSize, err = getInt(envTelemetryLogMaxQueueSize)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryLogMaxQueueSize, err)
		}
		e.TelemetryTraceCompressor = os.Getenv(envTelemetryTraceCompressor)
		e.TelemetryMetricCompressor = os.Getenv(envTelemetryMetricCompressor)
		e.TelemetryLogCompressor = os.Getenv(envTelemetryLogCompressor)
		// Optional
		e.ChipIngressEndpoint = os.Getenv(envChipIngressEndpoint)
		e.ChipIngressInsecureConnection, err = getBool(envChipIngressInsecureConnection)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envChipIngressInsecureConnection, err)
		}
	}

	e.CRESettings = os.Getenv(envCRESettings)
	e.CRESettingsDefault = os.Getenv(envCRESettingsDefault)

	return nil
}

// ManagedGRPCClientConfig return a Managed plugin and set grpc config values from the BrokerConfig.
// The innermost relevant BrokerConfig should be used, to include any relevant services in the logger name.
// Note: managed plugins shutdown when the parent process exits. We may want to change this behavior in the future
// to enable host process restarts without restarting the plugin. To do that we would also need
// supply the appropriate ReattachConfig to the plugin.ClientConfig.
func ManagedGRPCClientConfig(clientConfig *plugin.ClientConfig, c BrokerConfig) *plugin.ClientConfig {
	clientConfig.AllowedProtocols = []plugin.Protocol{plugin.ProtocolGRPC}
	clientConfig.GRPCDialOptions = c.DialOpts
	clientConfig.Logger = HCLogLogger(c.Logger)
	clientConfig.Managed = true
	return clientConfig
}

func getBool(envKey string) (bool, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return false, nil
	}
	return strconv.ParseBool(s)
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

func getUint32(envKey string) (uint32, error) {
	s := os.Getenv(envKey)
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(u), nil
}

func getDuration(envKey string) (time.Duration, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}

func getEnv[T any](key string, parse func(string) (T, error)) (t T, err error) {
	v := os.Getenv(key)
	t, err = parse(v)
	if err != nil {
		err = fmt.Errorf("failed to parse %s=%s: %w", key, v, err)
	}
	return
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
