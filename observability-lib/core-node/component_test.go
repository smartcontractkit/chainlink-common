package corenode_test

import (
	"fmt"
	"os"
	"testing"

	corenode "github.com/smartcontractkit/chainlink-common/observability-lib/core-node"
	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

	"github.com/stretchr/testify/require"
)

func TestNewDashboard(t *testing.T) {
	t.Run("NewDashboard creates a dashboard", func(t *testing.T) {
		testDashboard, err := corenode.NewDashboard(&corenode.Props{
			Name:              "Core Node Dashboard",
			Platform:          grafana.TypePlatformDocker,
			MetricsDataSource: grafana.NewDataSource("Prometheus", "1"),
		})
		if err != nil {
			t.Errorf("Error creating dashboard: %v", err)
		}
		require.IsType(t, grafana.Dashboard{}, *testDashboard)
		require.Equal(t, "Core Node Dashboard", *testDashboard.Dashboard.Title)
		json, errJSON := testDashboard.GenerateJSON()
		if errJSON != nil {
			t.Errorf("Error generating JSON: %v", errJSON)
		}
		fmt.Println(string(json))

		_, errCompared := os.ReadFile("test-output.json")
		if errCompared != nil {
			t.Errorf("Error reading file: %v", errCompared)
		}

		//require.ElementsMatch(t, jsonCompared, json)
	})
}
