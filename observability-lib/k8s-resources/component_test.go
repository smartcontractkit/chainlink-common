package k8sresources_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
	k8sresources "github.com/smartcontractkit/chainlink-common/observability-lib/k8s-resources"

	"github.com/stretchr/testify/require"
)

func TestNewDashboard(t *testing.T) {
	t.Run("NewDashboard creates a dashboard", func(t *testing.T) {
		options := grafana.DashboardOptions{
			Name:              "K8s resources",
			Platform:          "kubernetes",
			MetricsDataSource: grafana.NewDataSource("Prometheus", ""),
		}
		testDashboard, err := k8sresources.NewDashboard(&options)
		if err != nil {
			t.Errorf("Error creating dashboard: %v", err)
		}
		require.IsType(t, grafana.Dashboard{}, *testDashboard)
	})
}
