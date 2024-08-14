package corenodecomponents

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

func NewDashboard(options *grafana.DashboardOptions) (*grafana.Dashboard, error) {
	props := &Props{
		Name:              options.Name,
		MetricsDataSource: options.MetricsDataSource,
		LogsDataSource:    options.LogsDataSource,
		PlatformOpts:      PlatformPanelOpts(),
		FolderUID:         options.FolderUID,
	}

	builder := grafana.NewBuilder(options, &grafana.BuilderOptions{
		Tags:     []string{"Core", "Node", "Components"},
		Refresh:  "30s",
		TimeFrom: "now-30m",
		TimeTo:   "now",
	})

	builder.AddVars(vars(props)...)
	builder.AddPanel(panelsGeneralInfo(props)...)

	db, alerts, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return &grafana.Dashboard{
		Dashboard: db,
		Alerts:    alerts,
	}, nil
}

func vars(p *Props) []cog.Builder[dashboard.VariableModel] {
	var variables []cog.Builder[dashboard.VariableModel]

	variables = append(variables, grafana.NewIntervalVariable(&grafana.IntervalVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Interval",
			Name:  "interval",
		},
		Interval: "30s,1m,5m,15m,30m,1h,6h,12h",
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Environment",
			Name:  "env",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up, env)`,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Cluster",
			Name:  "cluster",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env"}, cluster)`,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Blockchain",
			Name:  "blockchain",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env", cluster="$cluster"}, blockchain)`,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Product",
			Name:  "product",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env", cluster="$cluster", blockchain="$blockchain"}, product)`,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Network Type",
			Name:  "network_type",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env", cluster="$cluster", blockchain="$blockchain", product="$product"}, network_type)`,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Component",
			Name:  "component",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env", cluster="$cluster", blockchain="$blockchain", network_type="$network_type"}, component)`,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Service",
			Name:  "service",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{env="$env", cluster="$cluster", blockchain="$blockchain", network_type="$network_type", component="$component"}, service)`,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Service ID",
			Name:  "service_id",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(health{cluster="$cluster", blockchain="$blockchain", network_type="$network_type", component="$component", service="$service"}, service_id)`,
		Multi:      true,
	}))

	return variables
}

func panelsGeneralInfo(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTablePanel(&grafana.TablePanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "List Nodes",
			Span:       24,
			Height:     1,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:    `max(up{` + p.PlatformOpts.LabelQuery + `}) by (env, cluster, blockchain, product, network_type, network, version, team, component, service)`,
					Legend:  "",
					Instant: true,
					Format:  prometheus.PromQueryFormatTable,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Uptime",
			Span:       24,
			Height:     4,
			Decimals:   1,
			Unit:       "percent",
			Query: []grafana.Query{
				{
					Expr:   `100 * up{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: "Team: {{team}} env: {{env}} cluster: {{cluster}} namespace: {{namespace}} job: {{job}} blockchain: {{blockchain}} product: {{product}} networkType: {{network_type}} component: {{component}} service: {{service}}",
				},
			},
			Min: grafana.Pointer[float64](0),
			Max: grafana.Pointer[float64](100),
		},
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Components Health Avg by Service",
			Span:       24,
			Height:     4,
			Decimals:   1,
			Unit:       "percent",
			Query: []grafana.Query{
				{
					Expr:   `100 * avg(avg_over_time(health{` + p.PlatformOpts.LabelQuery + `service_id=~"${service_id}"}[$interval])) by (service_id, version, service, cluster, env)`,
					Legend: "{{service_id}}",
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](80), Color: "orange"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
		},
		GraphMode:   common.BigValueGraphModeLine,
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationVertical,
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Components Health by Service",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "percent",
			Query: []grafana.Query{
				{
					Expr:   `100 * (health{` + p.PlatformOpts.LabelQuery + `service_id=~"${service_id}"})`,
					Legend: "{{service_id}}",
				},
			},
			Min: grafana.Pointer[float64](0),
			Max: grafana.Pointer[float64](100),
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Components Health Avg by Service",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "percent",
			Query: []grafana.Query{
				{
					Expr:   `100 * (avg(avg_over_time(health{` + p.PlatformOpts.LabelQuery + `service_id=~"${service_id}"}[$interval])) by (service_id, version, service, cluster, env))`,
					Legend: "{{service_id}}",
				},
			},
			Min: grafana.Pointer[float64](0),
			Max: grafana.Pointer[float64](100),
		},
	}))

	panels = append(panels, grafana.NewLogPanel(&grafana.LogPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.LogsDataSource.Name,
			Title:      "Logs with severity >= error",
			Span:       24,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `{env="${env}", cluster="${cluster}", product="${product}", network_type="${network_type}", instance=~"${service}"} | json | level=~"(error|panic|fatal|crit)"`,
					Legend: "",
				},
			},
		},
	}))

	return panels
}
