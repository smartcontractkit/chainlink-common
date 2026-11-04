package loop

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/smartcontractkit/chainlink-common/pkg/durableemitter"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestServer_MeteringConfig(t *testing.T) {
	t.Run("no durable emitter yields a nil emitter", func(t *testing.T) {
		s := &Server{EnvConfig: EnvConfig{MeterRecordsEnabled: true}}
		cfg := s.MeteringConfig()
		assert.True(t, cfg.MeterRecordsEnabled)
		assert.Nil(t, cfg.Emitter)
	})

	t.Run("injects the server's own durable emitter, not the process global", func(t *testing.T) {
		global := &durableemitter.DurableEmitter{}
		durableemitter.SetGlobalEmitter(global)
		t.Cleanup(func() { durableemitter.SetGlobalEmitter(nil) })

		own := &durableemitter.DurableEmitter{}
		s := &Server{
			EnvConfig:      EnvConfig{MeterRecordsEnabled: true},
			durableEmitter: own,
		}

		cfg := s.MeteringConfig()
		require.NotNil(t, cfg.Emitter)
		assert.Same(t, own, cfg.Emitter)
		assert.NotSame(t, global, cfg.Emitter)
	})
}

func TestOtelErrorHandler(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "plain error", err: errors.New("boom")},
		{
			name: "wrapped export failure",
			err:  fmt.Errorf("failed to upload metrics: %w", errors.New("rpc error: code = Unavailable desc = data refused due to high memory usage")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lggr, logs := logger.TestObserved(t, zapcore.DebugLevel)

			otelErrorHandler(lggr).Handle(tt.err)

			entries := logs.TakeAll()
			require.Len(t, entries, 1)
			assert.Equal(t, zapcore.ErrorLevel, entries[0].Level)
			assert.Equal(t, "Telemetry error", entries[0].Message)
			assert.Equal(t, tt.err.Error(), entries[0].ContextMap()["err"])
		})
	}
}

// Asserts that once telemetry starts, OTel SDK errors reach the server logger instead of the default stderr printer.
func TestServer_startTelemetryRegistersOtelErrorHandler(t *testing.T) {
	prev := otel.GetErrorHandler()
	t.Cleanup(func() { otel.SetErrorHandler(prev) })

	// An observer-only logger, since the global handler may outlive this test and zaptest panics on late writes.
	core, logs := observer.New(zapcore.DebugLevel)
	s := &Server{Logger: logger.Sugared(logger.NewWithCores(core))}

	otel.Handle(errors.New("before start"))
	require.Zero(t, logs.Len())

	require.NoError(t, s.startTelemetry(t.Context()))
	t.Cleanup(func() { assert.NoError(t, s.telemetry.Close()) })

	otel.Handle(errors.New("failed to upload metrics: exporter export timeout"))

	entries := logs.FilterMessage("Telemetry error").All()
	require.Len(t, entries, 1)
	assert.Equal(t, zapcore.ErrorLevel, entries[0].Level)
	assert.Equal(t, "failed to upload metrics: exporter export timeout", entries[0].ContextMap()["err"])
}
