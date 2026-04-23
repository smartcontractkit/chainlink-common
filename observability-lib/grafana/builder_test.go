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

	t.Run("NewBuilder builds a dashboard with row and panels inside", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})
		builder.AddRow("Row Title")
		builder.AddPanelToRow("Row Title", grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Panel Title"),
			},
		}))

		o, err := builder.Build()
		if err != nil {
			t.Errorf("Error during build: %v", err)
		}
		require.NotEmpty(t, o.Dashboard)
		require.Len(t, o.Dashboard.Panels, 1)
		rowPanel := o.Dashboard.Panels[0]
		require.IsType(t, dashboard.RowPanel{}, *rowPanel.RowPanel)
		require.True(t, rowPanel.RowPanel.Collapsed)
		require.Len(t, rowPanel.RowPanel.Panels, 1)
	})

	t.Run("NewBuilder builds a dashboard with row and multiple panels inside", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})
		builder.AddRow("Row Title")
		builder.AddPanelToRow("Row Title",
			grafana.NewStatPanel(&grafana.StatPanelOptions{
				PanelOptions: &grafana.PanelOptions{
					Title: grafana.Pointer("Stat Panel"),
				},
			}),
			grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
				PanelOptions: &grafana.PanelOptions{
					Title: grafana.Pointer("TimeSeries Panel"),
				},
			}),
			grafana.NewTablePanel(&grafana.TablePanelOptions{
				PanelOptions: &grafana.PanelOptions{
					Title: grafana.Pointer("Table Panel"),
				},
			}),
		)

		o, err := builder.Build()
		require.NoError(t, err)
		require.Len(t, o.Dashboard.Panels, 1)
		rowPanel := o.Dashboard.Panels[0]
		require.NotNil(t, rowPanel.RowPanel)
		require.True(t, rowPanel.RowPanel.Collapsed)
		require.Len(t, rowPanel.RowPanel.Panels, 3)
	})

	t.Run("NewBuilder preserves order with interleaved AddPanel and AddRow", func(t *testing.T) {
		// Layout: top-level panel, then a row, then another top-level panel
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})
		builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Top Panel 1"),
			},
		}))
		builder.AddRow("Row A")
		builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Top Panel 2"),
			},
		}))

		o, err := builder.Build()
		require.NoError(t, err)
		require.Len(t, o.Dashboard.Panels, 3)
		// First: top-level panel
		require.NotNil(t, o.Dashboard.Panels[0].Panel)
		require.Equal(t, "Top Panel 1", *o.Dashboard.Panels[0].Panel.Title)
		// Second: row
		require.NotNil(t, o.Dashboard.Panels[1].RowPanel)
		require.Equal(t, "Row A", *o.Dashboard.Panels[1].RowPanel.Title)
		// Third: top-level panel
		require.NotNil(t, o.Dashboard.Panels[2].Panel)
		require.Equal(t, "Top Panel 2", *o.Dashboard.Panels[2].Panel.Title)
	})

	t.Run("NewBuilder mixed rows with and without panels preserve order", func(t *testing.T) {
		// Layout: row without panels, top-level panel, row with 2 panels, top-level panel
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})
		builder.AddRow("Open Row")
		builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Panel After Open Row"),
			},
		}))
		builder.AddRow("Row With Panels")
		builder.AddPanelToRow("Row With Panels",
			grafana.NewStatPanel(&grafana.StatPanelOptions{
				PanelOptions: &grafana.PanelOptions{
					Title: grafana.Pointer("Inside Row 1"),
				},
			}),
			grafana.NewGaugePanel(&grafana.GaugePanelOptions{
				PanelOptions: &grafana.PanelOptions{
					Title: grafana.Pointer("Inside Row 2"),
				},
			}),
		)
		builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Panel After Row With Panels"),
			},
		}))

		o, err := builder.Build()
		require.NoError(t, err)
		// Expected top-level: Open Row, Panel After Open Row, Row With Panels, Panel After Row With Panels
		require.Len(t, o.Dashboard.Panels, 4)

		// 1. Row without panels
		require.NotNil(t, o.Dashboard.Panels[0].RowPanel)
		require.Equal(t, "Open Row", *o.Dashboard.Panels[0].RowPanel.Title)
		require.False(t, o.Dashboard.Panels[0].RowPanel.Collapsed)

		// 2. Top-level panel after the open row
		require.NotNil(t, o.Dashboard.Panels[1].Panel)
		require.Equal(t, "Panel After Open Row", *o.Dashboard.Panels[1].Panel.Title)

		// 3. Row with its 2 panels nested inside (automatically collapsed)
		require.NotNil(t, o.Dashboard.Panels[2].RowPanel)
		require.Equal(t, "Row With Panels", *o.Dashboard.Panels[2].RowPanel.Title)
		require.True(t, o.Dashboard.Panels[2].RowPanel.Collapsed)
		require.Len(t, o.Dashboard.Panels[2].RowPanel.Panels, 2)

		// 4. Top-level panel after the row with panels
		require.NotNil(t, o.Dashboard.Panels[3].Panel)
		require.Equal(t, "Panel After Row With Panels", *o.Dashboard.Panels[3].Panel.Title)
	})

	t.Run("NewBuilder multiple rows each with their own panels", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})
		builder.AddRow("Row A")
		builder.AddPanelToRow("Row A", grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Panel in A"),
			},
		}))
		builder.AddRow("Row B")
		builder.AddPanelToRow("Row B",
			grafana.NewStatPanel(&grafana.StatPanelOptions{
				PanelOptions: &grafana.PanelOptions{
					Title: grafana.Pointer("Panel in B1"),
				},
			}),
			grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
				PanelOptions: &grafana.PanelOptions{
					Title: grafana.Pointer("Panel in B2"),
				},
			}),
		)

		o, err := builder.Build()
		require.NoError(t, err)
		require.Len(t, o.Dashboard.Panels, 2)

		// First row
		require.NotNil(t, o.Dashboard.Panels[0].RowPanel)
		require.Equal(t, "Row A", *o.Dashboard.Panels[0].RowPanel.Title)
		require.Len(t, o.Dashboard.Panels[0].RowPanel.Panels, 1)

		// Second row
		require.NotNil(t, o.Dashboard.Panels[1].RowPanel)
		require.Equal(t, "Row B", *o.Dashboard.Panels[1].RowPanel.Title)
		require.Len(t, o.Dashboard.Panels[1].RowPanel.Panels, 2)
	})

	t.Run("NewBuilder row without panels is not collapsed", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})
		builder.AddRow("Open Row")

		o, err := builder.Build()
		require.NoError(t, err)
		require.Len(t, o.Dashboard.Panels, 1)
		require.NotNil(t, o.Dashboard.Panels[0].RowPanel)
		require.False(t, o.Dashboard.Panels[0].RowPanel.Collapsed)
		require.Empty(t, o.Dashboard.Panels[0].RowPanel.Panels)
	})
}

