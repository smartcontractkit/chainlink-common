package telemetry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestNewTelemetry_WithoutAnEndpointChangesNothing(t *testing.T) {
	before := beholder.GetClient()

	telemetry, err := NewTelemetry(logger.Test(t), TelemetryConfig{}, "app-id", nil)
	require.NoError(t, err)
	assert.Nil(t, telemetry.Client())
	assert.Same(t, before, beholder.GetClient())

	require.NoError(t, telemetry.Start(t.Context()))
	require.NoError(t, telemetry.Close())
	assert.Same(t, before, beholder.GetClient())
}

func TestNewTelemetry_DisabledIgnoresTheEndpoint(t *testing.T) {
	before := beholder.GetClient()

	telemetry, err := NewTelemetry(logger.Test(t), TelemetryConfig{TelemetryEndpoint: "otel:4317"}, "app-id", nil)
	require.NoError(t, err)
	assert.Nil(t, telemetry.Client())
	assert.Same(t, before, beholder.GetClient())
}

func TestTelemetry_InstallsTheClientWhenStarted(t *testing.T) {
	original := beholder.GetClient()
	t.Cleanup(func() {
		beholder.SetClient(original)
		beholder.SetGlobalOtelProviders()
	})
	client := beholder.NewNoopClient()

	telemetry := newTelemetry(logger.Test(t), client)
	require.Same(t, original, beholder.GetClient(), "not installed until started")

	require.NoError(t, telemetry.Start(t.Context()))
	assert.Same(t, client, beholder.GetClient())

	require.NoError(t, telemetry.Close())
	assert.Same(t, client, beholder.GetClient(), "left installed when closed")
}

func TestTelemetryConfig_tracingConfig(t *testing.T) {
	lggr, logs := detachedLogger()
	cfg := TelemetryConfig{TracingEnabled: true, TracingCollectorTarget: "otel:4317", TracingAttributes: map[string]string{"foo": "bar"}}
	tracing := cfg.tracingConfig(lggr, "app-id")
	assert.Equal(t, beholder.OtelAttributes{"foo": "bar", string(semconv.ServiceInstanceIDKey): "app-id"}, tracing.NodeAttributes)
	assert.Equal(t, map[string]string{"foo": "bar"}, cfg.TracingAttributes, "env config is not mutated")

	// NewSpanExporter's dialer calls this unchecked; building a real exporter here would leak its
	// connection, which Shutdown does not close.
	require.NotNil(t, tracing.OnDialError)
	tracing.OnDialError(errors.New("refused"))
	assert.Len(t, logs.FilterMessage("Failed to dial the tracing collector").All(), 1)
}

func TestTelemetryConfig_beholderConfig(t *testing.T) {
	t.Run("prometheus bridge", func(t *testing.T) {
		off, err := TelemetryConfig{}.beholderConfig(logger.Test(t), TracingConfig{}, nil)
		require.NoError(t, err)
		assert.Empty(t, off.MetricProducers)

		on, err := TelemetryConfig{TelemetryPrometheusBridgeEnabled: true, TelemetryPrometheusBridgePrefixes: []string{"foo_"}}.beholderConfig(logger.Test(t), TracingConfig{}, nil)
		require.NoError(t, err)
		assert.Len(t, on.MetricProducers, 1)
	})

	t.Run("cardinality limit defaults when unset", func(t *testing.T) {
		unset, err := TelemetryConfig{}.beholderConfig(logger.Test(t), TracingConfig{}, nil)
		require.NoError(t, err)
		assert.Equal(t, beholder.DefaultMetricCardinalityLimit, unset.MetricCardinalityLimit)

		disabled, err := TelemetryConfig{TelemetryMetricCardinalityLimit: new(0)}.beholderConfig(logger.Test(t), TracingConfig{}, nil)
		require.NoError(t, err)
		assert.Equal(t, 0, disabled.MetricCardinalityLimit)
	})
}

func detachedLogger() (logger.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.ErrorLevel)
	return logger.NewWithCores(core), logs
}

func TestTelemetryConfig_authHeaderEnv(t *testing.T) {
	t.Run("written in both forms", func(t *testing.T) {
		env := (&TelemetryConfig{TelemetryAuthHeaders: map[string]string{"authorization": "token"}}).AsCmdEnv()
		assert.Contains(t, env, "CL_TELEMETRY_AUTH_HEADER_authorization=token")
		assert.Contains(t, env, "CL_TELEMETRY_AUTH_HEADERauthorization=token")
	})

	t.Run("separated form wins", func(t *testing.T) {
		t.Setenv(envTelemetryEnabled, "true")
		t.Setenv("CL_TELEMETRY_AUTH_HEADER_authorization", "new")
		t.Setenv("CL_TELEMETRY_AUTH_HEADERauthorization", "legacy")

		var cfg TelemetryConfig
		require.NoError(t, cfg.Parse())
		assert.Equal(t, map[string]string{"authorization": "new"}, cfg.TelemetryAuthHeaders)
	})

	t.Run("legacy form is read when it is the only one", func(t *testing.T) {
		t.Setenv(envTelemetryEnabled, "true")
		t.Setenv("CL_TELEMETRY_AUTH_HEADERauthorization", "legacy")

		var cfg TelemetryConfig
		require.NoError(t, cfg.Parse())
		assert.Equal(t, map[string]string{"authorization": "legacy"}, cfg.TelemetryAuthHeaders)
	})
}

func TestGetMap(t *testing.T) {
	t.Setenv("TEST_PREFIX_KEY1", "value1")
	t.Setenv("TEST_PREFIX_KEY2", "value2")
	t.Setenv("OTHER_KEY", "othervalue")

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
