package k8sresources

import (
	"fmt"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

type Props struct {
	Name              string              // Name is the name of the dashboard
	MetricsDataSource *grafana.DataSource // MetricsDataSource is the datasource for querying metrics
}

func NewDashboard(props *Props) (*grafana.Observability, error) {
	if props.Name == "" {
		return nil, fmt.Errorf("Name is required")
	}

	builder := grafana.NewBuilder(&grafana.BuilderOptions{
		Name:     props.Name,
		Tags:     []string{"Core", "Node", "Kubernetes", "Resources"},
		Refresh:  "30s",
		TimeFrom: "now-30m",
		TimeTo:   "now",
	})

	builder.AddVars(vars(props)...)

	builder.AddRow("Headlines")
	builder.AddPanel(headlines(props)...)

	builder.AddRow("Pod Status")
	builder.AddPanel(podStatus(props)...)

	builder.AddRow("Resources Usage")
	builder.AddPanel(resourcesUsage(props)...)

	builder.AddRow("Network Usage")
	builder.AddPanel(networkUsage(props)...)

	builder.AddRow("Disk Usage")
	builder.AddPanel(diskUsage(props)...)

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

	return variables
}

func headlines(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

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
					Legend:  "{{pod}}",
					Instant: true,
				},
			},
		},
		Orientation: common.VizOrientationHorizontal,
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
					Legend:  "{{pod}}",
					Instant: true,
				},
			},
		},
		Orientation: common.VizOrientationHorizontal,
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
					Legend:  "{{pod}}",
					Instant: true,
				},
			},
		},
		Orientation: common.VizOrientationHorizontal,
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
					Legend:  "{{pod}}",
					Instant: true,
				},
			},
		},
		Orientation: common.VizOrientationHorizontal,
	}))

	return panels
}

func podStatus(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Pod Restarts",
			Description: "Number of pod restarts",
			Span:        8,
			Height:      4,
			Query: []grafana.Query{
				{
					Expr:   `sum(increase(kube_pod_container_status_restarts_total{pod=~"$pod", namespace=~"${namespace}"}[$__rate_interval])) by (pod)`,
					Legend: "{{pod}}",
				},
			},
		},
		ColorMode:   common.BigValueColorModeNone,
		GraphMode:   common.BigValueGraphModeLine,
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "OOM Events",
			Description: "Out-of-memory number of events",
			Span:        8,
			Height:      4,
			Query: []grafana.Query{
				{
					Expr:   `sum(container_oom_events_total{pod=~"$pod", namespace=~"${namespace}"}) by (pod)`,
					Legend: "{{pod}}",
				},
			},
		},
		ColorMode:   common.BigValueColorModeNone,
		GraphMode:   common.BigValueGraphModeLine,
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "OOM Killed",
			Span:       8,
			Height:     4,
			Query: []grafana.Query{
				{
					Expr:   `kube_pod_container_status_last_terminated_reason{reason="OOMKilled", pod=~"$pod", namespace=~"${namespace}"}`,
					Legend: "{{pod}}",
				},
			},
		},
		ColorMode:   common.BigValueColorModeNone,
		GraphMode:   common.BigValueGraphModeLine,
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	return panels
}

func resourcesUsage(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

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
	}))

	return panels
}

func networkUsage(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Receive Bandwidth",
			Span:       12,
			Height:     6,
			Unit:       "bps",
			Query: []grafana.Query{
				{
					Expr:   `sum(irate(container_network_receive_bytes_total{pod=~"$pod", namespace=~"${namespace}"}[$__rate_interval])) by (pod)`,
					Legend: "{{pod}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Transmit Bandwidth",
			Span:       12,
			Height:     6,
			Unit:       "bps",
			Query: []grafana.Query{
				{
					Expr:   `sum(irate(container_network_transmit_bytes_total{pod=~"$pod", namespace=~"${namespace}"}[$__rate_interval])) by (pod)`,
					Legend: "{{pod}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Average Container Bandwidth: Received",
			Span:       12,
			Height:     6,
			Unit:       "bps",
			Query: []grafana.Query{
				{
					Expr:   `avg(irate(container_network_receive_bytes_total{pod=~"$pod", namespace=~"${namespace}"}[$__rate_interval])) by (pod)`,
					Legend: "{{pod}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Average Container Bandwidth: Transmitted",
			Span:       12,
			Height:     6,
			Unit:       "bps",
			Query: []grafana.Query{
				{
					Expr:   `avg(irate(container_network_transmit_bytes_total{pod=~"$pod", namespace=~"${namespace}"}[$__rate_interval])) by (pod)`,
					Legend: "{{pod}}",
				},
			},
		},
	}))

	return panels
}

func diskUsage(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "IOPS(Read+Write)",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Unit:       "short",
			Query: []grafana.Query{
				{
					Expr:   `ceil(sum by(container, pod) (rate(container_fs_reads_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", cluster="$cluster", namespace="$namespace", pod="$pod"}[$__rate_interval]) + rate(container_fs_writes_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", cluster="$cluster", namespace="$namespace", pod="$pod"}[$__rate_interval])))`,
					Legend: "{{pod}}",
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "ThroughPut(Read+Write)",
			Span:       12,
			Height:     6,
			Decimals:   2,
			Unit:       "short",
			Query: []grafana.Query{
				{
					Expr:   `sum by(container, pod) (rate(container_fs_reads_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", cluster="$cluster", namespace="$namespace", pod="$pod"}[$__rate_interval]) + rate(container_fs_writes_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", cluster="$cluster", namespace="$namespace", pod="$pod"}[$__rate_interval]))`,
					Legend: "{{pod}}",
				},
			},
		},
	}))

	return panels
}
