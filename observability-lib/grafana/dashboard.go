package grafana

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
)

type TypePlatform string

const (
	TypePlatformKubernetes TypePlatform = "kubernetes"
	TypePlatformDocker     TypePlatform = "docker"
)

type Observability struct {
	Dashboard            *dashboard.Dashboard
	Alerts               []alerting.Rule
	ContactPoints        []alerting.ContactPoint
	NotificationPolicies []alerting.NotificationPolicy
}

func (o *Observability) GenerateJSON() ([]byte, error) {
	output, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return nil, err
	}

	return output, nil
}

type DeployOptions struct {
	GrafanaURL            string
	GrafanaToken          string
	FolderName            string
	EnableAlerts          bool
	NotificationTemplates string
}

func alertRuleExist(alerts []alerting.Rule, alert alerting.Rule) bool {
	for _, a := range alerts {
		if reflect.DeepEqual(a.Title, alert.Title) {
			return true
		}
	}
	return false
}

func getAlertRuleByTitle(alerts []alerting.Rule, title string) *alerting.Rule {
	for _, a := range alerts {
		if a.Title == title {
			return &a
		}
	}
	return nil
}

func (o *Observability) DeployToGrafana(options *DeployOptions) error {
	grafanaClient := api.NewClient(
		options.GrafanaURL,
		options.GrafanaToken,
	)

	if options.FolderName != "" {
		folder, errFolder := grafanaClient.FindOrCreateFolder(options.FolderName)
		if errFolder != nil {
			return errFolder
		}
		newDashboard, _, errPostDashboard := grafanaClient.PostDashboard(api.PostDashboardRequest{
			Dashboard: o.Dashboard,
			Overwrite: true,
			FolderID:  int(folder.ID),
		})
		if errPostDashboard != nil {
			return errPostDashboard
		}

		if !options.EnableAlerts && o.Alerts != nil && len(o.Alerts) > 0 {
			// Get alert rules for the dashboard
			alertsRule, errGetAlertRules := grafanaClient.GetAlertRulesByDashboardUID(*newDashboard.UID)
			if errGetAlertRules != nil {
				return errGetAlertRules
			}

			// delete existing alert rules for the dashboard if alerts are disabled
			for _, rule := range alertsRule {
				_, _, errDeleteAlertRule := grafanaClient.DeleteAlertRule(*rule.Uid)
				if errDeleteAlertRule != nil {
					return errDeleteAlertRule
				}
			}
		}

		// Create alerts for the dashboard
		if options.EnableAlerts && o.Alerts != nil && len(o.Alerts) > 0 {
			// Get alert rules for the dashboard
			alertsRule, errGetAlertRules := grafanaClient.GetAlertRulesByDashboardUID(*newDashboard.UID)
			if errGetAlertRules != nil {
				return errGetAlertRules
			}

			// delete alert rules for the dashboard
			for _, rule := range alertsRule {
				// delete alert rule only if it won't be created again from code
				if !alertRuleExist(o.Alerts, rule) {
					_, _, errDeleteAlertRule := grafanaClient.DeleteAlertRule(*rule.Uid)
					if errDeleteAlertRule != nil {
						return errDeleteAlertRule
					}
				}
			}

			// Create alert rules for the dashboard
			for _, alert := range o.Alerts {
				alert.RuleGroup = *o.Dashboard.Title
				alert.FolderUID = folder.UID
				alert.Annotations["__dashboardUid__"] = *newDashboard.UID

				panelId := panelIDByTitle(o.Dashboard, alert.Annotations["panel_title"])
				// we can clean it up as it was only used to get the panelId
				delete(alert.Annotations, "panel_title")
				if panelId != "" {
					alert.Annotations["__panelId__"] = panelId
				}
				if alertRuleExist(alertsRule, alert) {
					// update alert rule if it already exists
					alertToUpdate := getAlertRuleByTitle(alertsRule, alert.Title)
					if alertToUpdate != nil {
						_, _, errPutAlertRule := grafanaClient.UpdateAlertRule(*alertToUpdate.Uid, alert)
						if errPutAlertRule != nil {
							return errPutAlertRule
						}
					}
				} else {
					// create alert rule if it doesn't exist
					_, _, errPostAlertRule := grafanaClient.PostAlertRule(alert)
					if errPostAlertRule != nil {
						return errPostAlertRule
					}
				}
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
	if o.ContactPoints != nil && len(o.ContactPoints) > 0 {
		for _, contactPoint := range o.ContactPoints {
			errCreateOrUpdateContactPoint := grafanaClient.CreateOrUpdateContactPoint(contactPoint)
			if errCreateOrUpdateContactPoint != nil {
				return errCreateOrUpdateContactPoint
			}
		}
	}

	// Create notification policies for the alerts
	if o.NotificationPolicies != nil && len(o.NotificationPolicies) > 0 {
		for _, notificationPolicy := range o.NotificationPolicies {
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

	alertsRule, errGetAlertRules := grafanaClient.GetAlertRulesByDashboardUID(*db.UID)
	if errGetAlertRules != nil {
		return errGetAlertRules
	}

	// delete existing alert rules for the dashboard if alerts are disabled
	for _, rule := range alertsRule {
		_, _, errDeleteAlertRule := grafanaClient.DeleteAlertRule(*rule.Uid)
		if errDeleteAlertRule != nil {
			return errDeleteAlertRule
		}
	}

	_, errDelete := grafanaClient.DeleteDashboardByUID(*db.UID)
	if errDelete != nil {
		return errDelete
	}

	return nil
}
