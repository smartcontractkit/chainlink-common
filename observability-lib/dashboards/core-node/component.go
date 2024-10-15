package corenode

import (
	"fmt"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/expr"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

// NewDashboard creates a DON dashboard for the given OCR version
func NewDashboard(props *Props) (*grafana.Dashboard, error) {
	if props.Name == "" {
		return nil, fmt.Errorf("Name is required")
	}

	if props.Platform == "" {
		return nil, fmt.Errorf("Platform is required")
	}

	if props.MetricsDataSource == nil {
		return nil, fmt.Errorf("MetricsDataSource is required")
	} else {
		if props.MetricsDataSource.Name == "" {
			return nil, fmt.Errorf("MetricsDataSource.Name is required")
		}
		if props.MetricsDataSource.UID == "" {
			return nil, fmt.Errorf("MetricsDataSource.UID is required")
		}
	}

	props.platformOpts = platformPanelOpts(props.Platform)
	if props.Tested {
		props.platformOpts.LabelQuery = ""
	}

	builder := grafana.NewBuilder(&grafana.BuilderOptions{
		Name:       props.Name,
		Tags:       []string{"Core", "Node"},
		Refresh:    "30s",
		TimeFrom:   "now-30m",
		TimeTo:     "now",
		AlertsTags: props.AlertsTags,
	})

	if props.SlackChannel != "" && props.SlackWebhookURL != "" {
		builder.AddContactPoint(grafana.NewContactPoint(&grafana.ContactPointOptions{
			Name: "chainlink-slack",
			Type: "slack",
			Settings: map[string]interface{}{
				"url":       props.SlackWebhookURL,
				"recipient": props.SlackChannel,
				"username":  "Chainlink Alerts",
				"title":     `{{ template "slack.chainlink.title" . }}`,
				"text":      `{{ template "slack.chainlink.text" . }}`,
				"color":     `{{ template "slack.chainlink.color" . }}`,
			},
		}))
	}

	notificationPolicyOptions := &grafana.NotificationPolicyOptions{
		Receiver: "chainlink-slack",
		GroupBy:  []string{"grafana_folder", "alertname"},
	}
	for name, value := range props.AlertsTags {
		notificationPolicyOptions.ObjectMatchers = append(notificationPolicyOptions.ObjectMatchers, alerting.ObjectMatcher{name, "=", value})
	}

	builder.AddNotificationPolicy(grafana.NewNotificationPolicy(notificationPolicyOptions))

	builder.AddVars(vars(props)...)

	builder.AddRow("Headlines")
	builder.AddPanel(headlines(props)...)

	builder.AddRow("AppDBConnections")
	builder.AddPanel(appDBConnections(props)...)

	builder.AddRow("SQLQueries")
	builder.AddPanel(sqlQueries(props)...)

	builder.AddRow("HeadTracker")
	builder.AddPanel(headTracker(props)...)

	builder.AddRow("HeadReporter")
	builder.AddPanel(headReporter(props)...)

	builder.AddRow("TxManager")
	builder.AddPanel(txManager(props)...)

	builder.AddRow("LogPoller")
	builder.AddPanel(logPoller(props)...)

	builder.AddRow("Feeds Jobs")
	builder.AddPanel(feedsJobs(props)...)

	builder.AddRow("Mailbox")
	builder.AddPanel(mailbox(props)...)

	builder.AddRow("Logs Counters")
	builder.AddPanel(logsCounters(props)...)

	builder.AddRow("Logs Rate")
	builder.AddPanel(logsRate(props)...)

	builder.AddRow("EvmPoolLifecycle")
	builder.AddPanel(evmPoolLifecycle(props)...)

	builder.AddRow("Node RPC State")
	builder.AddPanel(nodesRPC(props)...)

	builder.AddRow("EVM Pool RPC Node Metrics (App)")
	builder.AddPanel(evmNodeRPC(props)...)

	builder.AddRow("EVM Pool RPC Node Latencies (App)")
	builder.AddPanel(evmPoolRPCNodeLatencies(props)...)

	builder.AddRow("Block History Estimator")
	builder.AddPanel(evmBlockHistoryEstimator(props)...)

	builder.AddRow("Pipeline Metrics (Runner)")
	builder.AddPanel(pipelines(props)...)

	builder.AddRow("HTTP API")
	builder.AddPanel(httpAPI(props)...)

	builder.AddRow("PromHTTP")
	builder.AddPanel(promHTTP(props)...)

	builder.AddRow("Go Metrics")
	builder.AddPanel(goMetrics(props)...)

	return builder.Build()
}

func vars(p *Props) []cog.Builder[dashboard.VariableModel] {
	var variables []cog.Builder[dashboard.VariableModel]

	if p.platformOpts.Platform == "kubernetes" {
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
				Label: "Namespace",
				Name:  "namespace",
			},
			Datasource: p.MetricsDataSource.Name,
			Query:      `label_values(up{env="$env", cluster="$cluster"}, namespace)`,
		}))

		variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Label: "Blockchain",
				Name:  "blockchain",
			},
			Datasource: p.MetricsDataSource.Name,
			Query:      `label_values(up{env="$env", cluster="$cluster", namespace="$namespace"}, blockchain)`,
		}))

		variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Label: "Product",
				Name:  "product",
			},
			Datasource: p.MetricsDataSource.Name,
			Query:      `label_values(up{env="$env", cluster="$cluster", namespace="$namespace", blockchain="$blockchain"}, product)`,
		}))

		variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Label: "Network Type",
				Name:  "network_type",
			},
			Datasource: p.MetricsDataSource.Name,
			Query:      `label_values(up{env="$env", cluster="$cluster", namespace="$namespace", blockchain="$blockchain", product="$product"}, network_type)`,
		}))

		variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Label: "Job",
				Name:  "job",
			},
			Datasource: p.MetricsDataSource.Name,
			Query:      `label_values(up{env="$env", cluster="$cluster", namespace="$namespace", blockchain="$blockchain", product="$product", network_type="$network_type"}, job)`,
			Multi:      true,
		}))

		variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Label: "Pod",
				Name:  "pod",
			},
			Datasource: p.MetricsDataSource.Name,
			Query:      `label_values(up{env="$env", cluster="$cluster", namespace="$namespace", job="$job"}, pod)`,
			Multi:      true,
			IncludeAll: true,
		}))
	} else {
		variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Label: "Instance",
				Name:  "instance",
			},
			Datasource: p.MetricsDataSource.Name,
			Query:      fmt.Sprintf("label_values(%s)", p.platformOpts.LabelFilter),
			Multi:      true,
			IncludeAll: true,
		}))
	}

	return variables
}

