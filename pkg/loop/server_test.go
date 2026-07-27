package loop

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/durableemitter"
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
