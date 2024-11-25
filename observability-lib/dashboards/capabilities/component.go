package capabilities

import (
	"fmt"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

type Props struct {
	Name              string              // Name is the name of the dashboard
	MetricsDataSource *grafana.DataSource // MetricsDataSource is the datasource for querying metrics
}

// NewDashboard creates a Capabilities dashboard
func NewDashboard(props *Props) (*grafana.Observability, error) {
	if props.Name == "" {
		return nil, fmt.Errorf("Name is required")
	}

	if props.MetricsDataSource == nil {
		return nil, fmt.Errorf("MetricsDataSource is required")
	} else {
		if props.MetricsDataSource.Name == "" {
			return nil, fmt.Errorf("MetricsDataSource.Name is required")
		}
	}

	builder := grafana.NewBuilder(&grafana.BuilderOptions{
		Name:     props.Name,
		Tags:     []string{"Capabilities"},
		Refresh:  "30s",
		TimeFrom: "now-7d",
		TimeTo:   "now",
	})

	builder.AddVars(vars(props)...)

	builder.AddRow("Common indicators for capabilities")
	builder.AddPanel(capabilitiesCommon(props)...)

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

func capabilitiesCommon(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Execution Time",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `capability_execution_time_ms`,
					Legend: "{{capability}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Runs Count",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `capability_runs_count`,
					Legend: "{{capability}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Runs Fault Count",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `capability_runs_fault_count`,
					Legend: "{{capability}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Runs Invalid Count",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `capability_runs_invalid_count`,
					Legend: "{{capability}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Runs Unauthorized Count",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `capability_runs_unauthorized_count`,
					Legend: "{{capability}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Runs No Resource Count",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `capability_runs_no_resource_count`,
					Legend: "{{capability}}",
				},
			},
		},
	}))

	return panels
}