func headlines(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "App Version",
			Description: "app version with commit and branch links",
			Span:        12,
			Height:      4,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:    `version{` + p.platformOpts.LabelQuery + `}`,
					Legend:  "Version: {{version}} https://github.com/smartcontractkit/chainlink/commit/{{commit}} https://github.com/smartcontractkit/chainlink/tree/release/{{version}}",
					Instant: true,
				},
			},
		},
		ColorMode:   common.BigValueColorModeNone,
		TextMode:    common.BigValueTextModeName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Uptime",
			Description: "instance uptime",
			Span:        12,
			Height:      4,
			Decimals:    2,
			Unit:        "s",
			Query: []grafana.Query{
				{
					Expr:   `uptime_seconds{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
		ColorMode:   common.BigValueColorModeNone,
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "ETH Balance Summary",
			Span:       12,
			Height:     4,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:    `sum(eth_balance{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `, account)`,
					Legend:  `{{` + p.platformOpts.LabelFilter + `}} - {{account}}`,
					Instant: true,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
		},
		GraphMode:   common.BigValueGraphModeLine,
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Solana Balance Summary",
			Span:       12,
			Height:     4,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:    `sum(solana_balance{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `, account)`,
					Legend:  `{{` + p.platformOpts.LabelFilter + `}} - {{account}}`,
					Instant: true,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
		},
		GraphMode:   common.BigValueGraphModeLine,
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Health Avg by Service over 15m",
			Span:       16,
			Height:     6,
			Decimals:   1,
			Unit:       "percent",
			Query: []grafana.Query{
				{
					Expr:   `100 * (avg(avg_over_time(health{` + p.platformOpts.LabelQuery + `}[15m])) by (` + p.platformOpts.LabelFilter + `, service_id, version, service, cluster, env))`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{service_id}}`,
				},
			},
			Min: grafana.Pointer[float64](0),
			Max: grafana.Pointer[float64](100),
			AlertOptions: &grafana.AlertOptions{
				Summary:     `Uptime less than 90% over last 15 minutes on one component in a Node`,
				Description: `Component {{ index $labels "service_id" }} uptime in the last 15m is {{ index $values "A" }}%`,
				RunbookURL:  "https://github.com/smartcontractkit/chainlink-common/tree/main/observability-lib",
				For:         "15m",
				Tags: map[string]string{
					"severity": "warning",
				},
				Query: []grafana.RuleQuery{
					{
						Expr:       `health{` + p.AlertsFilters + `}`,
						RefID:      "A",
						Datasource: p.MetricsDataSource.UID,
					},
				},
				QueryRefCondition: "D",
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
						MathExpression: &grafana.MathExpression{
							Expression: "$B * 100",
						},
					},
					{
						RefID: "D",
						ThresholdExpression: &grafana.ThresholdExpression{
							Expression: "C",
							ThresholdConditionsOptions: []grafana.ThresholdConditionsOption{
								{
									Params: []float64{90, 0},
									Type:   expr.TypeThresholdTypeLt,
								},
							},
						},
					},
				},
			},
		},
		LegendOptions: &grafana.LegendOptions{
			DisplayMode: common.LegendDisplayModeList,
			Placement:   common.LegendPlacementRight,
		},
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Health Avg by Service over 15m with health < 90%",
			Description: "Only displays services with health average < 90%",
			Span:        8,
			Height:      6,
			Decimals:    1,
			Unit:        "percent",
			Query: []grafana.Query{
				{
					Expr:   `100 * avg(avg_over_time(health{` + p.platformOpts.LabelQuery + `}[15m])) by (` + p.platformOpts.LabelFilter + `, service_id, version, service, cluster, env) < 90`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{service_id}}`,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "green"},
					{Value: grafana.Pointer[float64](1), Color: "red"},
					{Value: grafana.Pointer[float64](80), Color: "orange"},
					{Value: grafana.Pointer[float64](99), Color: "green"},
				},
			},
			NoValue: "All services healthy",
		},
		GraphMode:   common.BigValueGraphModeLine,
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "ETH Balance",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:   `sum(eth_balance{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `, account)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{account}}`,
				},
			},
			AlertOptions: &grafana.AlertOptions{
				Summary:     `ETH Balance is lower than threshold`,
				Description: `ETH Balance critically low at {{ index $values "A" }} on {{ index $labels "` + p.platformOpts.LabelFilter + `" }}`,
				RunbookURL:  "https://github.com/smartcontractkit/chainlink-common/tree/main/observability-lib",
				For:         "15m",
				NoDataState: alerting.RuleNoDataStateOK,
				Tags: map[string]string{
					"severity": "critical",
				},
				Query: []grafana.RuleQuery{
					{
						Expr:       `eth_balance{` + p.AlertsFilters + `}`,
						Instant:    true,
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
							ThresholdConditionsOptions: []grafana.ThresholdConditionsOption{
								{
									Params: []float64{1, 0},
									Type:   expr.TypeThresholdTypeLt,
								},
							},
						},
					},
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "SOL Balance",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:   `sum(solana_balance{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `, account)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{account}}`,
				},
			},
			AlertOptions: &grafana.AlertOptions{
				Summary:     `Solana Balance is lower than threshold`,
				Description: `Solana Balance critically low at {{ index $values "A" }} on {{ index $labels "` + p.platformOpts.LabelFilter + `" }}`,
				RunbookURL:  "https://github.com/smartcontractkit/chainlink-common/tree/main/observability-lib",
				For:         "15m",
				NoDataState: alerting.RuleNoDataStateOK,
				Tags: map[string]string{
					"severity": "critical",
				},
				Query: []grafana.RuleQuery{
					{
						Expr:       `solana_balance{` + p.AlertsFilters + `}`,
						Instant:    true,
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
							ThresholdConditionsOptions: []grafana.ThresholdConditionsOption{
								{
									Params: []float64{1, 0},
									Type:   expr.TypeThresholdTypeLt,
								},
							},
						},
					},
				},
			},
		},
	}))

	if p.platformOpts.Platform == "kubernetes" {
		panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      "CPU Utilisation (from requests)",
				Span:       6,
				Height:     4,
				Decimals:   1,
				Unit:       "percent",
				Query: []grafana.Query{
					{
						Expr:    `100 * sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{cluster="$cluster", namespace="$namespace", pod="$pod"}) by (container) / sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_requests{cluster="$cluster", namespace="$namespace", pod="$pod"}) by (container)`,
						Legend:  `{{pod}}`,
						Instant: true,
					},
				},
			},
		}))

		panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      "CPU Utilisation (from limits)",
				Span:       6,
				Height:     4,
				Decimals:   1,
				Unit:       "percent",
				Query: []grafana.Query{
					{
						Expr:    `100 * sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{cluster="$cluster", namespace="$namespace", pod="$pod"}) by (container) / sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits{cluster="$cluster", namespace="$namespace", pod="$pod"}) by (container)`,
						Legend:  `{{pod}}`,
						Instant: true,
					},
				},
			},
		}))

		panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      "Memory Utilisation (from requests)",
				Span:       6,
				Height:     4,
				Decimals:   1,
				Unit:       "percent",
				Query: []grafana.Query{
					{
						Expr:    `100 * sum(container_memory_working_set_bytes{job="kubelet", metrics_path="/metrics/cadvisor", cluster="$cluster", namespace="$namespace", pod="$pod", image!=""}) by (container) / sum(cluster:namespace:pod_memory:active:kube_pod_container_resource_requests{cluster="$cluster", namespace="$namespace", pod="$pod"}) by (container)`,
						Legend:  `{{pod}}`,
						Instant: true,
					},
				},
			},
		}))

		panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      "Memory Utilisation (from limits)",
				Span:       6,
				Height:     4,
				Decimals:   1,
				Unit:       "percent",
				Query: []grafana.Query{
					{
						Expr:    `100 * sum(container_memory_working_set_bytes{job="kubelet", metrics_path="/metrics/cadvisor", cluster="$cluster", namespace="$namespace", pod="$pod", container!="", image!=""}) by (container) / sum(cluster:namespace:pod_memory:active:kube_pod_container_resource_limits{cluster="$cluster", namespace="$namespace", pod="$pod"}) by (container)`,
						Legend:  `{{pod}}`,
						Instant: true,
					},
				},
			},
		}))

		panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      "CPU Usage",
				Span:       12,
				Height:     8,
				Decimals:   3,
				Query: []grafana.Query{
					{
						Expr:   `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate{pod=~"$pod", namespace=~"${namespace}"}) by (pod)`,
						Legend: "{{pod}}",
					},
					{
						Expr:   `sum(kube_pod_container_resource_requests{job="kube-state-metrics", cluster="$cluster", namespace="$namespace", pod="$pod", resource="cpu"})`,
						Legend: "Requests",
					},
					{
						Expr:   `sum(kube_pod_container_resource_limits{job="kube-state-metrics", cluster="$cluster", namespace="$namespace", pod="$pod", resource="cpu"})`,
						Legend: "Limits",
					},
				},
			},
			ScaleDistribution: common.ScaleDistributionLog,
		}))

		panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      "Memory Usage",
				Span:       12,
				Height:     8,
				Unit:       "bytes",
				Query: []grafana.Query{
					{
						Expr:   `sum(container_memory_rss{pod=~"$pod", namespace=~"${namespace}", container!=""}) by (pod)`,
						Legend: "{{pod}}",
					},
					{
						Expr:   `sum(kube_pod_container_resource_requests{job="kube-state-metrics", cluster="$cluster", namespace="$namespace", pod="$pod", resource="memory"})`,
						Legend: "Requests",
					},
					{
						Expr:   `sum(kube_pod_container_resource_limits{job="kube-state-metrics", cluster="$cluster", namespace="$namespace", pod="$pod", resource="memory"})`,
						Legend: "Limits",
					},
				},
			},
			ScaleDistribution: common.ScaleDistributionLog,
		}))
	}

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Open File Descriptors",
			Span:       6,
			Height:     4,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `process_open_fds{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
		GraphMode: common.BigValueGraphModeArea,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Go Version",
			Span:       4,
			Height:     4,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:    `go_info{` + p.platformOpts.LabelQuery + `}`,
					Legend:  "{{exported_version}}",
					Instant: true,
				},
			},
		},
		ColorMode: common.BigValueColorModeNone,
		TextMode:  common.BigValueTextModeName,
	}))

	return panels
}

