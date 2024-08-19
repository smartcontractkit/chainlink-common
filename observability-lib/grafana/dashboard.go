package grafana

import (
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
)

type TypePlatform string

const (
	TypePlatformKubernetes TypePlatform = "kubernetes"
	TypePlatformDocker     TypePlatform = "docker"
)

type DashboardOptions struct {
	Name              string
	FolderUID         string
	Platform          TypePlatform
	OCRVersion        string
	MetricsDataSource *DataSource
	LogsDataSource    *DataSource
	AlertsTags        map[string]string
	SlackWebhookURL   string
	SlackToken        string
	SlackChannel      string
}

type Dashboard struct {
	Dashboard     *dashboard.Dashboard
	Alerts        []alerting.Rule
	ContactPoints []alerting.ContactPoint
}
