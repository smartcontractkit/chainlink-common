package loop

import (
	"maps"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestEnvConfig_parse(t *testing.T) {
	cases := []struct {
		name        string
		envVars     map[string]string
		expectError bool

		expectedDatabaseURL                    string
		expectedDatabaseIdleInTxSessionTimeout time.Duration
		expectedDatabaseLockTimeout            time.Duration
		expectedDatabaseQueryTimeout           time.Duration
		expectedDatabaseLogSQL                 bool
		expectedDatabaseMaxOpenConns           int
		expectedDatabaseMaxIdleConns           int

		expectedPrometheusPort         int
		expectedTracingEnabled         bool
		expectedTracingCollectorTarget string
		expectedTracingSamplingRatio   float64
		expectedTracingTLSCertPath     string

		expectedTelemetryEnabled                   bool
		expectedTelemetryEndpoint                  string
		expectedTelemetryInsecureConn              bool
		expectedTelemetryCACertFile                string
		expectedTelemetryAttributes                OtelAttributes
		expectedTelemetryTraceSampleRatio          float64
		expectedTelemetryAuthHeaders               map[string]string
		expectedTelemetryAuthPubKeyHex             string
		expectedTelemetryEmitterBatchProcessor     bool
		expectedTelemetryEmitterExportTimeout      time.Duration
		expectedTelemetryEmitterExportInterval     time.Duration
		expectedTelemetryEmitterExportMaxBatchSize int
		expectedTelemetryEmitterMaxQueueSize       int
	}{
		{
			name: "All variables set correctly",
			envVars: map[string]string{
				envDatabaseURL:                    "postgres://user:password@localhost:5432/db",
				envDatabaseIdleInTxSessionTimeout: "42s",
				envDatabaseLockTimeout:            "8m",
				envDatabaseQueryTimeout:           "7s",
				envDatabaseLogSQL:                 "true",
				envDatabaseMaxOpenConns:           "9999",
				envDatabaseMaxIdleConns:           "8080",

				envPromPort:                 "8080",
				envTracingEnabled:           "true",
				envTracingCollectorTarget:   "some:target",
				envTracingSamplingRatio:     "1.0",
				envTracingTLSCertPath:       "internal/test/fixtures/client.pem",
				envTracingAttribute + "XYZ": "value",

				envTelemetryEnabled:                   "true",
				envTelemetryEndpoint:                  "example.com/beholder",
				envTelemetryInsecureConn:              "true",
				envTelemetryCACertFile:                "foo/bar",
				envTelemetryAttribute + "foo":         "bar",
				envTelemetryAttribute + "baz":         "42",
				envTelemetryTraceSampleRatio:          "0.42",
				envTelemetryAuthHeader + "header-key": "header-value",
				envTelemetryAuthPubKeyHex:             "pub-key-hex",
				envTelemetryEmitterBatchProcessor:     "true",
				envTelemetryEmitterExportTimeout:      "1s",
				envTelemetryEmitterExportInterval:     "2s",
				envTelemetryEmitterExportMaxBatchSize: "100",
				envTelemetryEmitterMaxQueueSize:       "1000",
			},
			expectError: false,

			expectedDatabaseURL:                    "postgres://user:password@localhost:5432/db",
			expectedDatabaseIdleInTxSessionTimeout: 42 * time.Second,
			expectedDatabaseLockTimeout:            8 * time.Minute,
			expectedDatabaseQueryTimeout:           7 * time.Second,
			expectedDatabaseLogSQL:                 true,
			expectedDatabaseMaxOpenConns:           9999,
			expectedDatabaseMaxIdleConns:           8080,

			expectedPrometheusPort:         8080,
			expectedTracingEnabled:         true,
			expectedTracingCollectorTarget: "some:target",
			expectedTracingSamplingRatio:   1.0,
			expectedTracingTLSCertPath:     "internal/test/fixtures/client.pem",

			expectedTelemetryEnabled:                   true,
			expectedTelemetryEndpoint:                  "example.com/beholder",
			expectedTelemetryInsecureConn:              true,
			expectedTelemetryCACertFile:                "foo/bar",
			expectedTelemetryAttributes:                OtelAttributes{"foo": "bar", "baz": "42"},
			expectedTelemetryTraceSampleRatio:          0.42,
			expectedTelemetryAuthHeaders:               map[string]string{"header-key": "header-value"},
			expectedTelemetryAuthPubKeyHex:             "pub-key-hex",
			expectedTelemetryEmitterBatchProcessor:     true,
			expectedTelemetryEmitterExportTimeout:      1 * time.Second,
			expectedTelemetryEmitterExportInterval:     2 * time.Second,
			expectedTelemetryEmitterExportMaxBatchSize: 100,
			expectedTelemetryEmitterMaxQueueSize:       1000,
		},
		{
			name: "CL_DATABASE_URL parse error",
			envVars: map[string]string{
				envDatabaseURL: "wrong-db-url",
			},
			expectError: true,
		},
		{
			name: "CL_PROMETHEUS_PORT parse error",
			envVars: map[string]string{
				envPromPort: "abc",
			},
			expectError: true,
		},
		{
			name: "TRACING_ENABLED parse error",
			envVars: map[string]string{
				envPromPort:       "8080",
				envTracingEnabled: "invalid_bool",
			},
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			var config EnvConfig
			err := config.parse()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					if config.DatabaseURL.URL().String() != tc.expectedDatabaseURL {
						t.Errorf("Expected Database URL %s, got %s", tc.expectedDatabaseURL, config.DatabaseURL.String())
					}
					if config.DatabaseIdleInTxSessionTimeout != tc.expectedDatabaseIdleInTxSessionTimeout {
						t.Errorf("Expected Database idle in tx session timeout %s, got %s", tc.expectedDatabaseIdleInTxSessionTimeout, config.DatabaseIdleInTxSessionTimeout)
					}
					if config.DatabaseLockTimeout != tc.expectedDatabaseLockTimeout {
						t.Errorf("Expected Database lock timeout %s, got %s", tc.expectedDatabaseLockTimeout, config.DatabaseLockTimeout)
					}
					if config.DatabaseQueryTimeout != tc.expectedDatabaseQueryTimeout {
						t.Errorf("Expected Database query timeout %s, got %s", tc.expectedDatabaseQueryTimeout, config.DatabaseQueryTimeout)
					}
					if config.DatabaseLogSQL != tc.expectedDatabaseLogSQL {
						t.Errorf("Expected Database log sql %t, got %t", tc.expectedDatabaseLogSQL, config.DatabaseLogSQL)
					}
					if config.DatabaseMaxOpenConns != tc.expectedDatabaseMaxOpenConns {
						t.Errorf("Expected Database max open conns %d, got %d", tc.expectedDatabaseMaxOpenConns, config.DatabaseMaxOpenConns)
					}
					if config.DatabaseMaxIdleConns != tc.expectedDatabaseMaxIdleConns {
						t.Errorf("Expected Database max idle conns %d, got %d", tc.expectedDatabaseMaxIdleConns, config.DatabaseMaxIdleConns)
					}

					if config.PrometheusPort != tc.expectedPrometheusPort {
						t.Errorf("Expected Prometheus port %d, got %d", tc.expectedPrometheusPort, config.PrometheusPort)
					}
					if config.TracingEnabled != tc.expectedTracingEnabled {
						t.Errorf("Expected tracingEnabled %v, got %v", tc.expectedTracingEnabled, config.TracingEnabled)
					}
					if config.TracingCollectorTarget != tc.expectedTracingCollectorTarget {
						t.Errorf("Expected tracingCollectorTarget %s, got %s", tc.expectedTracingCollectorTarget, config.TracingCollectorTarget)
					}
					if config.TracingSamplingRatio != tc.expectedTracingSamplingRatio {
						t.Errorf("Expected tracingSamplingRatio %f, got %f", tc.expectedTracingSamplingRatio, config.TracingSamplingRatio)
					}
					if config.TracingTLSCertPath != tc.expectedTracingTLSCertPath {
						t.Errorf("Expected tracingTLSCertPath %s, got %s", tc.expectedTracingTLSCertPath, config.TracingTLSCertPath)
					}
					if config.TelemetryEnabled != tc.expectedTelemetryEnabled {
						t.Errorf("Expected telemetryEnabled %v, got %v", tc.expectedTelemetryEnabled, config.TelemetryEnabled)
					}
					if config.TelemetryEndpoint != tc.expectedTelemetryEndpoint {
						t.Errorf("Expected telemetryEndpoint %s, got %s", tc.expectedTelemetryEndpoint, config.TelemetryEndpoint)
					}
					if config.TelemetryInsecureConnection != tc.expectedTelemetryInsecureConn {
						t.Errorf("Expected telemetryInsecureConn %v, got %v", tc.expectedTelemetryInsecureConn, config.TelemetryInsecureConnection)
					}
					if config.TelemetryCACertFile != tc.expectedTelemetryCACertFile {
						t.Errorf("Expected telemetryCACertFile %s, got %s", tc.expectedTelemetryCACertFile, config.TelemetryCACertFile)
					}
					if !maps.Equal(config.TelemetryAttributes, tc.expectedTelemetryAttributes) {
						t.Errorf("Expected telemetryAttributes %v, got %v", tc.expectedTelemetryAttributes, config.TelemetryAttributes)
					}
					if config.TelemetryTraceSampleRatio != tc.expectedTelemetryTraceSampleRatio {
						t.Errorf("Expected telemetryTraceSampleRatio %f, got %f", tc.expectedTelemetryTraceSampleRatio, config.TelemetryTraceSampleRatio)
					}
					if !maps.Equal(config.TelemetryAuthHeaders, tc.expectedTelemetryAuthHeaders) {
						t.Errorf("Expected telemetryAuthHeaders %v, got %v", tc.expectedTelemetryAuthHeaders, config.TelemetryAuthHeaders)
					}
					if config.TelemetryAuthPubKeyHex != tc.expectedTelemetryAuthPubKeyHex {
						t.Errorf("Expected telemetryAuthPubKeyHex %s, got %s", tc.expectedTelemetryAuthPubKeyHex, config.TelemetryAuthPubKeyHex)
					}
					if config.TelemetryEmitterBatchProcessor != tc.expectedTelemetryEmitterBatchProcessor {
						t.Errorf("Expected telemetryEmitterBatchProcessor %v, got %v", tc.expectedTelemetryEmitterBatchProcessor, config.TelemetryEmitterBatchProcessor)
					}
					if config.TelemetryEmitterExportTimeout != tc.expectedTelemetryEmitterExportTimeout {
						t.Errorf("Expected telemetryEmitterExportTimeout %v, got %v", tc.expectedTelemetryEmitterExportTimeout, config.TelemetryEmitterExportTimeout)
					}
					if config.TelemetryEmitterExportInterval != tc.expectedTelemetryEmitterExportInterval {
						t.Errorf("Expected telemetryEmitterExportInterval %v, got %v", tc.expectedTelemetryEmitterExportInterval, config.TelemetryEmitterExportInterval)
					}
					if config.TelemetryEmitterExportMaxBatchSize != tc.expectedTelemetryEmitterExportMaxBatchSize {
						t.Errorf("Expected telemetryEmitterExportMaxBatchSize %d, got %d", tc.expectedTelemetryEmitterExportMaxBatchSize, config.TelemetryEmitterExportMaxBatchSize)
					}
					if config.TelemetryEmitterMaxQueueSize != tc.expectedTelemetryEmitterMaxQueueSize {
						t.Errorf("Expected telemetryEmitterMaxQueueSize %d, got %d", tc.expectedTelemetryEmitterMaxQueueSize, config.TelemetryEmitterMaxQueueSize)
					}
				}
			}
		})
	}
}

