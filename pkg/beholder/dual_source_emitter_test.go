package beholder_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestNewDualSourceEmitter(t *testing.T) {
	// Test successful creation
	t.Run("successful creation", func(t *testing.T) {
		chipEmitter := &mockEmitter{}
		otelEmitter := &mockEmitter{}

		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter, logger.Test(t))
		require.NoError(t, err)

		assert.NotNil(t, emitter)
		assert.IsType(t, &beholder.DualSourceEmitter{}, emitter)
	})

	// Test nil chip ingress emitter
	t.Run("nil chip ingress emitter", func(t *testing.T) {
		otelEmitter := &mockEmitter{}
		emitter, err := beholder.NewDualSourceEmitter(nil, otelEmitter, logger.Test(t))

		assert.Error(t, err)
		assert.Nil(t, emitter)
	})

	// Test nil otel collector emitter
	t.Run("nil otel collector emitter", func(t *testing.T) {
		chipEmitter := &mockEmitter{}
		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, nil, logger.Test(t))

		assert.Error(t, err)
		assert.Nil(t, emitter)
	})
}
func TestDualSourceEmitterEmit(t *testing.T) {
	t.Run("successful emit to both destinations", func(t *testing.T) {
		chipEmitter := &mockEmitter{}
		otelEmitter := &mockEmitter{}

		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter, logger.Test(t))
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), []byte("test message"), "key", "value")
		assert.NoError(t, err)
	})

	t.Run("otel emitter fails", func(t *testing.T) {
		chipEmitter := &mockEmitter{}
		otelEmitter := &mockEmitter{
			emitFunc: func(ctx context.Context, body []byte, attrKVs ...any) error {
				return errors.New("otel emit error")
			},
		}

		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter, logger.Test(t))
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), []byte("test message"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "otel emit error")
	})
}

func TestDualSourceEmitterBlockingBehavior(t *testing.T) {
	t.Run("chip ingress emit does not block caller", func(t *testing.T) {
		var chipCalled atomic.Bool
		chipEmitter := &mockEmitter{
			emitFunc: func(ctx context.Context, body []byte, attrKVs ...any) error {
				// Simulate slow work; the emitter itself is non-blocking
				// (fire-and-forget lives inside ChipIngressEmitter or batch service).
				time.Sleep(200 * time.Millisecond)
				chipCalled.Store(true)
				return nil
			},
		}
		otelEmitter := &mockEmitter{}

		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter, logger.Test(t))
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), []byte("test"))
		assert.NoError(t, err)

		require.NoError(t, emitter.Close())
	})

	t.Run("chip ingress emit completes inline when emitter is synchronous", func(t *testing.T) {
		var chipCalled atomic.Bool
		chipEmitter := &mockEmitter{
			emitFunc: func(ctx context.Context, body []byte, attrKVs ...any) error {
				chipCalled.Store(true)
				return nil
			},
		}
		otelEmitter := &mockEmitter{}

		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter, logger.Test(t))
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), []byte("test"))
		assert.NoError(t, err)
		assert.True(t, chipCalled.Load(),
			"chip ingress emit should complete before Emit returns")

		require.NoError(t, emitter.Close())
	})
}

// Mock emitter for testing
type mockEmitter struct {
	emitFunc func(ctx context.Context, body []byte, attrKVs ...any) error
}

func (m *mockEmitter) Close() error { return nil }

func (m *mockEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	if m.emitFunc != nil {
		return m.emitFunc(ctx, body, attrKVs...)
	}
	return nil
}
