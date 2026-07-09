package resourcemanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/loop"
)

func TestNewBaseIdentity(t *testing.T) {
	dep := DeploymentIdentity{
		Product:         "cre",
		Tenant:          "mainline",
		NumericTenantID: "42",
		Environment:     "production",
		Zone:            "wf-zone-a",
		NodeID:          "clp-cre-wf-zone-a-1",
	}

	t.Run("with CapDONID sets Don", func(t *testing.T) {
		id := NewBaseIdentity(dep, 7, "cron-trigger", "trigger_registrations")
		assert.Equal(t, "cre", id.Product)
		assert.Equal(t, "cron-trigger", id.Service)
		assert.Equal(t, "trigger_registrations", id.ResourcePool)
		require.NotNil(t, id.Don)
		assert.Equal(t, "7", id.Don.DonID)
		assert.Equal(t, "clp-cre-wf-zone-a-1", id.Don.NodeID)
	})

	t.Run("product falls back to cre", func(t *testing.T) {
		d := dep
		d.Product = ""
		id := NewBaseIdentity(d, 7, "svc", "pool")
		assert.Equal(t, DefaultMeteringProduct, id.Product)
	})

	t.Run("CapDONID 0 leaves don_id empty but keeps node_id", func(t *testing.T) {
		id := NewBaseIdentity(dep, 0, "svc", "pool")
		require.NotNil(t, id.Don)
		assert.Empty(t, id.Don.DonID)
		assert.Equal(t, "clp-cre-wf-zone-a-1", id.Don.NodeID)
	})

	t.Run("no don id or node id leaves Don nil", func(t *testing.T) {
		d := dep
		d.NodeID = ""
		id := NewBaseIdentity(d, 0, "svc", "pool")
		assert.Nil(t, id.Don)
	})
}

func TestWithWorkflowDonFallback(t *testing.T) {
	t.Run("applies workflow DON when CapDONID absent", func(t *testing.T) {
		base := NewBaseIdentity(DeploymentIdentity{NodeID: "node-1"}, 0, "svc", "pool")
		got := base.WithWorkflowDonFallback(99)
		require.NotNil(t, got.Don)
		assert.Equal(t, "99", got.Don.DonID)
		assert.Equal(t, "node-1", got.Don.NodeID)
	})

	t.Run("keeps authoritative CapDONID over workflow DON", func(t *testing.T) {
		base := NewBaseIdentity(DeploymentIdentity{NodeID: "node-1"}, 5, "svc", "pool")
		got := base.WithWorkflowDonFallback(99)
		assert.Equal(t, "5", got.Don.DonID)
	})

	t.Run("workflow DON 0 is a no-op", func(t *testing.T) {
		base := NewBaseIdentity(DeploymentIdentity{NodeID: "node-1"}, 0, "svc", "pool")
		got := base.WithWorkflowDonFallback(0)
		assert.Empty(t, got.DonID())
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
		assert.NotNil(t, cfg.Emitter)
		assert.Equal(t, "cre", cfg.DeploymentIdentity.Product)
		assert.Equal(t, "mainline", cfg.DeploymentIdentity.Tenant)
		assert.Equal(t, "42", cfg.DeploymentIdentity.NumericTenantID)
		assert.Equal(t, "production", cfg.DeploymentIdentity.Environment)
		assert.Equal(t, "wf-zone-a", cfg.DeploymentIdentity.Zone)
		assert.Equal(t, "clp-cre-wf-zone-a-1", cfg.DeploymentIdentity.NodeID)
	})
}
