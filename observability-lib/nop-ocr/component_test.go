package nopocr_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
	nopocr "github.com/smartcontractkit/chainlink-common/observability-lib/nop-ocr"

	"github.com/stretchr/testify/require"
)

func TestNewDashboard(t *testing.T) {
	t.Run("NewDashboard creates a dashboard", func(t *testing.T) {
		options := grafana.DashboardOptions{
			Name:              "NOP OCR Dashboard",
			MetricsDataSource: grafana.NewDataSource("Prometheus", ""),
		}
		testDashboard, err := nopocr.NewDashboard(&options)
		if err != nil {
			t.Errorf("Error creating dashboard: %v", err)
		}
		require.IsType(t, grafana.Dashboard{}, *testDashboard)
	})
}
