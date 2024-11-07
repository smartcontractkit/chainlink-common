package keystone_workflows

import (
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/expr"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

func NewDashboard(props *Props) (*grafana.Observability, error) {
	if err := validateInput(props); err != nil {
		return nil, err
	}
	props.AlertsTitlePrefix = "[Keystone]"
	props.QueryFilters = `env=~"${env}", cluster=~"${cluster}"`
	props.AlertsTags = map[string]string{
		"team": "keystone",
	}

	builder := grafana.NewBuilder(&grafana.BuilderOptions{
		Name:       props.Name,
		Tags:       []string{"Keystone"},
		Refresh:    "30s",
		TimeFrom:   "now-30m",
		TimeTo:     "now",
		AlertsTags: props.AlertsTags,
	})

	builder.AddVars(vars(props)...)

	builder.AddRow("Engine")
	builder.AddPanel(engine(props)...)

	builder.AddRow("Registry Syncer")
	builder.AddPanel(registrySyncer(props)...)

	if props.SlackChannel != "" && props.SlackWebhookURL != "" {
		builder.AddContactPoint(grafana.NewContactPoint(&grafana.ContactPointOptions{
			Name: "keystone-slack",
			Type: "slack",
			Settings: map[string]interface{}{
				"url":       props.SlackWebhookURL,
				"recipient": props.SlackChannel,
				"username":  "Keystone Alerts",
				"title":     `{{ template "slack.chainlink.title" . }}`,
				"text":      `{{ template "slack.chainlink.text" . }}`,
				"color":     `{{ template "slack.chainlink.color" . }}`,
			},
		}))

		notificationPolicySlackOptions := &grafana.NotificationPolicyOptions{
			Receiver: "keystone-slack",
			GroupBy:  []string{"grafana_folder", "alertname"},
			Continue: grafana.Pointer(true),
		}
		for name, value := range props.AlertsTags {
			notificationPolicySlackOptions.ObjectMatchers = append(notificationPolicySlackOptions.ObjectMatchers, alerting.ObjectMatcher{name, "=", value})
		}
		builder.AddNotificationPolicy(grafana.NewNotificationPolicy(notificationPolicySlackOptions))
	}

	return builder.Build()
}

func vars(p *Props) []cog.Builder[dashboard.VariableModel] {
	var variables []cog.Builder[dashboard.VariableModel]

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Environment",
			Name:  "env",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(platform_engine_workflow_count, env)`,
		Multi:      false,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Cluster",
			Name:  "cluster",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(platform_engine_workflow_count{env="$env"}, cluster)`,
	}))

	return variables
}

