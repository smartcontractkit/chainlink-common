package grafana_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
	"github.com/stretchr/testify/require"
)

func TestGenerateJSON(t *testing.T) {
	t.Run("GenerateJSON return JSON from dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name:     "Dashboard Name",
			Tags:     []string{"foo", "bar"},
			Refresh:  "1m",
			TimeFrom: "now-1h",
			TimeTo:   "now",
			TimeZone: "UTC",
		})

		builder.AddPanel(grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: "datasource-name",
				Title:      "ETH Balance",
				Span:       12,
				Height:     6,
				Decimals:   2,
				Query: []grafana.Query{
					{
						Expr:   `eth_balance`,
						Legend: `{{account}}`,
					},
				},
			},
			AlertsOptions: []grafana.AlertOptions{
				{
					Summary:     `ETH Balance is lower than threshold`,
					Description: `ETH Balance critically low at {{ index $values "A" }}`,
					RunbookURL:  "https://github.com/smartcontractkit/chainlink-common/tree/main/observability-lib",
					For:         "1m",
					Tags: map[string]string{
						"severity": "warning",
					},
					Query: []grafana.RuleQuery{
						{
							Expr:       `eth_balance`,
							Instant:    true,
							RefID:      "A",
							Datasource: "datasource-uid",
						},
					},
					QueryRefCondition: "B",
					Condition: []grafana.ConditionQuery{
						{
							RefID: "B",
							ThresholdExpression: &grafana.ThresholdExpression{
								Expression: "A",
								ThresholdConditionsOptions: grafana.ThresholdConditionsOption{
									Params: []float64{2},
									Type:   grafana.TypeThresholdTypeLt,
								},
							},
						},
					},
				},
			},
		}))

		db, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}

		json, err := db.GenerateJSON()
		require.IsType(t, json, []byte{})
	})
}
