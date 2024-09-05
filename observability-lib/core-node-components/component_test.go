package corenodecomponents_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

	corenodecomponents "github.com/smartcontractkit/chainlink-common/observability-lib/core-node-components"
)

func TestNewDashboard(t *testing.T) {
	t.Run("NewDashboard creates a dashboard", func(t *testing.T) {
		options := grafana.DashboardOptions{
			Name:              "Core Node Components Dashboard",
			MetricsDataSource: grafana.NewDataSource("Prometheus", ""),
			LogsDataSource:    grafana.NewDataSource("Loki", ""),
		}
		testDashboard, err := corenodecomponents.NewDashboard(&options)
		if err != nil {
			t.Errorf("Error creating dashboard: %v", err)
		}
		require.IsType(t, grafana.Dashboard{}, *testDashboard)
	})
}
