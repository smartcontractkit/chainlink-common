package resourcemanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
