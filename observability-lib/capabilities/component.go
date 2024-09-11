package capabilities

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

type Props struct {
	Name              string
	FolderUID         string
	MetricsDataSource *grafana.DataSource
	LogsDataSource    *grafana.DataSource
}

func NewDashboard(options *grafana.DashboardOptions) (*grafana.Dashboard, error) {
	props := &Props{
		Name:              options.Name,
		MetricsDataSource: options.MetricsDataSource,
		LogsDataSource:    options.LogsDataSource,
		FolderUID:         options.FolderUID,
	}

	builder := grafana.NewBuilder(options, &grafana.BuilderOptions{
		Tags:     []string{"Capabilities"},
		Refresh:  "30s",
		TimeFrom: "now-7d",
		TimeTo:   "now",
	})

	builder.AddVars(vars(props)...)

	builder.AddRow("Headlines")
	builder.AddPanel(headlines(props)...)

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
		Query:      `label_values(up, env)`,
		Multi:      false,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Cluster",
			Name:  "cluster",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env"}, cluster)`,
		Multi:      false,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Namespace",
			Name:  "namespace",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env", cluster="$cluster"}, namespace)`,
		Multi:      false,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Job",
			Name:  "job",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env", cluster="$cluster", namespace="$namespace"}, job)`,
		Multi:      false,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Pod",
			Name:  "pod",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env", cluster="$cluster", namespace="$namespace", job="$job"}, pod)`,
		Multi:      false,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Capability",
			Name:  "capability",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env", cluster="$cluster", namespace="$namespace", job="$job"}, pod)`,
		Multi:      false,
	}))

	return variables
}

func headlines(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Execution Time",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:    `capability_cron_execution_time_ms`,
					Legend:  "{{capability}}",
					Instant: true,
				},
			},
		},
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Execution Time",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:    `capability_runs_count`,
					Legend:  "{{capability}}",
					Instant: true,
				},
			},
		},
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Execution Time",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:    `capability_runs_fault_count`,
					Legend:  "{{capability}}",
					Instant: true,
				},
			},
		},
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Execution Time",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:    `capability_runs_invalid_count`,
					Legend:  "{{capability}}",
					Instant: true,
				},
			},
		},
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Execution Time",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:    `capability_runs_unauthorized_count`,
					Legend:  "{{capability}}",
					Instant: true,
				},
			},
		},
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Execution Time",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:    `capability_runs_no_resource_count`,
					Legend:  "{{capability}}",
					Instant: true,
				},
			},
		},
		Orientation: common.VizOrientationHorizontal,
	}))

	return panels
}
