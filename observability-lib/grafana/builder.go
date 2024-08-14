package grafana

import (
	"maps"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
)

type Builder struct {
	dashboardBuilder *dashboard.DashboardBuilder
	alertsBuilder    []*alerting.RuleBuilder
	panelCounter     uint32
	alertsTags       map[string]string
}

type BuilderOptions struct {
	Tags     []string
	Refresh  string
	TimeFrom string
	TimeTo   string
	TimeZone string
}

func NewBuilder(options *DashboardOptions, builderOptions *BuilderOptions) *Builder {
	if builderOptions.TimeZone == "" {
		builderOptions.TimeZone = common.TimeZoneBrowser
	}

	return &Builder{
		dashboardBuilder: dashboard.NewDashboardBuilder(options.Name).
			Tags(builderOptions.Tags).
			Refresh(builderOptions.Refresh).
			Time(builderOptions.TimeFrom, builderOptions.TimeTo).
			Timezone(builderOptions.TimeZone),
		alertsTags: options.AlertsTags,
	}
}

func (b *Builder) AddVars(items ...cog.Builder[dashboard.VariableModel]) {
	for _, item := range items {
		b.dashboardBuilder.WithVariable(item)
	}
}

func (b *Builder) AddRow(title string) {
	b.dashboardBuilder.WithRow(dashboard.NewRowBuilder(title))
}

func (b *Builder) GetPanelCounter() uint32 {
	res := b.panelCounter
	b.panelCounter = Inc(&b.panelCounter)
	return res
}

func (b *Builder) AddPanel(panel ...*Panel) {
	for _, item := range panel {
		panelID := b.GetPanelCounter()
		if item.statPanelBuilder != nil {
			item.statPanelBuilder.Id(panelID)
			b.dashboardBuilder.WithPanel(item.statPanelBuilder)
		} else if item.timeSeriesPanelBuilder != nil {
			item.timeSeriesPanelBuilder.Id(panelID)
			b.dashboardBuilder.WithPanel(item.timeSeriesPanelBuilder)
		} else if item.gaugePanelBuilder != nil {
			item.gaugePanelBuilder.Id(panelID)
			b.dashboardBuilder.WithPanel(item.gaugePanelBuilder)
		} else if item.tablePanelBuilder != nil {
			item.tablePanelBuilder.Id(panelID)
			b.dashboardBuilder.WithPanel(item.tablePanelBuilder)
		} else if item.logPanelBuilder != nil {
			item.logPanelBuilder.Id(panelID)
			b.dashboardBuilder.WithPanel(item.logPanelBuilder)
		}
		if item.alertBuilder != nil {
			b.alertsBuilder = append(b.alertsBuilder, item.alertBuilder)
		}
	}
}

func (b *Builder) Build() (*dashboard.Dashboard, []alerting.Rule, error) {
	db, errBuildDashboard := b.dashboardBuilder.Build()
	if errBuildDashboard != nil {
		return nil, nil, errBuildDashboard
	}

	var alerts []alerting.Rule
	for _, alertBuilder := range b.alertsBuilder {
		alert, errBuildAlert := alertBuilder.Build()
		if errBuildAlert != nil {
			return nil, nil, errBuildAlert
		}

		// Add common tags to alerts
		if b.alertsTags != nil && len(b.alertsTags) > 0 {
			tags := maps.Clone(b.alertsTags)
			maps.Copy(tags, alert.Labels)

			alertBuildWithTags := alertBuilder.Labels(tags)
			alertWithTags, errBuildAlertWithTags := alertBuildWithTags.Build()
			if errBuildAlertWithTags != nil {
				return nil, nil, errBuildAlertWithTags
			}
			alerts = append(alerts, alertWithTags)
		} else {
			alerts = append(alerts, alert)
		}
	}

	return &db, alerts, nil
}