func engine(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Workflows Running by Node",
			Description: "",
			Span:        8,
			Height:      8,
			Query: []grafana.Query{
				{
					Expr: `sum(platform_engine_workflow_count) by (container_id)`,
				},
			},
		},
		AlertsOptions: []grafana.AlertOptions{
			{
				Title:       p.AlertsTitlePrefix + "[Engine] No Workflows Running",
				Summary:     "Platform Engine: No workflows are running",
				Description: `{{ index $labels "job" }} number of workflow running is {{ index $values "B" }} in the last 1h`,
				RunbookURL:  "https://github.com/smartcontractkit/chainlink-common/tree/main/observability-lib",
				For:         "1h",
				Tags: map[string]string{
					"severity": "critical",
				},
				NoDataState: alerting.RuleNoDataStateOK,
				Query: []grafana.RuleQuery{
					{
						Expr:       `sum(platform_engine_workflow_count{` + p.AlertsFilters + `}) by (container_id)`,
						RefID:      "A",
						Datasource: p.MetricsDataSource.UID,
					},
				},
				QueryRefCondition: "C",
				// SUM(A) < 1
				Condition: []grafana.ConditionQuery{
					{
						RefID: "B",
						ReduceExpression: &grafana.ReduceExpression{
							Expression: "A",
							Reducer:    expr.TypeReduceReducerSum,
						},
					},
					{
						RefID: "C",
						ThresholdExpression: &grafana.ThresholdExpression{
							Expression: "B",
							ThresholdConditionsOptions: grafana.ThresholdConditionsOption{
								Params: []float64{1},
								Type:   grafana.TypeThresholdTypeLt,
							},
						},
					},
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Workflows Running by status",
			Description: "",
			Span:        8,
			Height:      8,
			Query: []grafana.Query{
				{
					Expr:   `sum(platform_engine_workflow_count{` + p.QueryFilters + `}) by (status)`,
					Legend: "{{ status }}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Register Trigger Failure",
			Description: "",
			Span:        8,
			Height:      8,
			Query: []grafana.Query{
				{
					Expr:   `platform_engine_registertrigger_failures{` + p.QueryFilters + `}`,
					Legend: "",
				},
			},
		},
		AlertsOptions: []grafana.AlertOptions{
			{
				Title:       p.AlertsTitlePrefix + "[Engine] Register Trigger Failure",
				Summary:     "Platform Engine: More than 1 failure over last 15m",
				Description: `{{ index $labels "job" }} registered {{ index $values "A" }} trigger failures in the last 15m`,
				RunbookURL:  "https://github.com/smartcontractkit/chainlink-common/tree/main/observability-lib",
				For:         "15m",
				Tags: map[string]string{
					"severity": "critical",
				},
				NoDataState: alerting.RuleNoDataStateOK,
				Query: []grafana.RuleQuery{
					{
						Expr:       `platform_engine_registertrigger_failures{` + p.AlertsFilters + `}`,
						RefID:      "A",
						Datasource: p.MetricsDataSource.UID,
					},
				},
				QueryRefCondition: "C",
				// SUM(A) > 1
				Condition: []grafana.ConditionQuery{
					{
						RefID: "B",
						ReduceExpression: &grafana.ReduceExpression{
							Expression: "A",
							Reducer:    expr.TypeReduceReducerSum,
						},
					},
					{
						RefID: "C",
						ThresholdExpression: &grafana.ThresholdExpression{
							Expression: "B",
							ThresholdConditionsOptions: grafana.ThresholdConditionsOption{
								Params: []float64{1},
								Type:   grafana.TypeThresholdTypeGt,
							},
						},
					},
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Workflow Step Error",
			Description: "",
			Span:        8,
			Height:      8,
			Query: []grafana.Query{
				{
					Expr:   `platform_engine_workflow_errors_total{` + p.QueryFilters + `}`,
					Legend: "",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Workflow Execution Latency",
			Description: "",
			Span:        8,
			Height:      8,
			Unit:        "ms",
			Query: []grafana.Query{
				{
					Expr:   `platform_engine_workflow_time{` + p.QueryFilters + `}`,
					Legend: "",
				},
			},
		},
		AlertsOptions: []grafana.AlertOptions{
			{
				Title:       p.AlertsTitlePrefix + "[Engine] Workflow Execution Latency",
				Summary:     "Workflow Execution latency is high",
				Description: `{{ index $labels "job" }}/{{ index $labels "workflowID" }} workflow latency is {{ index $values "B" }}ms`,
				RunbookURL:  "https://github.com/smartcontractkit/chainlink-common/tree/main/observability-lib",
				For:         "5m",
				Tags: map[string]string{
					"severity": "critical",
				},
				NoDataState: alerting.RuleNoDataStateOK,
				Query: []grafana.RuleQuery{
					{
						Expr:       `platform_engine_workflow_time{` + p.AlertsFilters + `}`,
						RefID:      "A",
						Datasource: p.MetricsDataSource.UID,
					},
				},
				QueryRefCondition: "C",
				Condition: []grafana.ConditionQuery{
					{
						RefID: "B",
						ReduceExpression: &grafana.ReduceExpression{
							Expression: "A",
							Reducer:    expr.TypeReduceReducerMean,
						},
					},
					{
						RefID: "C",
						ThresholdExpression: &grafana.ThresholdExpression{
							Expression: "B",
							ThresholdConditionsOptions: grafana.ThresholdConditionsOption{
								Params: []float64{900000},
								Type:   grafana.TypeThresholdTypeGt,
							},
						},
					},
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Capability Invocation",
			Description: "",
			Span:        8,
			Height:      8,
			Query: []grafana.Query{
				{
					Expr:   `platform_engine_capabilities_count_total{` + p.QueryFilters + `}`,
					Legend: "",
				},
			},
		},
	}))

	return panels
}

func registrySyncer(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Registry Syncer Failures",
			Description: "",
			Span:        8,
			Height:      8,
			Query: []grafana.Query{
				{
					Expr:   `platform_registrysyncer_sync_failures_total{` + p.QueryFilters + `}`,
					Legend: "",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Registry Syncer Launcher Failures",
			Description: "",
			Span:        8,
			Height:      8,
			Query: []grafana.Query{
				{
					Expr:   `platform_registrysyncer_launch_failures{` + p.QueryFilters + `}`,
					Legend: "",
				},
			},
		},
	}))

	return panels
}