func equalOtelAttributes(a, b OtelAttributes) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func equalStringMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func TestEnvConfig_AsCmdEnv(t *testing.T) {
	envCfg := EnvConfig{
		DatabaseURL:    (*config.SecretURL)(&url.URL{Scheme: "postgres", Host: "localhost:5432", User: url.UserPassword("user", "password"), Path: "/db"}),
		PrometheusPort: 9090,

		TracingEnabled:         true,
		TracingCollectorTarget: "http://localhost:9000",
		TracingSamplingRatio:   0.1,
		TracingTLSCertPath:     "some/path",
		TracingAttributes:      map[string]string{"key": "value"},

		TelemetryEnabled:                   true,
		TelemetryEndpoint:                  "example.com/beholder",
		TelemetryInsecureConnection:        true,
		TelemetryCACertFile:                "foo/bar",
		TelemetryAttributes:                OtelAttributes{"foo": "bar", "baz": "42"},
		TelemetryTraceSampleRatio:          0.42,
		TelemetryAuthHeaders:               map[string]string{"header-key": "header-value"},
		TelemetryAuthPubKeyHex:             "pub-key-hex",
		TelemetryEmitterBatchProcessor:     true,
		TelemetryEmitterExportTimeout:      1 * time.Second,
		TelemetryEmitterExportInterval:     2 * time.Second,
		TelemetryEmitterExportMaxBatchSize: 100,
		TelemetryEmitterMaxQueueSize:       1000,
	}
	got := map[string]string{}
	for _, kv := range envCfg.AsCmdEnv() {
		pair := strings.SplitN(kv, "=", 2)
		require.Len(t, pair, 2)
		got[pair[0]] = pair[1]
	}

	assert.Equal(t, "postgres://user:password@localhost:5432/db", got[envDatabaseURL])
	assert.Equal(t, strconv.Itoa(9090), got[envPromPort])

	assert.Equal(t, "true", got[envTracingEnabled])
	assert.Equal(t, "http://localhost:9000", got[envTracingCollectorTarget])
	assert.Equal(t, "0.1", got[envTracingSamplingRatio])
	assert.Equal(t, "some/path", got[envTracingTLSCertPath])
	assert.Equal(t, "value", got[envTracingAttribute+"key"])

	assert.Equal(t, "true", got[envTelemetryEnabled])
	assert.Equal(t, "example.com/beholder", got[envTelemetryEndpoint])
	assert.Equal(t, "true", got[envTelemetryInsecureConn])
	assert.Equal(t, "foo/bar", got[envTelemetryCACertFile])
	assert.Equal(t, "0.42", got[envTelemetryTraceSampleRatio])
	assert.Equal(t, "bar", got[envTelemetryAttribute+"foo"])
	assert.Equal(t, "42", got[envTelemetryAttribute+"baz"])
	assert.Equal(t, "header-value", got[envTelemetryAuthHeader+"header-key"])
	assert.Equal(t, "pub-key-hex", got[envTelemetryAuthPubKeyHex])
	assert.Equal(t, "true", got[envTelemetryEmitterBatchProcessor])
	assert.Equal(t, "1s", got[envTelemetryEmitterExportTimeout])
	assert.Equal(t, "2s", got[envTelemetryEmitterExportInterval])
	assert.Equal(t, "100", got[envTelemetryEmitterExportMaxBatchSize])
	assert.Equal(t, "1000", got[envTelemetryEmitterMaxQueueSize])
}