func appDBConnections(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "DB Connections",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "Conn",
			Query: []grafana.Query{
				{
					Expr:   `sum(db_conns_max{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - Max`,
				},
				{
					Expr:   `sum(db_conns_open{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - Open`,
				},
				{
					Expr:   `sum(db_conns_used{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - Used`,
				},
				{
					Expr:   `sum(db_conns_wait{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - Wait`,
				},
			},
		},
		LegendOptions: &grafana.LegendOptions{
			DisplayMode: common.LegendDisplayModeList,
			Placement:   common.LegendPlacementRight,
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "DB Wait Count",
			Span:       12,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(db_wait_count{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "DB Wait Time",
			Span:       12,
			Height:     6,
			Unit:       "Sec",
			Query: []grafana.Query{
				{
					Expr:   `sum(db_wait_time_seconds{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}}`,
				},
			},
		},
	}))

	return panels
}

func sqlQueries(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "SQL Query Timeout Percent",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "percent",
			Query: []grafana.Query{
				{
					Expr:   `histogram_quantile(0.9, sum(rate(sql_query_timeout_percent_bucket{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (le))`,
					Legend: "p90",
				},
				{
					Expr:   `histogram_quantile(0.95, sum(rate(sql_query_timeout_percent_bucket{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (le))`,
					Legend: "p95",
				},
				{
					Expr:   `histogram_quantile(0.99, sum(rate(sql_query_timeout_percent_bucket{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (le))`,
					Legend: "p99",
				},
			},
		},
		LegendOptions: &grafana.LegendOptions{
			DisplayMode: common.LegendDisplayModeList,
			Placement:   common.LegendPlacementRight,
		},
	}))

	return panels
}

