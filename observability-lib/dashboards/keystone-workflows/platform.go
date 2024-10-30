package keystone_workflows

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

type platformOpts struct {
	LabelFilters map[string]string
	LabelFilter  string
	LegendString string
	LabelQuery   string
}

type Props struct {
	Name              string              // required: Name is the name of the dashboard
	MetricsDataSource *grafana.DataSource // required: MetricsDataSource is the datasource for querying metrics
	LogsDataSource    *grafana.DataSource // required: LogsDataSource is the datasource for querying logs
	platformOpts      platformOpts
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

func platformBuildOpts(props *Props) error {
	if err := validateInput(props); err != nil {
		return err
	}
	if !props.Tested {
		po := platformOpts{
			LabelFilters: map[string]string{
				"env":           `=~"${env}"`,
				"cluster":       `=~"${cluster}"`,
				"workflowOwner": `=~"${workflowOwner}"`,
				"workflowName":  `=~"${workflowName}"`,
			},
		}
		for key, value := range po.LabelFilters {
			po.LabelQuery += key + value + ", "
		}
		props.platformOpts = po
	}
	return nil
}
