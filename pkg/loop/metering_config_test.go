package loop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/resourcemanager"
)

type fakeMeteringEmitter struct{}

func (fakeMeteringEmitter) Emit(context.Context, []byte, ...any) error { return nil }

func TestEnvConfig_MeteringConfig(t *testing.T) {
	t.Run("nil env yields zero config", func(t *testing.T) {
		var e *EnvConfig
		cfg := e.MeteringConfig(nil)
		assert.False(t, cfg.MeterRecordsEnabled)
		assert.Nil(t, cfg.Emitter)
	})

	t.Run("nil emitter is a valid no-op", func(t *testing.T) {
		e := &EnvConfig{MeterRecordsEnabled: true}
		cfg := e.MeteringConfig(nil)
		assert.Nil(t, cfg.Emitter)
	})

	t.Run("maps env fields and injects the given emitter", func(t *testing.T) {
		e := &EnvConfig{
			MeterRecordsEnabled:   true,
			MeterSnapshotsEnabled: true,
			MeterProduct:          "cre",
			MeterTenant:           "mainline",
			MeterNumericTenantID:  "42",
			MeterEnvironment:      "production",
			MeterZone:             "wf-zone-a",
			MeterNodeID:           "clp-cre-wf-zone-a-1",
		}
		emitter := fakeMeteringEmitter{}

		cfg := e.MeteringConfig(emitter)
		assert.True(t, cfg.MeterRecordsEnabled)
		assert.True(t, cfg.MeterSnapshotsEnabled)
		assert.Equal(t, resourcemanager.DefaultSnapshotInterval, cfg.SnapshotInterval)
		require.NotNil(t, cfg.Emitter)
		assert.Equal(t, resourcemanager.Emitter(emitter), cfg.Emitter)
		assert.Equal(t, "cre", cfg.DeploymentIdentity.Product)
		assert.Equal(t, "mainline", cfg.DeploymentIdentity.Tenant)
		assert.Equal(t, "42", cfg.DeploymentIdentity.NumericTenantID)
		assert.Equal(t, "production", cfg.DeploymentIdentity.Environment)
		assert.Equal(t, "wf-zone-a", cfg.DeploymentIdentity.Zone)
		assert.Equal(t, "clp-cre-wf-zone-a-1", cfg.DeploymentIdentity.NodeID)
	})
}