func headTracker(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Head Tracker Current Head",
			Span:       18,
			Height:     6,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `sum(head_tracker_current_head{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Head Tracker Current Head Summary",
			Span:       6,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:    `head_tracker_current_head{` + p.platformOpts.LabelQuery + `}`,
					Legend:  `{{` + p.platformOpts.LegendString + `}}`,
					Instant: true,
				},
			},
		},
		ColorMode: common.BigValueColorModeNone,
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Head Tracker Heads Received Rate",
			Span:       24,
			Height:     6,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `rate(head_tracker_heads_received{` + p.platformOpts.LabelQuery + `}[1m])`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}}`,
				},
			},
			AlertOptions: &grafana.AlertOptions{
				Summary:     `No Headers Received`,
				Description: `{{ index $labels "` + p.platformOpts.LabelFilter + `" }} on ChainID {{ index $labels "ChainID" }} has received {{ index $values "A" }} heads over 10 minutes.`,
				RunbookURL:  "https://github.com/smartcontractkit/chainlink-common/tree/main/observability-lib",
				For:         "10m",
				NoDataState: alerting.RuleNoDataStateOK,
				Tags: map[string]string{
					"severity": "critical",
				},
				Query: []grafana.RuleQuery{
					{
						Expr:       `increase(head_tracker_heads_received{` + p.AlertsFilters + `}[10m])`,
						Instant:    true,
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
							ThresholdConditionsOptions: []grafana.ThresholdConditionsOption{
								{
									Params: []float64{1, 0},
									Type:   expr.TypeThresholdTypeLt,
								},
							},
						},
					},
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Head Tracker Very Old Head",
			Span:       12,
			Height:     6,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `head_tracker_very_old_head{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Head Tracker Connection Errors Rate",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `rate(head_tracker_connection_errors{` + p.platformOpts.LabelQuery + `}[1m])`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	return panels
}

func headReporter(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Unconfirmed Transactions",
			Span:       8,
			Height:     6,
			Unit:       "Tx",
			Query: []grafana.Query{
				{
					Expr:   `sum(unconfirmed_transactions{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Unconfirmed TX Age",
			Span:       8,
			Height:     6,
			Unit:       "s",
			Query: []grafana.Query{
				{
					Expr:   `sum(max_unconfirmed_tx_age{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Unconfirmed TX Blocks",
			Span:       8,
			Height:     6,
			Unit:       "Blocks",
			Query: []grafana.Query{
				{
					Expr:   `sum(max_unconfirmed_blocks{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	return panels
}

func txManager(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Confirmed",
			Span:       6,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(tx_manager_num_confirmed_transactions{` + p.platformOpts.LabelQuery + `}) by (blockchain, chainID, ` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{blockchain}} - {{chainID}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Successful",
			Span:       6,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(tx_manager_num_successful_transactions{` + p.platformOpts.LabelQuery + `}) by (blockchain, chainID, ` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{blockchain}} - {{chainID}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Reverted",
			Span:       6,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(tx_manager_num_tx_reverted{` + p.platformOpts.LabelQuery + `}) by (blockchain, chainID, ` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{blockchain}} - {{chainID}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Gas Bumps",
			Span:       6,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(tx_manager_num_gas_bumps{` + p.platformOpts.LabelQuery + `}) by (blockchain, chainID, ` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{blockchain}} - {{chainID}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Forwarded",
			Span:       6,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(tx_manager_fwd_tx_count{` + p.platformOpts.LabelQuery + `}) by (blockchain, chainID, ` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{blockchain}} - {{chainID}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Attempts",
			Span:       6,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(tx_manager_tx_attempt_count{` + p.platformOpts.LabelQuery + `}) by (blockchain, chainID, ` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{blockchain}} - {{chainID}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Gas Bump Exceeds Limit",
			Span:       6,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(tx_manager_gas_bump_exceeds_limit{` + p.platformOpts.LabelQuery + `}) by (blockchain, chainID, ` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{blockchain}} - {{chainID}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "TX Manager Time Until Broadcast",
			Description: "The amount of time elapsed from when a transaction is enqueued to until it is broadcast",
			Span:        6,
			Height:      6,
			Decimals:    1,
			Unit:        "ms",
			Query: []grafana.Query{
				{
					Expr:   `histogram_quantile(0.9, sum(rate(tx_manager_time_until_tx_broadcast_bucket{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (le, ` + p.platformOpts.LabelFilter + `, blockchain, chainID)) / 1e6`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{blockchain}} - {{chainID}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "TX Manager Time Until Confirmed",
			Description: "The amount of time elapsed from a transaction being broadcast to being included in a block",
			Span:        6,
			Height:      6,
			Decimals:    1,
			Unit:        "ms",
			Query: []grafana.Query{
				{
					Expr:   `histogram_quantile(0.9, sum(rate(tx_manager_time_until_tx_confirmed_bucket{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (le, ` + p.platformOpts.LabelFilter + `, blockchain, chainID)) / 1e6`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{blockchain}} - {{chainID}}`,
				},
			},
		},
	}))

	return panels
}

func logPoller(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Goroutines per ChainId",
			Description: "goroutines per chainId",
			Span:        12,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `count(log_poller_query_duration_sum{` + p.platformOpts.LabelQuery + `}) by (evmChainID)`,
					Legend: "chainId: {{evmChainID}}",
				},
			},
		},
		ColorMode:   common.BigValueColorModeValue,
		GraphMode:   common.BigValueGraphModeLine,
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "RPS",
			Description: "requests per second",
			Span:        12,
			Height:      6,
			Decimals:    2,
			Unit:        "reqps",
			Query: []grafana.Query{
				{
					Expr:   `avg by (query, ` + p.platformOpts.LabelFilter + `) (sum by (query, job) (rate(log_poller_query_duration_count{` + p.platformOpts.LabelQuery + `}[$__rate_interval])))`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{query}}`,
				},
				{
					Expr:   `avg (sum by(` + p.platformOpts.LabelFilter + `) (rate(log_poller_query_duration_count{` + p.platformOpts.LabelQuery + `}[$__rate_interval])))`,
					Legend: "Total",
				},
			},
		},
		LegendOptions: &grafana.LegendOptions{
			DisplayMode: common.LegendDisplayModeList,
			Placement:   common.LegendPlacementRight,
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "RPS by Type",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Unit:       "reqps",
			Query: []grafana.Query{
				{
					Expr:   `avg by (` + p.platformOpts.LabelFilter + `, type) (sum by (type, ` + p.platformOpts.LabelFilter + `) (rate(log_poller_query_duration_count{` + p.platformOpts.LabelQuery + `}[$__rate_interval])))`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{type}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Avg number of logs returned",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:   `avg by (` + p.platformOpts.LabelFilter + `, query) (log_poller_query_dataset_size{` + p.platformOpts.LabelQuery + `})`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{query}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Max number of logs returned",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:   `max by (` + p.platformOpts.LabelFilter + `, query) (log_poller_query_dataset_size{` + p.platformOpts.LabelQuery + `})`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{query}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Logs returned by chain",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:   `max by (evmChainID) (log_poller_query_dataset_size{` + p.platformOpts.LabelQuery + `})`,
					Legend: "{{evmChainID}}",
				},
			},
		},
	}))

	quantiles := []string{"0.5", "0.9", "0.99"}
	for _, quantile := range quantiles {
		panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      `Queries duration by query ` + quantile + ` quantile`,
				Span:       24,
				Height:     6,
				Decimals:   2,
				Unit:       "ms",
				Query: []grafana.Query{
					{
						Expr:   `histogram_quantile(` + quantile + `, sum(rate(log_poller_query_duration_bucket{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (le, ` + p.platformOpts.LabelFilter + `, query)) / 1e6`,
						Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{query}}`,
					},
				},
			},
			LegendOptions: &grafana.LegendOptions{
				DisplayMode: common.LegendDisplayModeList,
				Placement:   common.LegendPlacementRight,
			},
		}))
	}

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Number of logs inserted",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:   `avg by (evmChainID) (log_poller_logs_inserted{` + p.platformOpts.LabelQuery + `})`,
					Legend: "{{evmChainID}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Logs insertion rate",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:   `avg by (evmChainID) (rate(log_poller_logs_inserted{` + p.platformOpts.LabelQuery + `}[$__rate_interval]))`,
					Legend: "{{evmChainID}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Number of blocks inserted",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:   `avg by (evmChainID) (log_poller_blocks_inserted{` + p.platformOpts.LabelQuery + `})`,
					Legend: "{{evmChainID}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Blocks insertion rate",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:   `avg by (evmChainID) (rate(log_poller_blocks_inserted{` + p.platformOpts.LabelQuery + `}[$__rate_interval]))`,
					Legend: "{{evmChainID}}",
				},
			},
		},
	}))

	return panels
}

