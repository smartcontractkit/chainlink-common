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

	props.PlatformOpts = PlatformPanelOpts(props.Platform)

	builder := grafana.NewBuilder(&grafana.BuilderOptions{
		Name:     props.Name,
		Tags:     []string{"Core", "Node"},
		Refresh:  "30s",
		TimeFrom: "now-30m",
		TimeTo:   "now",
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

	builder.AddRow("LogPoller")
	builder.AddPanel(logPoller(props)...)

	builder.AddRow("Feeds Jobs")
	builder.AddPanel(feedsJobs(props)...)

	builder.AddRow("Mailbox")
	builder.AddPanel(mailbox(props)...)

	builder.AddRow("PromReporter")
	builder.AddPanel(promReporter(props)...)

	builder.AddRow("TxManager")
	builder.AddPanel(txManager(props)...)

	builder.AddRow("HeadTracker")
	builder.AddPanel(headTracker(props)...)

	builder.AddRow("Logs Counters")
	builder.AddPanel(logsCounters(props)...)

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

	if p.PlatformOpts.Platform == "kubernetes" {
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
		}))

		variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Label: "Pod",
				Name:  "pod",
			},
			Datasource: p.MetricsDataSource.Name,
			Query:      `label_values(up{env="$env", cluster="$cluster", namespace="$namespace", job="$job"}, pod)`,
			Multi:      true,
		}))
	} else {
		variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
			VariableOption: &grafana.VariableOption{
				Label: "Instance",
				Name:  "instance",
			},
			Datasource: p.MetricsDataSource.Name,
			Query:      fmt.Sprintf("label_values(%s)", p.PlatformOpts.LabelFilter),
			Multi:      true,
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
					Expr:    `version{` + p.PlatformOpts.LabelQuery + `}`,
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
					Expr:   `uptime_seconds{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:    `eth_balance{` + p.PlatformOpts.LabelQuery + `}`,
					Legend:  `{{` + p.PlatformOpts.LegendString + `}} - {{account}}`,
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
					Expr:    `solana_balance{` + p.PlatformOpts.LabelQuery + `}`,
					Legend:  `{{` + p.PlatformOpts.LegendString + `}} - {{account}}`,
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
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Components Health Avg by Service over 15m",
			Description: "Only displays services with health average < 90%",
			Span:        24,
			Height:      4,
			Decimals:    1,
			Unit:        "percent",
			Query: []grafana.Query{
				{
					Expr:   `100 * avg(avg_over_time(health{` + p.PlatformOpts.LabelQuery + `}[15m])) by (service_id, version, service, cluster, env) < 90`,
					Legend: "{{service_id}}",
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
					Expr:   `eth_balance{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{account}}`,
				},
			},
			AlertOptions: &grafana.AlertOptions{
				Summary:     `ETH Balance is lower than threshold`,
				Description: `ETH Balance critically low at {{ index $values "A" }} on {{ index $labels "` + p.PlatformOpts.LegendString + `" }}`,
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
						Datasource: p.MetricsDataSource.UID,
					},
				},
				Condition: &grafana.ConditionQuery{
					RefID: "B",
					ThresholdExpression: &grafana.ThresholdExpression{
						Expression: "A",
						ThresholdConditionsOptions: []grafana.ThresholdConditionsOption{
							{
								Params: []float64{2, 0},
								Type:   expr.TypeThresholdTypeLt,
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
					Expr:   `solana_balance{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{account}}`,
				},
			},
		},
	}))

	if p.PlatformOpts.Platform == "kubernetes" {
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
				Height:     6,
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
				Height:     6,
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
					Expr:   `process_open_fds{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:    `go_info{` + p.PlatformOpts.LabelQuery + `}`,
					Legend:  "{{version}}",
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
					Expr:   `db_conns_max{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Max`,
				},
				{
					Expr:   `db_conns_open{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Open`,
				},
				{
					Expr:   `db_conns_used{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Used`,
				},
				{
					Expr:   `db_conns_wait{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Wait`,
				},
			},
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
					Expr:   `db_wait_count{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `db_wait_time_seconds{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `histogram_quantile(0.9, sum(rate(sql_query_timeout_percent_bucket{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (le))`,
					Legend: "p90",
				},
				{
					Expr:   `histogram_quantile(0.95, sum(rate(sql_query_timeout_percent_bucket{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (le))`,
					Legend: "p95",
				},
				{
					Expr:   `histogram_quantile(0.99, sum(rate(sql_query_timeout_percent_bucket{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (le))`,
					Legend: "p99",
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
					Expr:   `count(log_poller_query_duration_sum{` + p.PlatformOpts.LabelQuery + `}) by (evmChainID)`,
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
					Expr:   `avg by (query) (sum by (query, job) (rate(log_poller_query_duration_count{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])))`,
					Legend: "{{query}} - {{job}}",
				},
				{
					Expr:   `avg (sum by(job) (rate(log_poller_query_duration_count{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])))`,
					Legend: "Total",
				},
			},
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
					Expr:   `avg by (type) (sum by (type, job) (rate(log_poller_query_duration_count{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])))`,
					Legend: "{{query}} - {{job}}",
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
					Expr:   `avg by (query) (log_poller_query_dataset_size{` + p.PlatformOpts.LabelQuery + `})`,
					Legend: "{{query}} - {{job}}",
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
					Expr:   `max by (query) (log_poller_query_dataset_size{` + p.PlatformOpts.LabelQuery + `})`,
					Legend: "{{query}} - {{job}}",
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
					Expr:   `max by (evmChainID) (log_poller_query_dataset_size{` + p.PlatformOpts.LabelQuery + `})`,
					Legend: "{{evmChainID}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Queries duration by type (0.5 perc)",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `histogram_quantile(0.5, sum(rate(log_poller_query_duration_bucket{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (le, query)) / 1e6`,
					Legend: "{{query}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Queries duration by type (0.9 perc)",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `histogram_quantile(0.9, sum(rate(log_poller_query_duration_bucket{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (le, query)) / 1e6`,
					Legend: "{{query}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Queries duration by type (0.99 perc)",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `histogram_quantile(0.99, sum(rate(log_poller_query_duration_bucket{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (le, query)) / 1e6`,
					Legend: "{{query}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Number of logs inserted",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Query: []grafana.Query{
				{
					Expr:   `avg by (evmChainID) (log_poller_logs_inserted{` + p.PlatformOpts.LabelQuery + `})`,
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
					Expr:   `avg by (evmChainID) (rate(log_poller_logs_inserted{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval]))`,
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
					Expr:   `avg by (evmChainID) (log_poller_blocks_inserted{` + p.PlatformOpts.LabelQuery + `})`,
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
					Expr:   `avg by (evmChainID) (rate(log_poller_blocks_inserted{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval]))`,
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
					Expr:   `feeds_job_proposal_requests{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `feeds_job_proposal_count{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
			Query: []grafana.Query{
				{
					Expr:   `mailbox_load_percent{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{name}}`,
				},
			},
		},
	}))

	return panels
}

