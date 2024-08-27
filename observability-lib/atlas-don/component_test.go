package atlasdon_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

	atlasdon "github.com/smartcontractkit/chainlink-common/observability-lib/atlas-don"
)

func TestNewDashboard(t *testing.T) {
	t.Run("NewDashboard creates a dashboard", func(t *testing.T) {
		options := grafana.DashboardOptions{
			Name:              "DON OCR Dashboard",
			MetricsDataSource: grafana.NewDataSource("Prometheus", ""),
			Platform:          grafana.TypePlatformDocker,
			OCRVersion:        "ocr2",
		}
		testDashboard, err := atlasdon.NewDashboard(&options)
		if err != nil {
			t.Errorf("Error creating dashboard: %v", err)
		}
		require.IsType(t, grafana.Dashboard{}, *testDashboard)
	})
}