func feedsJobs(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Feeds Job Proposal Requests",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `sum(feeds_job_proposal_requests{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Feeds Job Proposal Count",
			Description: "",
			Span:        12,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum(feeds_job_proposal_count{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	return panels
}

func mailbox(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Mailbox Load Percent",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "percent",
			Query: []grafana.Query{
				{
					Expr:   `sum(mailbox_load_percent{` + p.platformOpts.LabelQuery + `}) by (capacity, name, ` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - Capacity: {{capacity}} - {{name}}`,
				},
			},
		},
		LegendOptions: &grafana.LegendOptions{
			DisplayMode: common.LegendDisplayModeList,
			Placement:   common.LegendPlacementRight,
		},
	}))

	return panels
}

func logsCounters(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	logStatuses := []string{"panic", "fatal", "critical", "warn", "error"}
	for _, status := range logStatuses {
		panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      "Logs Counter - " + status,
				Span:       8,
				Height:     6,
				Query: []grafana.Query{
					{
						Expr:   `sum(log_` + status + `_count{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
						Legend: `{{` + p.platformOpts.LabelFilter + `}} - ` + status,
					},
				},
			},
		}))
	}

	return panels
}

func logsRate(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	logStatuses := []string{"panic", "fatal", "critical", "warn", "error"}
	for _, status := range logStatuses {
		panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      "Logs Rate - " + status,
				Span:       8,
				Height:     6,
				Query: []grafana.Query{
					{
						Expr:   `sum(rate(log_` + status + `_count{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LabelFilter + `)`,
						Legend: `{{` + p.platformOpts.LabelFilter + `}} - error`,
					},
				},
			},
		}))
	}

	return panels
}

func evmPoolLifecycle(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool Highest Seen Block",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `evm_pool_rpc_node_highest_seen_block{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool Num Seen Blocks",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `evm_pool_rpc_node_num_seen_blocks{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool Node Polls Total",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `evm_pool_rpc_node_polls_total{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool Node Polls Failed",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `evm_pool_rpc_node_polls_failed{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool Node Polls Success",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `evm_pool_rpc_node_polls_success{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	return panels
}

func nodesRPC(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	nodeRPCStates := []string{"Alive", "Closed", "Dialed", "InvalidChainID", "OutOfSync", "Undialed", "Unreachable", "Unusable"}
	for _, state := range nodeRPCStates {
		panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      "Node RPC " + state,
				Span:       6,
				Height:     6,
				Decimals:   1,
				Query: []grafana.Query{
					{
						Expr:   `sum(multi_node_states{` + p.platformOpts.LabelQuery + `state="` + state + `"}) by (` + p.platformOpts.LabelFilter + `, chainId)`,
						Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{chainId}}`,
					},
				},
			},
			TextMode:    common.BigValueTextModeValueAndName,
			Orientation: common.VizOrientationHorizontal,
		}))
	}

	return panels
}

func evmNodeRPC(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool RPC Node Calls Success Rate",
			Span:       24,
			Height:     6,
			Unit:       "percentunit",
			Max:        grafana.Pointer[float64](1),
			Query: []grafana.Query{
				{
					Expr:   `sum(increase(evm_pool_rpc_node_calls_success{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LabelFilter + `, evmChainID, nodeName) / sum(increase(evm_pool_rpc_node_calls_total{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LabelFilter + `, evmChainID, nodeName)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{nodeName}}`,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.8), Color: "orange"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
		},
		LegendOptions: &grafana.LegendOptions{
			DisplayMode: common.LegendDisplayModeList,
			Placement:   common.LegendPlacementRight,
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool RPC Node Dials Failure Rate",
			Span:       24,
			Height:     6,
			Unit:       "percentunit",
			Max:        grafana.Pointer[float64](1),
			Query: []grafana.Query{
				{
					Expr:   `sum(increase(evm_pool_rpc_node_dials_failed{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LabelFilter + `, evmChainID, nodeName) / sum(increase(evm_pool_rpc_node_calls_total{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LabelFilter + `, evmChainID, nodeName)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{evmChainID}} - {{nodeName}}`,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.8), Color: "orange"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
		},
		LegendOptions: &grafana.LegendOptions{
			DisplayMode: common.LegendDisplayModeList,
			Placement:   common.LegendPlacementRight,
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool RPC Node Transitions",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_alive{` + p.platformOpts.LabelQuery + `}`,
					Legend: "Alive",
				},
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_in_sync{` + p.platformOpts.LabelQuery + `}`,
					Legend: "InSync",
				},
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_out_of_sync{` + p.platformOpts.LabelQuery + `}`,
					Legend: "OutOfSync",
				},
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_unreachable{` + p.platformOpts.LabelQuery + `}`,
					Legend: "UnReachable",
				},
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_invalid_chain_id{` + p.platformOpts.LabelQuery + `}`,
					Legend: "InvalidChainID",
				},
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_unusable{` + p.platformOpts.LabelQuery + `}`,
					Legend: "TransitionToUnusable",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool RPC Node States",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `evm_pool_rpc_node_states{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - {{evmChainID}} - {{state}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool RPC Node Verifies Success Rate",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `sum(increase(evm_pool_rpc_node_verifies_success{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `, evmChainID, nodeName) / sum(increase(evm_pool_rpc_node_verifies{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `, evmChainID, nodeName) * 100`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - {{evmChainID}} - {{nodeName}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool RPC Node Verifies Failure Rate",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `sum(increase(evm_pool_rpc_node_verifies_failed{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `, evmChainID, nodeName) / sum(increase(evm_pool_rpc_node_verifies{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `, evmChainID, nodeName) * 100`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - {{evmChainID}} - {{nodeName}}`,
				},
			},
		},
	}))

	return panels
}

func evmPoolRPCNodeLatencies(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	quantiles := []string{"0.90", "0.95", "0.99"}
	for _, quantile := range quantiles {
		panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Datasource: p.MetricsDataSource.Name,
				Title:      `EVM Pool RPC Node Calls Latency ` + quantile + ` quantile`,
				Span:       24,
				Height:     6,
				Decimals:   1,
				Unit:       "ms",
				Query: []grafana.Query{
					{
						Expr:   `histogram_quantile(` + quantile + `, sum(rate(evm_pool_rpc_node_rpc_call_time_bucket{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LabelFilter + `, le, rpcCallName)) / 1e6`,
						Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{rpcCallName}}`,
					},
				},
			},
			LegendOptions: &grafana.LegendOptions{
				DisplayMode: common.LegendDisplayModeList,
				Placement:   common.LegendPlacementRight,
			},
		}))
	}

	return panels
}

func evmBlockHistoryEstimator(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Gas Updater All Gas Price Percentiles",
			Description: "Gas price at given percentile",
			Span:        24,
			Height:      6,
			Unit:        "gwei",
			Query: []grafana.Query{
				{
					Expr:   `sum(gas_updater_all_gas_price_percentiles{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `, evmChainID, percentile)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{evmChainID}} - {{percentile}}`,
				},
			},
		},
		LegendOptions: &grafana.LegendOptions{
			DisplayMode: common.LegendDisplayModeList,
			Placement:   common.LegendPlacementRight,
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Gas Updater All Tip Cap Percentiles",
			Description: "Tip cap at given percentile",
			Span:        24,
			Height:      6,
			Query: []grafana.Query{
				{
					Expr:   `sum(gas_updater_all_tip_cap_percentiles{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `, evmChainID, percentile)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}} - {{evmChainID}} - {{percentile}}`,
				},
			},
		},
		LegendOptions: &grafana.LegendOptions{
			DisplayMode: common.LegendDisplayModeList,
			Placement:   common.LegendPlacementRight,
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Gas Updater Set Gas Price",
			Span:       12,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(gas_updater_set_gas_price{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Gas Updater Set Tip Cap",
			Span:       12,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(gas_updater_set_tip_cap{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Gas Updater Current Base Fee",
			Span:       12,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(gas_updater_current_base_fee{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Block History Estimator Connectivity Failure Count",
			Span:       12,
			Height:     6,
			Query: []grafana.Query{
				{
					Expr:   `sum(block_history_estimator_connectivity_failure_count{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LabelFilter + `)`,
					Legend: `{{` + p.platformOpts.LabelFilter + `}}`,
				},
			},
		},
	}))

	return panels
}

func pipelines(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Pipeline Task Execution Time",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "s",
			Query: []grafana.Query{
				{
					Expr:   `pipeline_task_execution_time{` + p.platformOpts.LabelQuery + `} / 1e6`,
					Legend: `{{` + p.platformOpts.LegendString + `}} JobID: {{job_id}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Pipeline Run Errors",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `pipeline_run_errors{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} JobID: {{job_id}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Pipeline Run Total Time to Completion",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "s",
			Query: []grafana.Query{
				{
					Expr:   `pipeline_run_total_time_to_completion{` + p.platformOpts.LabelQuery + `} / 1e6`,
					Legend: `{{` + p.platformOpts.LegendString + `}} JobID: {{job_id}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Pipeline Tasks Total Finished",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `pipeline_tasks_total_finished{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} JobID: {{job_id}}`,
				},
			},
		},
	}))

	return panels
}

func httpAPI(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Request Duration p95",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "s",
			Query: []grafana.Query{
				{
					Expr:   `histogram_quantile(0.95, sum(rate(service_gonic_request_duration_bucket{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `, le, path, method))`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - {{method}} - {{path}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Request Total Rate over interval",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `sum(rate(service_gonic_requests_total{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `, path, method, code)`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - {{method}} - {{path}} - {{code}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Average Request Size",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "bytes",
			Query: []grafana.Query{
				{
					Expr:   `avg(rate(service_gonic_request_size_bytes_sum{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `)/avg(rate(service_gonic_request_size_bytes_count{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `)`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Response Size",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "bytes",
			Query: []grafana.Query{
				{
					Expr:   `avg(rate(service_gonic_response_size_bytes_sum{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `)/avg(rate(service_gonic_response_size_bytes_count{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `)`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	return panels
}

func promHTTP(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "HTTP rate by return code",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `sum(rate(promhttp_metric_handler_requests_total{` + p.platformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.platformOpts.LegendString + `, code)`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - {{code}}`,
				},
			},
		},
	}))

	return panels
}

func goMetrics(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Threads",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `sum(go_threads{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LegendString + `)`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Heap Allocations Stats",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "bytes",
			Query: []grafana.Query{
				{
					Expr:   `sum(go_memstats_heap_alloc_bytes{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LegendString + `)`,
					Legend: "",
				},
			},
		},
		ColorMode:   common.BigValueColorModeNone,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Heap allocations Graph",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "bytes",
			Query: []grafana.Query{
				{
					Expr:   `sum(go_memstats_heap_alloc_bytes{` + p.platformOpts.LabelQuery + `}) by (` + p.platformOpts.LegendString + `)`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Heap Usage",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "bytes",
			Query: []grafana.Query{
				{
					Expr:   `go_memstats_heap_alloc_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Alloc`,
				},
				{
					Expr:   `go_memstats_heap_sys_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Sys`,
				},
				{
					Expr:   `go_memstats_heap_idle_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Idle`,
				},
				{
					Expr:   `go_memstats_heap_inuse_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - InUse`,
				},
				{
					Expr:   `go_memstats_heap_released_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Released`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Memory in Off-Heap",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "bytes",
			Query: []grafana.Query{
				{
					Expr:   `go_memstats_mspan_inuse_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Total InUse`,
				},
				{
					Expr:   `go_memstats_mspan_sys_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Total Sys`,
				},
				{
					Expr:   `go_memstats_mcache_inuse_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Cache InUse`,
				},
				{
					Expr:   `go_memstats_mcache_sys_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Cache Sys`,
				},
				{
					Expr:   `go_memstats_buck_hash_sys_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Hash Sys`,
				},
				{
					Expr:   `go_memstats_gc_sys_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - GC Sys`,
				},
				{
					Expr:   `go_memstats_other_sys_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - bytes of memory are used for other runtime allocations`,
				},
				{
					Expr:   `go_memstats_next_gc_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Next GC`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Memory in Stack",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "bytes",
			Query: []grafana.Query{
				{
					Expr:   `go_memstats_stack_inuse_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - InUse`,
				},
				{
					Expr:   `go_memstats_stack_sys_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}} - Sys`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Total Used Memory",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "bytes",
			Query: []grafana.Query{
				{
					Expr:   `go_memstats_sys_bytes{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Number of Live Objects",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `go_memstats_mallocs_total{` + p.platformOpts.LabelQuery + `} - go_memstats_frees_total{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Rate of Objects Allocated",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `rate(go_memstats_mallocs_total{` + p.platformOpts.LabelQuery + `}[1m])`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Rate of a Pointer Dereferences",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "ops",
			Query: []grafana.Query{
				{
					Expr:   `rate(go_memstats_lookups_total{` + p.platformOpts.LabelQuery + `}[1m])`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Goroutines",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "ops",
			Query: []grafana.Query{
				{
					Expr:   `go_goroutines{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + p.platformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	return panels
}
