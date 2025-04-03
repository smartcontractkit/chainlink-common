package beholder_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDualSourceEmitter(t *testing.T) {
	// Test successful creation
	t.Run("successful creation", func(t *testing.T) {
		
		chipEmitter := &mockEmitter{}
		otelEmitter := &mockEmitter{}
		
		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter)
		require.NoError(t, err)
		
		assert.NotNil(t, emitter)
		assert.IsType(t, &beholder.DualSourceEmitter{}, emitter)
	})
	
	// Test nil chip ingress emitter
	t.Run("nil chip ingress emitter", func(t *testing.T) {
		
		otelEmitter := &mockEmitter{}
		emitter, err := beholder.NewDualSourceEmitter(nil, otelEmitter)
		
		assert.Error(t, err)
		assert.Nil(t, emitter)
	})
	
	// Test nil otel collector emitter
	t.Run("nil otel collector emitter", func(t *testing.T) {
		
		chipEmitter := &mockEmitter{}
		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, nil)
		
		assert.Error(t, err)
		assert.Nil(t, emitter)
	})
}
func TestDualSourceEmitterEmit(t *testing.T) {
	t.Run("successful emit to both destinations", func(t *testing.T) {
		
		chipEmitter := &mockEmitter{}
		otelEmitter := &mockEmitter{}
		
		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter)
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
		
		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter)
		require.NoError(t, err)
		
		err = emitter.Emit(t.Context(), []byte("test message"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "otel emit error")
	})
	
	t.Run("chip emitter fails but doesn't affect result", func(t *testing.T) {
		
		chipCalled := false
		
		otelEmitter := &mockEmitter{}
		chipEmitter := &mockEmitter{
			emitFunc: func(ctx context.Context, body []byte, attrKVs ...any) error {
				chipCalled = true
				return fmt.Errorf("chip emit error")
			},
		}
		
		emitter, err := beholder.NewDualSourceEmitter(chipEmitter, otelEmitter)
		require.NoError(t, err)
				
		err = emitter.Emit(t.Context(), []byte("test message"))
		assert.NoError(t, err)
		
		// Wait for the goroutine to finish
		time.Sleep(time.Millisecond)
		assert.True(t, chipCalled, "chip emitter should have been called")
	})
}
// Mock emitter for testing
type mockEmitter struct {
	emitFunc func(ctx context.Context, body []byte, attrKVs ...any) error
}

func (m *mockEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	if m.emitFunc != nil {
		return m.emitFunc(ctx, body, attrKVs...)
	}
	return nil
}