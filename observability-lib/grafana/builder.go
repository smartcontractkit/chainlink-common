package grafana

import (
	"errors"
	"maps"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/cog/plugins"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana/businessvariable"
	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana/polystat"
)

type entryKind int

const (
	entryRow entryKind = iota
	entryPanel
	entryPanelToRow
)

type buildEntry struct {
	kind     entryKind
	rowTitle string
	panel    *Panel
}

type Builder struct {
	dashboardBuilder            *dashboard.DashboardBuilder
	alertsBuilder               []*alerting.RuleBuilder
	alertGroupsBuilder          []*alerting.RuleGroupBuilder
	contactPointsBuilder        []*alerting.ContactPointBuilder
	notificationPoliciesBuilder []*alerting.NotificationPolicyBuilder
	panelCounter                uint32
	alertsTags                  map[string]string
	rows                        map[string]*dashboard.RowBuilder
	entries                     []buildEntry
	built                       bool
}

type BuilderOptions struct {
	Name         string
	Tags         []string
	Refresh      string
	TimeFrom     string
	TimeTo       string
	TimeZone     string
	GraphTooltip dashboard.DashboardCursorSync
	AlertsTags   map[string]string
}

func NewBuilder(options *BuilderOptions) *Builder {
	plugins.RegisterDefaultPlugins()
	cog.NewRuntime().RegisterPanelcfgVariant(businessvariable.VariantConfig())
	cog.NewRuntime().RegisterPanelcfgVariant(polystat.VariantConfig())

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
		builder.dashboardBuilder.
			Timezone(options.TimeZone).
			Tooltip(options.GraphTooltip)
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
	row := dashboard.NewRowBuilder(title)
	if b.rows == nil {
		b.rows = make(map[string]*dashboard.RowBuilder)
	}
	b.rows[title] = row
	b.entries = append(b.entries, buildEntry{kind: entryRow, rowTitle: title})
}

func (b *Builder) getPanelCounter() uint32 {
	b.panelCounter = inc(&b.panelCounter)
	res := b.panelCounter
	return res
}

func (b *Builder) AddPanelToRow(rowTitle string, panel ...*Panel) {
	for _, item := range panel {
		b.entries = append(b.entries, buildEntry{kind: entryPanelToRow, rowTitle: rowTitle, panel: item})
		if len(item.alertBuilders) > 0 {
			b.AddAlert(item.alertBuilders...)
		}
	}
}

func (b *Builder) AddPanel(panel ...*Panel) {
	for _, item := range panel {
		b.entries = append(b.entries, buildEntry{kind: entryPanel, panel: item})
		if len(item.alertBuilders) > 0 {
			b.AddAlert(item.alertBuilders...)
		}
	}
}

func (b *Builder) AddAlert(alerts ...*alerting.RuleBuilder) {
	b.alertsBuilder = append(b.alertsBuilder, alerts...)
}

func (b *Builder) AddAlertGroup(alertGroups ...*alerting.RuleGroupBuilder) {
	b.alertGroupsBuilder = append(b.alertGroupsBuilder, alertGroups...)
}

func (b *Builder) AddContactPoint(contactPoints ...*alerting.ContactPointBuilder) {
	b.contactPointsBuilder = append(b.contactPointsBuilder, contactPoints...)
}

func (b *Builder) AddNotificationPolicy(notificationPolicies ...*alerting.NotificationPolicyBuilder) {
	b.notificationPoliciesBuilder = append(b.notificationPoliciesBuilder, notificationPolicies...)
}

// addPanelToBuilder assigns an ID and adds the panel to the dashboard builder.
func (b *Builder) addPanelToBuilder(item *Panel) {
	panelID := b.getPanelCounter()
	if item.statPanelBuilder != nil {
		item.statPanelBuilder.Id(panelID)
		b.dashboardBuilder.WithPanel(item.statPanelBuilder)
	} else if item.timeSeriesPanelBuilder != nil {
		item.timeSeriesPanelBuilder.Id(panelID)
		b.dashboardBuilder.WithPanel(item.timeSeriesPanelBuilder)
	} else if item.barGaugePanelBuilder != nil {
		item.barGaugePanelBuilder.Id(panelID)
		b.dashboardBuilder.WithPanel(item.barGaugePanelBuilder)
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
	} else if item.textPanelBuilder != nil {
		item.textPanelBuilder.Id(panelID)
		b.dashboardBuilder.WithPanel(item.textPanelBuilder)
	} else if item.histogramPanelBuilder != nil {
		item.histogramPanelBuilder.Id(panelID)
		b.dashboardBuilder.WithPanel(item.histogramPanelBuilder)
	} else if item.businessVariablePanelBuilder != nil {
		item.businessVariablePanelBuilder.Id(panelID)
		b.dashboardBuilder.WithPanel(item.businessVariablePanelBuilder)
	} else if item.polystatPanelBuilder != nil {
		item.polystatPanelBuilder.Id(panelID)
		b.dashboardBuilder.WithPanel(item.polystatPanelBuilder)
	}
}

