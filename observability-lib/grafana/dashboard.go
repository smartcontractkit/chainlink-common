package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
)

type TypePlatform string

const (
	TypePlatformKubernetes TypePlatform = "kubernetes"
	TypePlatformDocker     TypePlatform = "docker"
)

type Dashboard struct {
	Dashboard            *dashboard.Dashboard
	Alerts               []alerting.Rule
	ContactPoints        []alerting.ContactPoint
	NotificationPolicies []alerting.NotificationPolicy
}

func (db *Dashboard) GenerateJSON() ([]byte, error) {
	dashboardJSON, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return nil, err
	}

	return dashboardJSON, nil
}

type DeployOptions struct {
	GrafanaURL            string
	GrafanaToken          string
	FolderName            string
	EnableAlerts          bool
	NotificationTemplates string
}

func (db *Dashboard) DeployToGrafana(options *DeployOptions) error {
	grafanaClient := api.NewClient(
		options.GrafanaURL,
		options.GrafanaToken,
	)

	folder, errFolder := grafanaClient.FindOrCreateFolder(options.FolderName)
	if errFolder != nil {
		return errFolder
	}

	newDashboard, _, errPostDashboard := grafanaClient.PostDashboard(api.PostDashboardRequest{
		Dashboard: db.Dashboard,
		Overwrite: true,
		FolderID:  int(folder.ID),
	})
	if errPostDashboard != nil {
		return errPostDashboard
	}

	// Create alerts for the dashboard
	if options.EnableAlerts && db.Alerts != nil && len(db.Alerts) > 0 {
		// Get alert rules for the dashboard
		alertsRule, errGetAlertRules := grafanaClient.GetAlertRulesByDashboardUID(*newDashboard.UID)
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
		for _, alert := range db.Alerts {
			alert.RuleGroup = *db.Dashboard.Title
			alert.FolderUID = folder.UID
			alert.Annotations["__dashboardUid__"] = *newDashboard.UID
			alert.Annotations["__panelId__"] = panelIDByTitle(db.Dashboard, alert.Title)

			_, _, errPostAlertRule := grafanaClient.PostAlertRule(alert)
			if errPostAlertRule != nil {
				return errPostAlertRule
			}
		}
	}

	// Create notification templates for the alerts
	if options.NotificationTemplates != "" {
		notificationTemplates, errNotificationTemplate := NewNotificationTemplatesFromFile(
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
		}
	}

	// Create contact points for the alerts
	if db.ContactPoints != nil && len(db.ContactPoints) > 0 {
		for _, contactPoint := range db.ContactPoints {
			errCreateOrUpdateContactPoint := grafanaClient.CreateOrUpdateContactPoint(contactPoint)
			if errCreateOrUpdateContactPoint != nil {
				return errCreateOrUpdateContactPoint
			}
		}
	}

	// Create notification policies for the alerts
	if db.NotificationPolicies != nil && len(db.NotificationPolicies) > 0 {
		for _, notificationPolicy := range db.NotificationPolicies {
			errAddNestedPolicy := grafanaClient.AddNestedPolicy(notificationPolicy)
			if errAddNestedPolicy != nil {
				return errAddNestedPolicy
			}
		}
	}

	return nil
}

func panelIDByTitle(db *dashboard.Dashboard, title string) string {
	for _, panel := range db.Panels {
		if panel.Panel != nil && panel.Panel.Title != nil && *panel.Panel.Title == title {
			return fmt.Sprintf("%d", *panel.Panel.Id)
		}
	}

	return ""
}

type DeleteOptions struct {
	Name         string
	GrafanaURL   string
	GrafanaToken string
}

func DeleteDashboard(options *DeleteOptions) error {
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

	return nil
}
