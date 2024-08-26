package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

type CommandOptions struct {
	Name                  string
	GrafanaURL            string
	GrafanaToken          string
	FolderName            string
	Platform              string
	TypeDashboard         string
	MetricsDataSourceName string
	LogsDataSourceName    string
	EnableAlerts          bool
	AlertsTags            map[string]string
	NotificationTemplates string
	SlackWebhookURL       string
	SlackToken            string
	SlackChannel          string
}

func NewDashboard(options *CommandOptions) error {
	grafanaClient := api.NewClient(
		options.GrafanaURL,
		options.GrafanaToken,
	)

	folder, errFolder := grafanaClient.FindOrCreateFolder(options.FolderName)
	if errFolder != nil {
		return errFolder
	}

	metricsDataSource, _, errMetricsDataSource := grafanaClient.GetDataSourceByName(options.MetricsDataSourceName)
	if errMetricsDataSource != nil {
		return errMetricsDataSource
	}

	buildOptions := &BuildOptions{
		Name:              options.Name,
		FolderUID:         folder.UID,
		Platform:          grafana.TypePlatform(options.Platform),
		TypeDashboard:     TypeDashboard(options.TypeDashboard),
		MetricsDataSource: grafana.NewDataSource(metricsDataSource.Name, metricsDataSource.UID),
		AlertsTags:        options.AlertsTags,
		SlackWebhookURL:   options.SlackWebhookURL,
		SlackToken:        options.SlackToken,
		SlackChannel:      options.SlackChannel,
	}

	if options.LogsDataSourceName != "" {
		logsDataSource, _, errLogsDataSource := grafanaClient.GetDataSourceByName(options.LogsDataSourceName)
		if errLogsDataSource != nil {
			return errLogsDataSource
		}
		buildOptions.LogsDataSource = grafana.NewDataSource(logsDataSource.Name, logsDataSource.UID)
	}

	build, errBuild := Build(buildOptions)
	if errBuild != nil {
		return errBuild
	}

	db, _, errPostDashboard := grafanaClient.PostDashboard(api.PostDashboardRequest{
		Dashboard: build.Dashboard,
		Overwrite: true,
		FolderID:  int(folder.ID),
	})
	if errPostDashboard != nil {
		return errPostDashboard
	}

	Logger.Info().
		Str("Name", options.Name).
		Str("URL", options.GrafanaURL).
		Str("Folder", options.FolderName).
		Msg("Dashboard deployed")

	// Create alerts for the dashboard
	if options.EnableAlerts && build.Alerts != nil && len(build.Alerts) > 0 {
		// Get alert rules for the dashboard
		alertsRule, errGetAlertRules := grafanaClient.GetAlertRulesByDashboardUID(*db.UID)
		if errGetAlertRules != nil {
			return errGetAlertRules
		}

		// delete alert rules for the dashboard
		for _, rule := range alertsRule {
			_, _, errDeleteAlertRule := grafanaClient.DeleteAlertRule(*rule.Uid)
			if errDeleteAlertRule != nil {
				return errDeleteAlertRule
			}
		}

		// Create alert rules for the dashboard
		for _, alert := range build.Alerts {
			alert.Annotations["__dashboardUid__"] = *db.UID
			alert.Annotations["__panelId__"] = panelIDByTitle(build.Dashboard, alert.Title)

			_, _, errPostAlertRule := grafanaClient.PostAlertRule(alert)
			if errPostAlertRule != nil {
				return errPostAlertRule
			}
			Logger.Info().
				Str("Name", alert.Annotations["summary"]).
				Str("URL", options.GrafanaURL).
				Str("Folder", options.FolderName).
				Msg("Alert created")
		}
	}

	// Create notification templates for the alerts
	if options.NotificationTemplates != "" {
		notificationTemplates, errNotificationTemplate := grafana.NewNotificationTemplatesFromFile(
			options.NotificationTemplates,
		)
		if errNotificationTemplate != nil {
			return errNotificationTemplate
		}
		for _, notificationTemplate := range notificationTemplates {
			_, _, errPostNotificationTemplate := grafanaClient.PutNotificationTemplate(notificationTemplate)
			if errPostNotificationTemplate != nil {
				return errPostNotificationTemplate
			}
			Logger.Info().
				Str("Name", *notificationTemplate.Name).
				Str("URL", options.GrafanaURL).
				Msg("Notification template created")
		}
	}

	// Create contact points for the alerts
	if build.ContactPoints != nil && len(build.ContactPoints) > 0 {
		for _, contactPoint := range build.ContactPoints {
			errCreateOrUpdateContactPoint := grafanaClient.CreateOrUpdateContactPoint(contactPoint)
			if errCreateOrUpdateContactPoint != nil {
				return errCreateOrUpdateContactPoint
			}
			Logger.Info().
				Str("Name", *contactPoint.Name).
				Str("URL", options.GrafanaURL).
				Msg("Contact Point created")
		}
	}

	if build.NotificationPolicies != nil && len(build.NotificationPolicies) > 0 {
		for _, notificationPolicy := range build.NotificationPolicies {
			errAddNestedPolicy := grafanaClient.AddNestedPolicy(notificationPolicy)
			if errAddNestedPolicy != nil {
				return errAddNestedPolicy
			}
			Logger.Info().
				Str("Receiver", *notificationPolicy.Receiver).
				Str("URL", options.GrafanaURL).
				Msg("Notification Policy created")
		}
	}

	return nil
}

func DeleteDashboard(options *CommandOptions) error {
	grafanaClient := api.NewClient(
		options.GrafanaURL,
		options.GrafanaToken,
	)

	db, _, errGetDashboard := grafanaClient.GetDashboardByName(options.Name)
	if errGetDashboard != nil {
		return errGetDashboard
	}

	_, errDelete := grafanaClient.DeleteDashboardByUID(*db.UID)
	if errDelete != nil {
		return errDelete
	}

	Logger.Info().
		Str("Name", options.Name).
		Str("URL", options.GrafanaURL).
		Msg("Dashboard deleted")

	return nil
}

func GetDashboardJSON(options *CommandOptions) ([]byte, error) {
	buildOptions := &BuildOptions{
		Name:              options.Name,
		Platform:          grafana.TypePlatform(options.Platform),
		TypeDashboard:     TypeDashboard(options.TypeDashboard),
		MetricsDataSource: grafana.NewDataSource(options.MetricsDataSourceName, ""),
	}

	if options.LogsDataSourceName != "" {
		buildOptions.LogsDataSource = grafana.NewDataSource(options.LogsDataSourceName, "")
	}

	build, errBuild := Build(buildOptions)
	if errBuild != nil {
		return nil, errBuild
	}

	dashboardJSON, err := json.MarshalIndent(build.Dashboard, "", "  ")
	if err != nil {
		return nil, err
	}

	return dashboardJSON, nil
}

func panelIDByTitle(db *dashboard.Dashboard, title string) string {
	for _, panel := range db.Panels {
		if panel.Panel != nil && panel.Panel.Title != nil && *panel.Panel.Title == title {
			return fmt.Sprintf("%d", *panel.Panel.Id)
		}
	}

	return ""
}