func promReporter(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Unconfirmed Transactions",
			Span:       8,
			Height:     6,
			Decimals:   1,
			Unit:       "Tx",
			Query: []grafana.Query{
				{
					Expr:   `unconfirmed_transactions{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
			Decimals:   1,
			Unit:       "s",
			Query: []grafana.Query{
				{
					Expr:   `max_unconfirmed_tx_age{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
			Decimals:   1,
			Unit:       "Blocks",
			Query: []grafana.Query{
				{
					Expr:   `max_unconfirmed_blocks{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
			Title:      "TX Manager Time Until TX Broadcast",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `tx_manager_time_until_tx_broadcast{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Num Gas Bumps",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `tx_manager_num_gas_bumps{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Num Gas Bumps Exceeds Limit",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `tx_manager_gas_bump_exceeds_limit{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Num Confirmed Transactions",
			Span:       6,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `tx_manager_num_confirmed_transactions{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Num Successful Transactions",
			Span:       6,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `tx_manager_num_successful_transactions{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Num Reverted Transactions",
			Span:       6,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `tx_manager_num_tx_reverted{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Num Fwd Transactions",
			Span:       6,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `tx_manager_fwd_tx_count{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Num Transactions Attempts",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `tx_manager_tx_attempt_count{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Time Until TX Confirmed",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `tx_manager_time_until_tx_confirmed{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "TX Manager Block Until TX Confirmed",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `tx_manager_blocks_until_tx_confirmed{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
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
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `head_tracker_current_head{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
			Decimals:   1,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `head_tracker_very_old_head{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Head Tracker Heads Received",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `head_tracker_heads_received{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Head Tracker Connection Errors",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "Block",
			Query: []grafana.Query{
				{
					Expr:   `head_tracker_connection_errors{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	return panels
}

func logsCounters(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Logs Counter",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `log_panic_count{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - panic`,
				},
				{
					Expr:   `log_fatal_count{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - fatal`,
				},
				{
					Expr:   `log_critical_count{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - critical`,
				},
				{
					Expr:   `log_warn_count{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - warn`,
				},
				{
					Expr:   `log_error_count{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - error`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Logs Rate",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `sum(rate(log_panic_count{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - panic`,
				},
				{
					Expr:   `sum(rate(log_fatal_count{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - fatal`,
				},
				{
					Expr:   `sum(rate(log_critical_count{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - critical`,
				},
				{
					Expr:   `sum(rate(log_warn_count{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - warn`,
				},
				{
					Expr:   `sum(rate(log_error_count{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - error`,
				},
			},
		},
	}))

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
					Expr:   `evm_pool_rpc_node_highest_seen_block{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `evm_pool_rpc_node_num_seen_blocks{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `evm_pool_rpc_node_polls_total{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `evm_pool_rpc_node_polls_failed{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `evm_pool_rpc_node_polls_success{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	return panels
}

func nodesRPC(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Node RPC Alive",
			Span:       6,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `sum(multi_node_states{` + p.PlatformOpts.LabelQuery + `state="Alive"}) by (` + p.PlatformOpts.LegendString + `, chainId)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{chainId}}`,
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Node RPC Closed",
			Description: "",
			Span:        6,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum(multi_node_states{` + p.PlatformOpts.LabelQuery + `state="Closed"}) by (` + p.PlatformOpts.LegendString + `, chainId)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{chainId}}`,
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Node RPC Dialed",
			Description: "",
			Span:        6,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum(multi_node_states{` + p.PlatformOpts.LabelQuery + `state="Dialed"}) by (` + p.PlatformOpts.LegendString + `, chainId)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{chainId}}`,
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Node RPC InvalidChainID",
			Description: "",
			Span:        6,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum(multi_node_states{` + p.PlatformOpts.LabelQuery + `state="InvalidChainID"}) by (` + p.PlatformOpts.LegendString + `, chainId)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{chainId}}`,
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Node RPC OutOfSync",
			Description: "",
			Span:        6,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum(multi_node_states{` + p.PlatformOpts.LabelQuery + `state="OutOfSync"}) by (` + p.PlatformOpts.LegendString + `, chainId)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{chainId}}`,
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Node RPC UnDialed",
			Description: "",
			Span:        6,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum(multi_node_states{` + p.PlatformOpts.LabelQuery + `state="Undialed"}) by (` + p.PlatformOpts.LegendString + `, chainId)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{chainId}}`,
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Node RPC Unreachable",
			Description: "",
			Span:        6,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum(multi_node_states{` + p.PlatformOpts.LabelQuery + `state="Unreachable"}) by (` + p.PlatformOpts.LegendString + `, chainId)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{chainId}}`,
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Node RPC Unusable",
			Description: "",
			Span:        6,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum(multi_node_states{` + p.PlatformOpts.LabelQuery + `state="Unusable"}) by (` + p.PlatformOpts.LegendString + `, chainId)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{chainId}}`,
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

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
			Decimals:   1,
			Unit:       "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `sum(increase(evm_pool_rpc_node_calls_success{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, evmChainID, nodeName) / sum(increase(evm_pool_rpc_node_calls_total{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, evmChainID, nodeName)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{evmChainID}} - {{nodeName}}`,
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
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool RPC Node Dials Failure Rate",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `sum(increase(evm_pool_rpc_node_dials_failed{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, evmChainID, nodeName) / sum(increase(evm_pool_rpc_node_calls_total{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, evmChainID, nodeName)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{evmChainID}} - {{nodeName}}`,
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
					Expr:   `evm_pool_rpc_node_num_transitions_to_alive{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: "Alive",
				},
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_in_sync{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: "InSync",
				},
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_out_of_sync{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: "OutOfSync",
				},
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_unreachable{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: "UnReachable",
				},
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_invalid_chain_id{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: "InvalidChainID",
				},
				{
					Expr:   `evm_pool_rpc_node_num_transitions_to_unusable{` + p.PlatformOpts.LabelQuery + `}`,
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
					Expr:   `evm_pool_rpc_node_states{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{evmChainID}} - {{state}}`,
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
					Expr:   `sum(increase(evm_pool_rpc_node_verifies_success{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, evmChainID, nodeName) / sum(increase(evm_pool_rpc_node_verifies{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, evmChainID, nodeName) * 100`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{evmChainID}} - {{nodeName}}`,
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
					Expr:   `sum(increase(evm_pool_rpc_node_verifies_failed{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, evmChainID, nodeName) / sum(increase(evm_pool_rpc_node_verifies{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, evmChainID, nodeName) * 100`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{evmChainID}} - {{nodeName}}`,
				},
			},
		},
	}))

	return panels
}

func evmPoolRPCNodeLatencies(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool RPC Node Calls Latency 0.90 quantile",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `histogram_quantile(0.90, sum(rate(evm_pool_rpc_node_rpc_call_time_bucket{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, le, rpcCallName)) / 1e6`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{rpcCallName}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool RPC Node Calls Latency 0.95 quantile",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `histogram_quantile(0.95, sum(rate(evm_pool_rpc_node_rpc_call_time_bucket{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, le, rpcCallName)) / 1e6`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{rpcCallName}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "EVM Pool RPC Node Calls Latency 0.99 quantile",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Unit:       "ms",
			Query: []grafana.Query{
				{
					Expr:   `histogram_quantile(0.99, sum(rate(evm_pool_rpc_node_rpc_call_time_bucket{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, le, rpcCallName)) / 1e6`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{rpcCallName}}`,
				},
			},
		},
	}))

	return panels
}

func evmBlockHistoryEstimator(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Gas Updater All Gas Price Percentiles",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `gas_updater_all_gas_price_percentiles{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{percentile}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Gas Updater All Tip Cap Percentiles",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `gas_updater_all_tip_cap_percentiles{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{percentile}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Gas Updater Set Gas Price",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `gas_updater_set_gas_price{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `gas_updater_set_tip_cap{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `gas_updater_current_base_fee{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `block_history_estimator_connectivity_failure_count{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `pipeline_task_execution_time{` + p.PlatformOpts.LabelQuery + `} / 1e6`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} JobID: {{job_id}}`,
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
					Expr:   `pipeline_run_errors{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} JobID: {{job_id}}`,
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
					Expr:   `pipeline_run_total_time_to_completion{` + p.PlatformOpts.LabelQuery + `} / 1e6`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} JobID: {{job_id}}`,
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
					Expr:   `pipeline_tasks_total_finished{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} JobID: {{job_id}}`,
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
					Expr:   `histogram_quantile(0.95, sum(rate(service_gonic_request_duration_bucket{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, le, path, method))`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{method}} - {{path}}`,
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
					Expr:   `sum(rate(service_gonic_requests_total{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, path, method, code)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{method}} - {{path}} - {{code}}`,
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
					Expr:   `avg(rate(service_gonic_request_size_bytes_sum{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `)/avg(rate(service_gonic_request_size_bytes_count{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `avg(rate(service_gonic_response_size_bytes_sum{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `)/avg(rate(service_gonic_response_size_bytes_count{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `sum(rate(promhttp_metric_handler_requests_total{` + p.PlatformOpts.LabelQuery + `}[$__rate_interval])) by (` + p.PlatformOpts.LegendString + `, code)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - {{code}}`,
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
					Expr:   `sum(go_threads{` + p.PlatformOpts.LabelQuery + `}) by (` + p.PlatformOpts.LegendString + `)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `sum(go_memstats_heap_alloc_bytes{` + p.PlatformOpts.LabelQuery + `}) by (` + p.PlatformOpts.LegendString + `)`,
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
					Expr:   `sum(go_memstats_heap_alloc_bytes{` + p.PlatformOpts.LabelQuery + `}) by (` + p.PlatformOpts.LegendString + `)`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `go_memstats_heap_alloc_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Alloc`,
				},
				{
					Expr:   `go_memstats_heap_sys_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Sys`,
				},
				{
					Expr:   `go_memstats_heap_idle_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Idle`,
				},
				{
					Expr:   `go_memstats_heap_inuse_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - InUse`,
				},
				{
					Expr:   `go_memstats_heap_released_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Released`,
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
					Expr:   `go_memstats_mspan_inuse_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Total InUse`,
				},
				{
					Expr:   `go_memstats_mspan_sys_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Total Sys`,
				},
				{
					Expr:   `go_memstats_mcache_inuse_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Cache InUse`,
				},
				{
					Expr:   `go_memstats_mcache_sys_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Cache Sys`,
				},
				{
					Expr:   `go_memstats_buck_hash_sys_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Hash Sys`,
				},
				{
					Expr:   `go_memstats_gc_sys_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - GC Sys`,
				},
				{
					Expr:   `go_memstats_other_sys_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - bytes of memory are used for other runtime allocations`,
				},
				{
					Expr:   `go_memstats_next_gc_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Next GC`,
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
					Expr:   `go_memstats_stack_inuse_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - InUse`,
				},
				{
					Expr:   `go_memstats_stack_sys_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}} - Sys`,
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
					Expr:   `go_memstats_sys_bytes{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `go_memstats_mallocs_total{` + p.PlatformOpts.LabelQuery + `} - go_memstats_frees_total{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `rate(go_memstats_mallocs_total{` + p.PlatformOpts.LabelQuery + `}[1m])`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `rate(go_memstats_lookups_total{` + p.PlatformOpts.LabelQuery + `}[1m])`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
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
					Expr:   `go_goroutines{` + p.PlatformOpts.LabelQuery + `}`,
					Legend: `{{` + p.PlatformOpts.LegendString + `}}`,
				},
			},
		},
	}))

	return panels
}
