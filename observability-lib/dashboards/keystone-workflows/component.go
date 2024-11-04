package keystone_workflows

import (
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

func NewDashboard(props *Props) (*grafana.Dashboard, error) {
	if err := platformBuildOpts(props); err != nil {
		return nil, err
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

	builder.AddRow("General")
	builder.AddPanel(general(props)...)

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
		Query:      `label_values(platform.engine.workflows.count, env)`,
		Multi:      false,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Cluster",
			Name:  "cluster",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(platform.engine.workflows.count{env="$env"}, cluster)`,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Workflow Owner",
			Name:  "workflowOwner",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(platform.engine.workflows.count{env="$env", cluster="$cluster"}, workflowOwner)`,
		Multi:      false,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Workflow Name",
			Name:  "workflowName",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(platform.engine.workflows.count{env="$env", cluster="$cluster", workflowOwner="$workflowOwner"}, workflowName)`,
	}))

	return variables
}

func general(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Workflows Running",
			Description: "",
			Span:        8,
			Height:      8,
			Query: []grafana.Query{
				{
					Expr:   `sum(platform.engine.workflows.count{` + p.platformOpts.LabelQuery + `}) by (workflowOwner, workflowName)`,
					Legend: "{{ workflowOwner }} - {{ workflowName }}",
				},
			},
		},
		AlertOptions: &grafana.AlertOptions{
			Summary:     "Keystone: No workflows are running",
			Description: `The number of workflow running is  {{ index $values "A" }}%`,
			RunbookURL:  "https://github.com/smartcontractkit/chainlink-common/tree/main/observability-lib",
			For:         "15m",
			Tags: map[string]string{
				"severity": "critical",
			},
			NoDataState: alerting.RuleNoDataStateOK,
			Query: []grafana.RuleQuery{
				{
					Expr:       `sum(platform.engine.workflows.count{` + p.AlertsFilters + `})`,
					RefID:      "A",
					Datasource: p.MetricsDataSource.UID,
				},
			},
			QueryRefCondition: "B",
			Condition: []grafana.ConditionQuery{
				{
					RefID: "B",
					ThresholdExpression: &grafana.ThresholdExpression{
						Expression: "A",
						ThresholdConditionsOptions: grafana.ThresholdConditionsOption{
							Params: []float64{1},
							Type:   grafana.TypeThresholdTypeLt,
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
					Expr:   `sum(platform.engine.workflows.count{` + p.platformOpts.LabelQuery + `}) by (status)`,
					Legend: "{{ status }}",
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
					Expr:   `sum(platform.engine.workflow.time{` + p.platformOpts.LabelQuery + `}) by (workflowExecutionID)`,
					Legend: "WorkflowExecID: {{workflowExecutionID}}",
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
					Expr:   `platform.engine.workflow.errors{` + p.platformOpts.LabelQuery + `}`,
					Legend: "",
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
					Expr:   `platform.engine.register_trigger.failures{` + p.platformOpts.LabelQuery + `}`,
					Legend: "",
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
					Expr:   `platform.engine.capabilities_invoked.count{` + p.platformOpts.LabelQuery + `}`,
					Legend: "",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Registry Syncer Failures",
			Description: "",
			Span:        8,
			Height:      8,
			Query: []grafana.Query{
				{
					Expr:   `platform.registry_syncer.sync.failures{` + p.platformOpts.LabelQuery + `}`,
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
					Expr:   `platform.registry_syncer.launch.failures{` + p.platformOpts.LabelQuery + `}`,
					Legend: "",
				},
			},
		},
	}))

	return panels
}
