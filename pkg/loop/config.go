package loop

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
)

const (
	envDatabaseURL = "CL_DATABASE_URL"
	envPromPort    = "CL_PROMETHEUS_PORT"

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
	envTelemetryEmitterBatchProcessor     = "CL_TELEMETRY_EMITTER_BATCH_PROCESSOR"
	envTelemetryEmitterExportTimeout      = "CL_TELEMETRY_EMITTER_EXPORT_TIMEOUT"
	envTelemetryEmitterExportInterval     = "CL_TELEMETRY_EMITTER_EXPORT_INTERVAL"
	envTelemetryEmitterExportMaxBatchSize = "CL_TELEMETRY_EMITTER_EXPORT_MAX_BATCH_SIZE"
	envTelemetryEmitterMaxQueueSize       = "CL_TELEMETRY_EMITTER_MAX_QUEUE_SIZE"
)

// EnvConfig is the configuration between the application and the LOOP executable. The values
// are fully resolved and static and passed via the environment.
type EnvConfig struct {
	DatabaseURL *url.URL

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
	TelemetryEmitterBatchProcessor     bool
	TelemetryEmitterExportTimeout      time.Duration
	TelemetryEmitterExportInterval     time.Duration
	TelemetryEmitterExportMaxBatchSize int
	TelemetryEmitterMaxQueueSize       int
}

// AsCmdEnv returns a slice of environment variable key/value pairs for an exec.Cmd.
func (e *EnvConfig) AsCmdEnv() (env []string) {
	add := func(k, v string) {
		env = append(env, k+"="+v)
	}

	if e.DatabaseURL != nil { // optional
		add(envDatabaseURL, e.DatabaseURL.String())
	}
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
	add(envTelemetryEmitterBatchProcessor, strconv.FormatBool(e.TelemetryEmitterBatchProcessor))
	add(envTelemetryEmitterExportTimeout, e.TelemetryEmitterExportTimeout.String())
	add(envTelemetryEmitterExportInterval, e.TelemetryEmitterExportInterval.String())
	add(envTelemetryEmitterExportMaxBatchSize, strconv.Itoa(e.TelemetryEmitterExportMaxBatchSize))
	add(envTelemetryEmitterMaxQueueSize, strconv.Itoa(e.TelemetryEmitterMaxQueueSize))
	return
}

// parse deserializes environment variables
func (e *EnvConfig) parse() error {
	promPortStr := os.Getenv(envPromPort)
	var err error
	e.DatabaseURL, err = getDatabaseURL()
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envDatabaseURL, err)
	}

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
		e.TelemetryEmitterBatchProcessor, err = getBool(envTelemetryEmitterBatchProcessor)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterBatchProcessor, err)
		}
		e.TelemetryEmitterExportTimeout, err = time.ParseDuration(os.Getenv(envTelemetryEmitterExportTimeout))
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterExportTimeout, err)
		}
		e.TelemetryEmitterExportInterval, err = time.ParseDuration(os.Getenv(envTelemetryEmitterExportInterval))
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterExportInterval, err)
		}
		e.TelemetryEmitterExportMaxBatchSize, err = strconv.Atoi(os.Getenv(envTelemetryEmitterExportMaxBatchSize))
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterExportMaxBatchSize, err)
		}
		e.TelemetryEmitterMaxQueueSize, err = strconv.Atoi(os.Getenv(envTelemetryEmitterMaxQueueSize))
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envTelemetryEmitterMaxQueueSize, err)
		}
	}
	return nil
}

// ManagedGRPCClientConfig return a Managed plugin and set grpc config values from the BrokerConfig.
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

// getDatabaseURL parses the CL_DATABASE_URL environment variable.
func getDatabaseURL() (*url.URL, error) {
	databaseURL := os.Getenv(envDatabaseURL)
	if databaseURL == "" {
		// DatabaseURL is optional
		return nil, nil
	}
	u, err := url.Parse(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", envDatabaseURL, err)
	}
	return u, nil
}
