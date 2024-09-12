package k8sresources_test

import (
	"os"
	"testing"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
	k8sresources "github.com/smartcontractkit/chainlink-common/observability-lib/k8s-resources"

	"github.com/stretchr/testify/require"
)

func TestNewDashboard(t *testing.T) {
	t.Run("NewDashboard creates a dashboard", func(t *testing.T) {
		testDashboard, err := k8sresources.NewDashboard(&k8sresources.Props{
			Name:              "K8s resources",
			MetricsDataSource: grafana.NewDataSource("Prometheus", ""),
		})
		if err != nil {
			t.Errorf("Error creating dashboard: %v", err)
		}
		require.IsType(t, grafana.Dashboard{}, *testDashboard)
		require.Equal(t, "K8s resources", *testDashboard.Dashboard.Title)
		json, errJSON := testDashboard.GenerateJSON()
		if errJSON != nil {
			t.Errorf("Error generating JSON: %v", errJSON)
		}

		jsonCompared, errCompared := os.ReadFile("test-output.json")
		if errCompared != nil {
			t.Errorf("Error reading file: %v", errCompared)
		}

		require.ElementsMatch(t, jsonCompared, json)
	})
}