func TestGetMap(t *testing.T) {
	os.Setenv("TEST_PREFIX_KEY1", "value1")
	os.Setenv("TEST_PREFIX_KEY2", "value2")
	os.Setenv("OTHER_KEY", "othervalue")

	defer func() {
		os.Unsetenv("TEST_PREFIX_KEY1")
		os.Unsetenv("TEST_PREFIX_KEY2")
		os.Unsetenv("OTHER_KEY")
	}()

	result := getMap("TEST_PREFIX_")

	expected := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
	}

	if len(result) != len(expected) {
		t.Errorf("Expected map length %d, got %d", len(expected), len(result))
	}

	for k, v := range expected {
		if result[k] != v {
			t.Errorf("Expected key %s to have value %s, but got %s", k, v, result[k])
		}
	}
}

func TestManagedGRPCClientConfig(t *testing.T) {
	t.Parallel()

	t.Run("returns a new client config with the provided broker config", func(t *testing.T) {
		t.Parallel()

		brokerConfig := BrokerConfig{
			Logger: logger.Test(t),
			GRPCOpts: GRPCOpts{
				DialOpts: []grpc.DialOption{
					grpc.WithNoProxy(), // any grpc.DialOption will do
				},
			},
		}

		clientConfig := ManagedGRPCClientConfig(&plugin.ClientConfig{}, brokerConfig)

		assert.NotNil(t, clientConfig.Logger)
		assert.Equal(t, []plugin.Protocol{plugin.ProtocolGRPC}, clientConfig.AllowedProtocols)
		assert.Equal(t, brokerConfig.GRPCOpts.DialOpts, clientConfig.GRPCDialOptions)
		assert.True(t, clientConfig.Managed)
	})
}
