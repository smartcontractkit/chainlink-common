package grafana

import (
	"maps"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
)

type Builder struct {
	dashboardBuilder            *dashboard.DashboardBuilder
	alertsBuilder               []*alerting.RuleBuilder
	contactPointsBuilder        []*alerting.ContactPointBuilder
	notificationPoliciesBuilder []*alerting.NotificationPolicyBuilder
	panelCounter                uint32
	alertsTags                  map[string]string
}

type BuilderOptions struct {
	Name       string
	Tags       []string
	Refresh    string
	TimeFrom   string
	TimeTo     string
	TimeZone   string
	AlertsTags map[string]string
}

func NewBuilder(options *BuilderOptions) *Builder {
	if options.TimeZone == "" {
		options.TimeZone = common.TimeZoneBrowser
	}

	builder := &Builder{
		dashboardBuilder: dashboard.NewDashboardBuilder(options.Name).
			Tags(options.Tags).
			Refresh(options.Refresh).
			Time(options.TimeFrom, options.TimeTo).
			Timezone(options.TimeZone),
	}

	if options.AlertsTags != nil {
		builder.alertsTags = options.AlertsTags
	}

	return builder
}

func (b *Builder) AddVars(items ...cog.Builder[dashboard.VariableModel]) {
	for _, item := range items {
		b.dashboardBuilder.WithVariable(item)
	}
}

func (b *Builder) AddRow(title string) {
	b.dashboardBuilder.WithRow(dashboard.NewRowBuilder(title))
}

func (b *Builder) getPanelCounter() uint32 {
	b.panelCounter = inc(&b.panelCounter)
	res := b.panelCounter
	return res
}

func (b *Builder) AddPanel(panel ...*Panel) {
	for _, item := range panel {
		panelID := b.getPanelCounter()
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
		} else if item.heatmapBuilder != nil {
			item.heatmapBuilder.Id(panelID)
			b.dashboardBuilder.WithPanel(item.heatmapBuilder)
		}
		if item.alertBuilder != nil {
			b.alertsBuilder = append(b.alertsBuilder, item.alertBuilder)
		}
	}
}

func (b *Builder) AddAlert(alerts ...*alerting.RuleBuilder) {
	b.alertsBuilder = append(b.alertsBuilder, alerts...)
}

func (b *Builder) AddContactPoint(contactPoints ...*alerting.ContactPointBuilder) {
	b.contactPointsBuilder = append(b.contactPointsBuilder, contactPoints...)
}

func (b *Builder) AddNotificationPolicy(notificationPolicies ...*alerting.NotificationPolicyBuilder) {
	b.notificationPoliciesBuilder = append(b.notificationPoliciesBuilder, notificationPolicies...)
}

func (b *Builder) Build() (*Dashboard, error) {
	db, errBuildDashboard := b.dashboardBuilder.Build()
	if errBuildDashboard != nil {
		return nil, errBuildDashboard
	}

	var alerts []alerting.Rule
	for _, alertBuilder := range b.alertsBuilder {
		alert, errBuildAlert := alertBuilder.Build()
		if errBuildAlert != nil {
			return nil, errBuildAlert
		}

		// Add common tags to alerts
		if b.alertsTags != nil && len(b.alertsTags) > 0 {
			tags := maps.Clone(b.alertsTags)
			maps.Copy(tags, alert.Labels)

			alertBuildWithTags := alertBuilder.Labels(tags)
			alertWithTags, errBuildAlertWithTags := alertBuildWithTags.Build()
			if errBuildAlertWithTags != nil {
				return nil, errBuildAlertWithTags
			}
			alerts = append(alerts, alertWithTags)
		} else {
			alerts = append(alerts, alert)
		}
	}

	var contactPoints []alerting.ContactPoint
	for _, contactPointBuilder := range b.contactPointsBuilder {
		contactPoint, errBuildContactPoint := contactPointBuilder.Build()
		if errBuildContactPoint != nil {
			return nil, errBuildContactPoint
		}
		contactPoints = append(contactPoints, contactPoint)
	}

	var notificationPolicies []alerting.NotificationPolicy
	for _, notificationPolicyBuilder := range b.notificationPoliciesBuilder {
		notificationPolicy, errBuildNotificationPolicy := notificationPolicyBuilder.Build()
		if errBuildNotificationPolicy != nil {
			return nil, errBuildNotificationPolicy
		}
		notificationPolicies = append(notificationPolicies, notificationPolicy)
	}

	return &Dashboard{
		Dashboard:            &db,
		Alerts:               alerts,
		ContactPoints:        contactPoints,
		NotificationPolicies: notificationPolicies,
	}, nil
}
