package grafana_test

import (
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/expr"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
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
				Title:      grafana.Pointer("ETH Balance"),
				Span:       12,
				Height:     6,
				Decimals:   grafana.Pointer(2.0),
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
							ReduceExpression: &grafana.ReduceExpression{
								Expression: "A",
								Reducer:    expr.TypeReduceReducerSum,
								ReduceSettings: &expr.ExprTypeReduceSettings{
									Mode: expr.ExprTypeReduceSettingsModeDropNN,
								},
							},
						},
						{
							RefID: "C",
							ThresholdExpression: &grafana.ThresholdExpression{
								Expression: "B",
								ThresholdConditionsOptions: grafana.ThresholdConditionsOption{
									Params: []float64{2},
									Type:   expr.ExprTypeThresholdConditionsEvaluatorTypeLt,
								},
							},
						},
					},
				},
			},
		}))

		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}

		json, err := o.GenerateJSON()
		t.Log(string(json))
		require.NoError(t, err)
		require.IsType(t, json, []byte{})
	})
}