func TestBuilder_BuildOnce(t *testing.T) {
	t.Run("Build returns error on second call", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})

		_, err := builder.Build()
		require.NoError(t, err)

		_, err = builder.Build()
		require.Error(t, err)
		require.Contains(t, err.Error(), "already been called")
	})

	t.Run("Build returns error when panels added without dashboard name", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{})
		builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Panel Title"),
			},
		}))

		_, err := builder.Build()
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot add rows or panels without a dashboard")
	})

	t.Run("Build returns error when AddPanelToRow references unknown row", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{
			Name: "Dashboard Name",
		})
		builder.AddPanelToRow("NonExistent Row", grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Panel Title"),
			},
		}))

		_, err := builder.Build()
		require.Error(t, err)
		require.Contains(t, err.Error(), `unknown row "NonExistent Row"`)
	})

	t.Run("Build succeeds for alerts-only without dashboard name", func(t *testing.T) {
		builder := grafana.NewBuilder(&grafana.BuilderOptions{})
		builder.AddAlert(grafana.NewAlertRule(&grafana.AlertOptions{
			Title: "Alert Title",
		}))

		o, err := builder.Build()
		require.NoError(t, err)
		require.Empty(t, o.Dashboard)
		require.Len(t, o.Alerts, 1)
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
