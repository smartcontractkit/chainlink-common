package corenode_test

import (
	"testing"

	corenode "github.com/smartcontractkit/chainlink-common/observability-lib/core-node"
	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

	"github.com/stretchr/testify/require"
)

func TestNewDashboard(t *testing.T) {
	t.Run("NewDashboard creates a dashboard", func(t *testing.T) {
		options := grafana.DashboardOptions{
			Name:              "Core Node Dashboard",
			MetricsDataSource: grafana.NewDataSource("Prometheus", ""),
			Platform:          grafana.TypePlatformDocker,
		}
		testDashboard, err := corenode.NewDashboard(&options)
		if err != nil {
			t.Errorf("Error creating dashboard: %v", err)
		}
		require.IsType(t, grafana.Dashboard{}, *testDashboard)
	})
}
