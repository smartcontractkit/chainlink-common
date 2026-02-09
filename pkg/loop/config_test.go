package loop

import (
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

		expectConfig EnvConfig
	}{
		{
			name: "All variables set correctly",
			envVars: map[string]string{
				envAppID:                                "app-id",
				envDatabaseURL:                          "postgres://user:password@localhost:5432/db",
				envDatabaseIdleInTxSessionTimeout:       "42s",
				envDatabaseLockTimeout:                  "8m",
				envDatabaseQueryTimeout:                 "7s",
				envDatabaseListenerFallbackPollInterval: "17s",
				envDatabaseLogSQL:                       "true",
				envDatabaseMaxOpenConns:                 "9999",
				envDatabaseMaxIdleConns:                 "8080",
				envDatabaseTracingEnabled:               "true",

				envFeatureLogPoller: "true",

				envGRPCServerMaxRecvMsgSize: "42",

				envMercuryCacheLatestReportDeadline: "1ms",
				envMercuryCacheLatestReportTTL:      "1µs",
				envMercuryCacheMaxStaleAge:          "1ns",

				envMercuryTransmitterProtocol:             "foo",
				envMercuryTransmitterTransmitQueueMaxSize: "42",
				envMercuryTransmitterTransmitTimeout:      "1s",
				envMercuryTransmitterTransmitConcurrency:  "13",
				envMercuryTransmitterReaperFrequency:      "1h",
				envMercuryTransmitterReaperMaxAge:         "1m",
				envMercuryVerboseLogging:                  "true",

				envPromPort: "8080",

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
				envTelemetryLogStreamingEnabled:       "false",

				envChipIngressEndpoint:           "chip-ingress.example.com:50051",
				envChipIngressInsecureConnection: "true",

				envCRESettings:        `{"global":{}}`,
				envCRESettingsDefault: `{"foo":"bar"}`,
			},
			expectError:  false,
			expectConfig: envCfgFull,
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
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectConfig, config)
			}
		})
	}
}

var envCfgFull = EnvConfig{
	AppID: "app-id",

	DatabaseURL:                          config.MustSecretURL("postgres://user:password@localhost:5432/db"),
	DatabaseIdleInTxSessionTimeout:       42 * time.Second,
	DatabaseLockTimeout:                  8 * time.Minute,
	DatabaseQueryTimeout:                 7 * time.Second,
	DatabaseListenerFallbackPollInterval: 17 * time.Second,
	DatabaseLogSQL:                       true,
	DatabaseMaxOpenConns:                 9999,
	DatabaseMaxIdleConns:                 8080,
	DatabaseTracingEnabled:               true,

	FeatureLogPoller: true,

	GRPCServerMaxRecvMsgSize: 42,

	MercuryCacheLatestReportDeadline: time.Millisecond,
	MercuryCacheLatestReportTTL:      time.Microsecond,
	MercuryCacheMaxStaleAge:          time.Nanosecond,

	MercuryTransmitterProtocol:             "foo",
	MercuryTransmitterTransmitQueueMaxSize: 42,
	MercuryTransmitterTransmitTimeout:      time.Second,
	MercuryTransmitterTransmitConcurrency:  13,
	MercuryTransmitterReaperFrequency:      time.Hour,
	MercuryTransmitterReaperMaxAge:         time.Minute,
	MercuryVerboseLogging:                  true,

	PrometheusPort: 8080,

	TracingEnabled:         true,
	TracingAttributes:      map[string]string{"XYZ": "value"},
	TracingCollectorTarget: "some:target",
	TracingSamplingRatio:   1.0,
	TracingTLSCertPath:     "internal/test/fixtures/client.pem",

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
	TelemetryLogStreamingEnabled:       false,

	ChipIngressEndpoint:           "chip-ingress.example.com:50051",
	ChipIngressInsecureConnection: true,

	CRESettings:        `{"global":{}}`,
	CRESettingsDefault: `{"foo":"bar"}`,
}

func TestEnvConfig_AsCmdEnv(t *testing.T) {
	got := map[string]string{}
	for _, kv := range envCfgFull.AsCmdEnv() {
		pair := strings.SplitN(kv, "=", 2)
		require.Len(t, pair, 2)
		got[pair[0]] = pair[1]
	}

	assert.Equal(t, "postgres://user:password@localhost:5432/db", got[envDatabaseURL])
	assert.Equal(t, "true", got["CL_DATABASE_TRACING_ENABLED"])

	assert.Equal(t, "42", got[envGRPCServerMaxRecvMsgSize])
	assert.Equal(t, "1ms", got[envMercuryCacheLatestReportDeadline])
	assert.Equal(t, "1µs", got[envMercuryCacheLatestReportTTL])
	assert.Equal(t, "1ns", got[envMercuryCacheMaxStaleAge])
	assert.Equal(t, "foo", got[envMercuryTransmitterProtocol])
	assert.Equal(t, "42", got[envMercuryTransmitterTransmitQueueMaxSize])
	assert.Equal(t, "1s", got[envMercuryTransmitterTransmitTimeout])
	assert.Equal(t, "13", got[envMercuryTransmitterTransmitConcurrency])
	assert.Equal(t, "1h0m0s", got[envMercuryTransmitterReaperFrequency])
	assert.Equal(t, "1m0s", got[envMercuryTransmitterReaperMaxAge])
	assert.Equal(t, "true", got[envMercuryVerboseLogging])

	assert.Equal(t, strconv.Itoa(8080), got[envPromPort])

	assert.Equal(t, "true", got[envTracingEnabled])
	assert.Equal(t, "some:target", got[envTracingCollectorTarget])
	assert.Equal(t, "1", got[envTracingSamplingRatio])
	assert.Equal(t, "internal/test/fixtures/client.pem", got[envTracingTLSCertPath])
	assert.Equal(t, "value", got[envTracingAttribute+"XYZ"])

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
	assert.Equal(t, "false", got[envTelemetryLogStreamingEnabled])

	// Assert ChipIngress environment variables
	assert.Equal(t, "chip-ingress.example.com:50051", got[envChipIngressEndpoint])
	assert.Equal(t, "true", got[envChipIngressInsecureConnection])

	assert.Equal(t, `{"global":{}}`, got[envCRESettings])
	assert.Equal(t, `{"foo":"bar"}`, got[envCRESettingsDefault])
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
