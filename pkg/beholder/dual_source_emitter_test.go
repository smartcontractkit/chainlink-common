package beholder_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func TestNewDualSourceEmitter(t *testing.T) {
	// Test successful creation
	t.Run("successful creation", func(t *testing.T) {

		chipEmitter := &mockEmitter{}
		otelEmitter := &mockEmitter{}

		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter, false)
		require.NoError(t, err)

		assert.NotNil(t, emitter)
		assert.IsType(t, &beholder.DualSourceEmitter{}, emitter)
	})

	// Test nil chip ingress emitter
	t.Run("nil chip ingress emitter", func(t *testing.T) {

		otelEmitter := &mockEmitter{}
		emitter, err := beholder.NewDualSourceEmitter(nil, otelEmitter, false)

		assert.Error(t, err)
		assert.Nil(t, emitter)
	})

	// Test nil otel collector emitter
	t.Run("nil otel collector emitter", func(t *testing.T) {

		chipEmitter := &mockEmitter{}
		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, nil, false)

		assert.Error(t, err)
		assert.Nil(t, emitter)
	})
}
func TestDualSourceEmitterEmit(t *testing.T) {
	t.Run("successful emit to both destinations", func(t *testing.T) {

		chipEmitter := &mockEmitter{}
		otelEmitter := &mockEmitter{}

		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter, false)
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), []byte("test message"), "key", "value")
		assert.NoError(t, err)
	})

	t.Run("otel emitter fails", func(t *testing.T) {

		chipEmitter := &mockEmitter{}
		otelEmitter := &mockEmitter{
			emitFunc: func(ctx context.Context, body []byte, attrKVs ...any) error {
				return fmt.Errorf("otel emit error")
			},
		}

		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter, false)
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), []byte("test message"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "otel emit error")
	})
}

func TestDualSourceEmitterBlockingBehavior(t *testing.T) {
	t.Run("legacy mode does not block caller", func(t *testing.T) {
		var chipCalled atomic.Bool
		chipEmitter := &mockEmitter{
			emitFunc: func(ctx context.Context, body []byte, attrKVs ...any) error {
				time.Sleep(200 * time.Millisecond)
				chipCalled.Store(true)
				return nil
			},
		}
		otelEmitter := &mockEmitter{}

		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter, false)
		require.NoError(t, err)

		start := time.Now()
		err = emitter.Emit(t.Context(), []byte("test"))
		elapsed := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, elapsed, 100*time.Millisecond,
			"Emit should return immediately; chip ingress runs in a goroutine")
		assert.False(t, chipCalled.Load(),
			"chip ingress emit should still be in-flight when Emit returns")

		require.NoError(t, emitter.Close())
		assert.True(t, chipCalled.Load(),
			"Close should wait for the in-flight chip ingress emit to finish")
	})

	t.Run("non-blocking mode emits inline", func(t *testing.T) {
		var chipCalled atomic.Bool
		chipEmitter := &mockEmitter{
			emitFunc: func(ctx context.Context, body []byte, attrKVs ...any) error {
				chipCalled.Store(true)
				return nil
			},
		}
		otelEmitter := &mockEmitter{}

		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter, true)
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
