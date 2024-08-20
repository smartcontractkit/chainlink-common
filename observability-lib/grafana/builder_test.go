package grafana_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

func TestNewBuilder(t *testing.T) {
	t.Run("NewBuilder builds a dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.DashboardOptions{
			Name: "Dashboard Name",
		}, &grafana.BuilderOptions{
			Tags:     []string{"foo", "bar"},
			Refresh:  "1m",
			TimeFrom: "now-1h",
			TimeTo:   "now",
			TimeZone: "UTC",
		})

		db, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}

		require.IsType(t, dashboard.Dashboard{}, *db.Dashboard)
	})
}

func TestBuilder_AddVars(t *testing.T) {
	t.Run("AddVars adds variables to the dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.DashboardOptions{
			Name: "Dashboard Name",
		}, &grafana.BuilderOptions{})

		variable := grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Name:  "Variable Name",
				Label: "Variable Label",
			},
			Query:      "query",
			Datasource: grafana.NewDataSource("Prometheus", "").Name,
		})

		builder.AddVars(variable)
		db, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}
		require.IsType(t, dashboard.Dashboard{}, *db.Dashboard)
	})
}

func TestBuilder_AddRow(t *testing.T) {
	t.Run("AddRow adds a row to the dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.DashboardOptions{
			Name: "Dashboard Name",
		}, &grafana.BuilderOptions{})

		builder.AddRow("Row Title")
		db, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}
		require.IsType(t, dashboard.Dashboard{}, *db.Dashboard)
	})
}

func TestBuilder_AddPanel(t *testing.T) {
	t.Run("AddPanel adds a panel to the dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.DashboardOptions{
			Name: "Dashboard Name",
		}, &grafana.BuilderOptions{})

		panel := grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: "Panel Title",
			},
		})

		builder.AddPanel(panel)
		db, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}
		require.IsType(t, dashboard.Dashboard{}, *db.Dashboard)
	})
}
