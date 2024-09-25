package corenodecomponents_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

	corenodecomponents "github.com/smartcontractkit/chainlink-common/observability-lib/core-node-components"
)

func TestNewDashboard(t *testing.T) {
	t.Run("NewDashboard creates a dashboard", func(t *testing.T) {
		testDashboard, err := corenodecomponents.NewDashboard(&corenodecomponents.Props{
			Name:              "Core Node Components Dashboard",
			MetricsDataSource: grafana.NewDataSource("Prometheus", ""),
			LogsDataSource:    grafana.NewDataSource("Loki", ""),
		})
		if err != nil {
			t.Errorf("Error creating dashboard: %v", err)
		}
		require.IsType(t, grafana.Dashboard{}, *testDashboard)
		require.Equal(t, "Core Node Components Dashboard", *testDashboard.Dashboard.Title)
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
