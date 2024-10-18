package nopocr_test

import (
	"os"
	"testing"

	nopocr "github.com/smartcontractkit/chainlink-common/observability-lib/dashboards/nop-ocr"
	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

	"github.com/stretchr/testify/require"
)

func TestNewDashboard(t *testing.T) {
	t.Run("NewDashboard creates a dashboard", func(t *testing.T) {
		testDashboard, err := nopocr.NewDashboard(&nopocr.Props{
			Name:              "NOP OCR Dashboard",
			MetricsDataSource: grafana.NewDataSource("Prometheus", ""),
		})
		if err != nil {
			t.Errorf("Error creating dashboard: %v", err)
		}
		require.IsType(t, grafana.Dashboard{}, *testDashboard)
		require.Equal(t, "NOP OCR Dashboard", *testDashboard.Dashboard.Title)
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