// addPanelToRow assigns an ID and adds the panel to a row builder.
func (b *Builder) addPanelToRow(row *dashboard.RowBuilder, item *Panel) {
	panelID := b.getPanelCounter()
	if item.statPanelBuilder != nil {
		item.statPanelBuilder.Id(panelID)
		row.WithPanel(item.statPanelBuilder)
	} else if item.timeSeriesPanelBuilder != nil {
		item.timeSeriesPanelBuilder.Id(panelID)
		row.WithPanel(item.timeSeriesPanelBuilder)
	} else if item.barGaugePanelBuilder != nil {
		item.barGaugePanelBuilder.Id(panelID)
		row.WithPanel(item.barGaugePanelBuilder)
	} else if item.gaugePanelBuilder != nil {
		item.gaugePanelBuilder.Id(panelID)
		row.WithPanel(item.gaugePanelBuilder)
	} else if item.tablePanelBuilder != nil {
		item.tablePanelBuilder.Id(panelID)
		row.WithPanel(item.tablePanelBuilder)
	} else if item.logPanelBuilder != nil {
		item.logPanelBuilder.Id(panelID)
		row.WithPanel(item.logPanelBuilder)
	} else if item.heatmapBuilder != nil {
		item.heatmapBuilder.Id(panelID)
		row.WithPanel(item.heatmapBuilder)
	} else if item.textPanelBuilder != nil {
		item.textPanelBuilder.Id(panelID)
		row.WithPanel(item.textPanelBuilder)
	} else if item.histogramPanelBuilder != nil {
		item.histogramPanelBuilder.Id(panelID)
		row.WithPanel(item.histogramPanelBuilder)
	} else if item.businessVariablePanelBuilder != nil {
		item.businessVariablePanelBuilder.Id(panelID)
		row.WithPanel(item.businessVariablePanelBuilder)
	} else if item.polystatPanelBuilder != nil {
		item.polystatPanelBuilder.Id(panelID)
		row.WithPanel(item.polystatPanelBuilder)
	}
}

func (b *Builder) Build() (*Observability, error) {
	if b.built {
		return nil, errors.New("Build() has already been called; create a new Builder for a new build")
	}
	b.built = true

	observability := Observability{}

	if b.dashboardBuilder != nil {
		// First pass: attach panels to their row builders (needed before WithRow snapshots them)
		for _, e := range b.entries {
			if e.kind == entryPanelToRow {
				if row, ok := b.rows[e.rowTitle]; ok {
					b.addPanelToRow(row, e.panel)
				}
			}
		}

		// Second pass: add rows and top-level panels to the dashboard in order
		for _, e := range b.entries {
			switch e.kind {
			case entryRow:
				if row, ok := b.rows[e.rowTitle]; ok {
					b.dashboardBuilder.WithRow(row)
				}
			case entryPanel:
				b.addPanelToBuilder(e.panel)
			default:
				continue
			}
		}

		db, errBuildDashboard := b.dashboardBuilder.Build()
		if errBuildDashboard != nil {
			return nil, errBuildDashboard
		}
		observability.Dashboard = &db
	}

	var alerts []alerting.Rule
	for _, alertBuilder := range b.alertsBuilder {
		alert, errBuildAlert := alertBuilder.Build()
		if errBuildAlert != nil {
			return nil, errBuildAlert
		}

		// Add common tags to alerts
		if len(b.alertsTags) > 0 {
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

	var alertGroups []alerting.RuleGroup
	for _, alertGroupBuilder := range b.alertGroupsBuilder {
		alertGroup, errBuildAlertGroup := alertGroupBuilder.Build()
		if errBuildAlertGroup != nil {
			return nil, errBuildAlertGroup
		}
		alertGroups = append(alertGroups, alertGroup)
	}
	observability.AlertGroups = alertGroups

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
