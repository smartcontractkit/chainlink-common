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
	builder := &Builder{}

	if options.Name != "" {
		builder.dashboardBuilder = dashboard.NewDashboardBuilder(options.Name)
		if options.Tags != nil {
			builder.dashboardBuilder.Tags(options.Tags)
		}
		if options.Refresh != "" {
			builder.dashboardBuilder.Refresh(options.Refresh)
		}
		if options.TimeFrom != "" && options.TimeTo != "" {
			builder.dashboardBuilder.Time(options.TimeFrom, options.TimeTo)
		}
		if options.TimeZone == "" {
			options.TimeZone = common.TimeZoneBrowser
		}
		builder.dashboardBuilder.Timezone(options.TimeZone)
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
		if item.alertBuilders != nil && len(item.alertBuilders) > 0 {
			b.AddAlert(item.alertBuilders...)
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

func (b *Builder) Build() (*Observability, error) {
	observability := Observability{}

	if b.dashboardBuilder != nil {
		db, errBuildDashboard := b.dashboardBuilder.Build()
		if errBuildDashboard != nil {
			return nil, errBuildDashboard
		}
		observability.Dashboard = &db

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
		observability.Alerts = alerts
	}

	var contactPoints []alerting.ContactPoint
	for _, contactPointBuilder := range b.contactPointsBuilder {
		contactPoint, errBuildContactPoint := contactPointBuilder.Build()
		if errBuildContactPoint != nil {
			return nil, errBuildContactPoint
		}
		contactPoints = append(contactPoints, contactPoint)
	}
	observability.ContactPoints = contactPoints

	var notificationPolicies []alerting.NotificationPolicy
	for _, notificationPolicyBuilder := range b.notificationPoliciesBuilder {
		notificationPolicy, errBuildNotificationPolicy := notificationPolicyBuilder.Build()
		if errBuildNotificationPolicy != nil {
			return nil, errBuildNotificationPolicy
		}
		notificationPolicies = append(notificationPolicies, notificationPolicy)
	}
	observability.NotificationPolicies = notificationPolicies

	return &observability, nil
}
