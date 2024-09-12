package atlasdon_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

	atlasdon "github.com/smartcontractkit/chainlink-common/observability-lib/atlas-don"
)

func TestNewDashboard(t *testing.T) {
	t.Run("NewDashboard creates a dashboard", func(t *testing.T) {
		testDashboard, err := atlasdon.NewDashboard(&atlasdon.Props{
			Name:              "DON OCR Dashboard",
			Platform:          grafana.TypePlatformDocker,
			MetricsDataSource: grafana.NewDataSource("Prometheus", ""),
			OCRVersion:        "ocr2",
		})
		if err != nil {
			t.Errorf("Error creating dashboard: %v", err)
		}
		require.IsType(t, grafana.Dashboard{}, *testDashboard)
		require.Equal(t, "DON OCR Dashboard", *testDashboard.Dashboard.Title)
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
