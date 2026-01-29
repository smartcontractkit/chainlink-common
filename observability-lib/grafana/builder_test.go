package grafana_test

import (
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/expr"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

func TestNewBuilder(t *testing.T) {
	t.Run("NewBuilder builds a dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name:     "Dashboard Name",
			Tags:     []string{"foo", "bar"},
			Refresh:  "1m",
			TimeFrom: "now-1h",
			TimeTo:   "now",
			TimeZone: "UTC",
		})

		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error during build: %v", err)
		}

		require.NotEmpty(t, o.Dashboard)
		require.Empty(t, o.Alerts)
		require.Empty(t, o.ContactPoints)
		require.Empty(t, o.NotificationPolicies)
	})

	t.Run("NewBuilder builds a dashboard with alerts", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name:     "Dashboard Name",
			Tags:     []string{"foo", "bar"},
			Refresh:  "1m",
			TimeFrom: "now-1h",
			TimeTo:   "now",
			TimeZone: "UTC",
		})
		builder.AddAlert(grafana.NewAlertRule(&grafana.AlertOptions{
			Title: "Alert Title",
		}))

		builder.AddAlert(grafana.NewAlertRule(&grafana.AlertOptions{
			Title:          "Alert Title 2",
			RuleGroupTitle: "RuleGroup Title",
		}))

		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error during build: %v", err)
		}
		require.NotEmpty(t, o.Dashboard)
		require.NotEmpty(t, o.Alerts)
		require.Len(t, o.Alerts, 2)
		require.Equal(t, "Default", o.Alerts[0].RuleGroup)
		require.Equal(t, "RuleGroup Title", o.Alerts[1].RuleGroup)

		require.Empty(t, o.ContactPoints)
		require.Empty(t, o.NotificationPolicies)
	})

	t.Run("NewBuilder builds only alerts", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{})
		builder.AddAlert(grafana.NewAlertRule(&grafana.AlertOptions{
			Title: "Alert Title",
		}))

		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error during build: %v", err)
		}
		require.Empty(t, o.Dashboard)
		require.NotEmpty(t, o.Alerts)
		require.Len(t, o.Alerts, 1)
		require.Empty(t, o.ContactPoints)
		require.Empty(t, o.NotificationPolicies)
	})

	t.Run("NewBuilder builds an alert group", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{})
		builder.AddAlertGroup(grafana.NewAlertGroup(&grafana.AlertGroupOptions{
			Title:    "Group Title",
			Interval: 30, // duration in seconds
		}))

		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error during build: %v", err)
		}
		require.Empty(t, o.Dashboard)
		require.NotEmpty(t, o.AlertGroups)
		require.Len(t, o.AlertGroups, 1)
		require.Empty(t, o.ContactPoints)
		require.Empty(t, o.NotificationPolicies)
	})

	t.Run("NewBuilder builds a contact point", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{})
		builder.AddContactPoint(grafana.NewContactPoint(&grafana.ContactPointOptions{
			Name: "slack",
			Type: "slack",
		}))

		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error during build: %v", err)
		}

		require.Empty(t, o.Dashboard)
		require.Empty(t, o.Alerts)
		require.NotEmpty(t, o.ContactPoints)
		require.Empty(t, o.NotificationPolicies)
	})

	t.Run("NewBuilder builds a notification policy", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{})
		builder.AddNotificationPolicy(grafana.NewNotificationPolicy(&grafana.NotificationPolicyOptions{
			Receiver: "slack",
			GroupBy:  []string{"grafana_folder", "alertname"},
			ObjectMatchers: []alerting.ObjectMatcher{
				{"team", "=", "chainlink"},
			},
		}))

		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error during build: %v", err)
		}

		require.Empty(t, o.Dashboard)
		require.Empty(t, o.Alerts)
		require.Empty(t, o.ContactPoints)
		require.NotEmpty(t, o.NotificationPolicies)
	})
}

func TestBuilder_AddVars(t *testing.T) {
	t.Run("AddVars adds variables to the dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})

		variable := grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Name:  "Variable Name",
				Label: "Variable Label",
			},
			Query:      "query",
			Datasource: grafana.NewDataSource("Prometheus", "").Name,
		})

		builder.AddVars(variable)
		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}
		require.Len(t, o.Dashboard.Templating.List, 1)
	})

	t.Run("AddVars adds variables with AllValue to the dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})

		variable := grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Name:  "Variable Name",
				Label: "Variable Label",
			},
			Query:      "query",
			Datasource: grafana.NewDataSource("Prometheus", "").Name,
			IncludeAll: true,
			AllValue:   ".*",
		})

		builder.AddVars(variable)
		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}
		require.Len(t, o.Dashboard.Templating.List, 1)

		// Verify the AllValue is set correctly
		varModel := o.Dashboard.Templating.List[0]
		require.NotNil(t, varModel.AllValue)
		require.Equal(t, ".*", *varModel.AllValue)
		require.NotNil(t, varModel.IncludeAll)
		require.True(t, *varModel.IncludeAll)
	})
}

func TestBuilder_AddRow(t *testing.T) {
	t.Run("AddRow adds a row to the dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})

		builder.AddRow("Row Title")
		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}
		require.IsType(t, dashboard.RowPanel{}, *o.Dashboard.Panels[0].RowPanel)
	})
}

func TestBuilder_AddPanel(t *testing.T) {
	t.Run("AddPanel adds a panel to the dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})

		panel := grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Panel Title"),
			},
		})

		builder.AddPanel(panel)
		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}
		require.IsType(t, dashboard.Panel{}, *o.Dashboard.Panels[0].Panel)
	})
}

func TestBuilder_AddTimeSeriesPanelWithAlert(t *testing.T) {
	t.Run("AddPanel adds a panel to the dashboard", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})

		panel := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Panel Title"),
			},
			AlertsOptions: []grafana.AlertOptions{
				{
					Summary:     `Test Alert Summary`,
					Description: `Test Description with value {{ index $values "A" }}`,
					RunbookURL:  "https://github.com/smartcontractkit/chainlink-common/tree/main/observability-lib",
					For:         "1m",
					Tags: map[string]string{
						"severity": "warning",
					},
					Query: []grafana.RuleQuery{
						{
							Expr:       `my_metric`,
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
		})

		builder.AddPanel(panel)
		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error building dashboard: %v", err)
		}
		require.IsType(t, dashboard.Panel{}, *o.Dashboard.Panels[0].Panel)
		require.Len(t, o.Alerts, 1)
		// RuleGroup defaults to Panel Title
		require.Equal(t, "Panel Title", o.Alerts[0].RuleGroup)
	})
}
