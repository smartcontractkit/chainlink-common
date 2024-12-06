package grafana

import (
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
)

type AlertGroupOptions struct {
	Title    string
	Interval alerting.Duration // duration in seconds
}

func NewAlertGroup(options *AlertGroupOptions) *alerting.RuleGroupBuilder {
	return alerting.NewRuleGroupBuilder(options.Title).
		Interval(options.Interval)
}
