package keystone_workflows

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

type Props struct {
	Name              string              // required: Name is the name of the dashboard
	MetricsDataSource *grafana.DataSource // required: MetricsDataSource is the datasource for querying metrics
	LogsDataSource    *grafana.DataSource // required: LogsDataSource is the datasource for querying logs
	QueryFilters      string
	AlertsTitlePrefix string //optional
	AlertsFilters     string //optional
	AlertsTags        map[string]string
	SlackChannel      string // optional
	SlackWebhookURL   string //optional
	PagerDutyKey      string //optional
	Tested            bool
}

func validateInput(props *Props) error {
	if props.Name == "" {
		return fmt.Errorf("Name is required")
	}

	if props.MetricsDataSource == nil {
		return fmt.Errorf("MetricsDataSource is required")
	} else {
		if props.MetricsDataSource.Name == "" {
			return fmt.Errorf("MetricsDataSource.Name is required")
		}
		if props.MetricsDataSource.UID == "" {
			return fmt.Errorf("MetricsDataSource.UID is required")
		}
	}

	if props.LogsDataSource == nil {
		return fmt.Errorf("LogsDataSource is required")
	} else {
		if props.LogsDataSource.Name == "" {
			return fmt.Errorf("LogsDataSource.Name is required")
		}
		if props.LogsDataSource.UID == "" {
			return fmt.Errorf("LogsDataSource.UID is required")
		}
	}
	return nil
}
