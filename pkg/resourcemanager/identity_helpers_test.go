package resourcemanager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/durableemitter"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
)

// stubDurableStore is a no-op DurableEventStore so tests can construct a real
// *DurableEmitter without a database.
type stubDurableStore struct{}

func (stubDurableStore) Insert(context.Context, []byte) (int64, error) { return 0, nil }
func (stubDurableStore) Delete(context.Context, int64) error           { return nil }
func (stubDurableStore) BatchDelete(context.Context, []int64) (int64, error) {
	return 0, nil
}
func (stubDurableStore) ListPending(context.Context, time.Time, time.Time, int64, int) ([]durableemitter.DurableEvent, error) {
	return nil, nil
}
func (stubDurableStore) DeleteExpired(context.Context, time.Duration) (int64, error) {
	return 0, nil
}

// stubBatchEmitter is a no-op BatchEmitter.
type stubBatchEmitter struct{}

func (stubBatchEmitter) QueueMessage(*chipingress.CloudEventPb, func(error)) error { return nil }
func (stubBatchEmitter) Start(context.Context)                                     {}
func (stubBatchEmitter) Stop()                                                     {}

func TestNewBaseIdentity(t *testing.T) {
	dep := DeploymentIdentity{
		Product:         "cre",
		Tenant:          "mainline",
		NumericTenantID: "42",
		Environment:     "production",
		Zone:            "wf-zone-a",
		NodeID:          "clp-cre-wf-zone-a-1",
	}

	t.Run("sets coarse dimensions and node id", func(t *testing.T) {
		id := NewBaseIdentity(dep, "cron-trigger", "trigger_registrations")
		assert.Equal(t, "cre", id.Product)
		assert.Equal(t, "cron-trigger", id.Service)
		assert.Equal(t, "trigger_registrations", id.ResourcePool)
		require.NotNil(t, id.Don)
		assert.Empty(t, id.Don.DonID)
		assert.Equal(t, "clp-cre-wf-zone-a-1", id.Don.NodeID)
	})

	t.Run("product falls back to unset", func(t *testing.T) {
		d := dep
		d.Product = ""
		id := NewBaseIdentity(d, "svc", "pool")
		assert.Equal(t, UnsetProduct, id.Product)
	})

	t.Run("no node id leaves Don nil", func(t *testing.T) {
		d := dep
		d.NodeID = ""
		id := NewBaseIdentity(d, "svc", "pool")
		assert.Nil(t, id.Don)
	})
}

func TestWithDonID(t *testing.T) {
	base := NewBaseIdentity(DeploymentIdentity{NodeID: "node-1"}, "svc", "pool")

	t.Run("stamps DON id preserving node id", func(t *testing.T) {
		got := base.WithDonID("7")
		require.NotNil(t, got.Don)
		assert.Equal(t, "7", got.Don.DonID)
		assert.Equal(t, "node-1", got.Don.NodeID)
	})

	t.Run("empty DON id is a no-op", func(t *testing.T) {
		got := base.WithDonID("")
		assert.Empty(t, got.DonID())
		assert.Equal(t, "node-1", got.NodeID())
	})

	t.Run("overwrites previous DON id", func(t *testing.T) {
		first := base.WithDonID("5")
		second := first.WithDonID("99")
		assert.Equal(t, "99", second.DonID())
		assert.Equal(t, "node-1", second.NodeID())
	})
}

func TestConfigFromEnv(t *testing.T) {
	t.Run("nil env yields zero config", func(t *testing.T) {
		cfg := ConfigFromEnv(nil)
		assert.False(t, cfg.MeterRecordsEnabled)
		assert.Nil(t, cfg.Emitter)
	})

	t.Run("maps env fields", func(t *testing.T) {
		cfg := ConfigFromEnv(&loop.EnvConfig{
			MeterRecordsEnabled:   true,
			MeterSnapshotsEnabled: true,
			MeterProduct:          "cre",
			MeterTenant:           "mainline",
			MeterNumericTenantID:  "42",
			MeterEnvironment:      "production",
			MeterZone:             "wf-zone-a",
			MeterNodeID:           "clp-cre-wf-zone-a-1",
		})
		assert.True(t, cfg.MeterRecordsEnabled)
		assert.True(t, cfg.MeterSnapshotsEnabled)
		assert.Equal(t, DefaultSnapshotInterval, cfg.SnapshotInterval)
		// The durable emitter is not initialized in this test binary, so the
		// emitter falls back to a true nil interface (no-op, fail-open) rather
		// than a non-nil interface wrapping a nil *DurableEmitter.
		assert.Nil(t, cfg.Emitter)
		assert.Equal(t, "cre", cfg.DeploymentIdentity.Product)
		assert.Equal(t, "mainline", cfg.DeploymentIdentity.Tenant)
		assert.Equal(t, "42", cfg.DeploymentIdentity.NumericTenantID)
		assert.Equal(t, "production", cfg.DeploymentIdentity.Environment)
		assert.Equal(t, "wf-zone-a", cfg.DeploymentIdentity.Zone)
		assert.Equal(t, "clp-cre-wf-zone-a-1", cfg.DeploymentIdentity.NodeID)
	})

	t.Run("wires the global durable emitter when initialized", func(t *testing.T) {
		de, err := durableemitter.NewDurableEmitter(stubDurableStore{}, stubBatchEmitter{}, false, durableemitter.Config{}, logger.Test(t), nil)
		require.NoError(t, err)
		durableemitter.SetGlobalEmitter(de)
		t.Cleanup(func() { durableemitter.SetGlobalEmitter(nil) })

		cfg := ConfigFromEnv(&loop.EnvConfig{MeterRecordsEnabled: true})
		require.NotNil(t, cfg.Emitter)
		assert.Equal(t, Emitter(de), cfg.Emitter)
	})
}
