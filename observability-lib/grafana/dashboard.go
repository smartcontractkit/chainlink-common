package grafana

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
)

type Observability struct {
	Dashboard            *dashboard.Dashboard
	Alerts               []alerting.Rule
	AlertGroups          []alerting.RuleGroup
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
	GrafanaURL             string
	GrafanaToken           string
	FolderName             string
	EnableAlerts           bool
	RuleGroupFromDashboard bool // if true, set the alert rule group to the dashboard title on all alerts
	NotificationTemplates  string
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

func getAlertRules(grafanaClient *api.Client, dashboardUID *string, folderUID string, alertGroups []alerting.RuleGroup) ([]alerting.Rule, error) {
	var alertsRule []alerting.Rule
	var errGetAlertRules error

	// check for alert rules by dashboard UID
	if dashboardUID != nil {
		alertsRule, errGetAlertRules = grafanaClient.GetAlertRulesByDashboardUID(*dashboardUID)
		if errGetAlertRules != nil {
			return nil, errGetAlertRules
		}
	}

	// check for alert rules by folder UID and group name
	if len(alertGroups) > 0 {
		for _, alertGroup := range alertGroups {
			alertsRulePerGroup, errGetAlertRulesPerGroup := grafanaClient.GetAlertRulesByFolderUIDAndGroupName(folderUID, *alertGroup.Title)
			if errGetAlertRulesPerGroup != nil {
				return nil, errGetAlertRulesPerGroup
			}
			alertsRule = append(alertsRule, alertsRulePerGroup...)
		}
	}

	return alertsRule, nil
}

func (o *Observability) DeployToGrafana(options *DeployOptions) error {
	grafanaClient := api.NewClient(
		options.GrafanaURL,
		options.GrafanaToken,
	)

	// Create or update folder
	var folder *api.Folder
	var errFolder error
	if options.FolderName != "" {
		folder, errFolder = grafanaClient.FindOrCreateFolder(options.FolderName)
		if errFolder != nil {
			return errFolder
		}
	}

	// Create or update dashboard
	var newDashboard api.PostDashboardResponse
	var errPostDashboard error

	dashboardFound, _, err := grafanaClient.GetDashboardByName(*o.Dashboard.Title)
	if err != nil {
		return err
	}
	if dashboardFound.UID != nil {
		if o.Dashboard.Uid == nil {
			o.Dashboard.Uid = dashboardFound.UID
		}
	}

	if folder != nil && o.Dashboard != nil {
		newDashboard, _, errPostDashboard = grafanaClient.PostDashboard(api.PostDashboardRequest{
			Dashboard: o.Dashboard,
			Overwrite: true,
			FolderID:  int(folder.ID),
		})
		if errPostDashboard != nil {
			return errPostDashboard
		}
	}

	// If disabling alerts delete alerts for the folder and alert groups scope
	if folder != nil && !options.EnableAlerts && o.Alerts != nil && len(o.Alerts) > 0 {
		alertsRule, errGetAlertRules := getAlertRules(grafanaClient, newDashboard.UID, folder.UID, o.AlertGroups)
		if errGetAlertRules != nil {
			return errGetAlertRules
		}

		for _, rule := range alertsRule {
			_, _, errDeleteAlertRule := grafanaClient.DeleteAlertRule(*rule.Uid)
			if errDeleteAlertRule != nil {
				return errDeleteAlertRule
			}
		}
	}

	// Create or update alerts
	if folder != nil && options.EnableAlerts && o.Alerts != nil && len(o.Alerts) > 0 {
		alertsRule, errGetAlertRules := getAlertRules(grafanaClient, newDashboard.UID, folder.UID, o.AlertGroups)
		if errGetAlertRules != nil {
			return errGetAlertRules
		}

		// delete alert rules that are not defined anymore in the code
		for _, rule := range alertsRule {
			if !alertRuleExist(o.Alerts, rule) {
				_, _, errDeleteAlertRule := grafanaClient.DeleteAlertRule(*rule.Uid)
				if errDeleteAlertRule != nil {
					return errDeleteAlertRule
				}
			}
		}

		// Create alert rules
		for _, alert := range o.Alerts {
			if folder.UID != "" {
				alert.FolderUID = folder.UID
			}
			if o.Dashboard != nil {
				if options.RuleGroupFromDashboard {
					alert.RuleGroup = *o.Dashboard.Title
				}
				if alert.Annotations["panel_title"] != "" {
					panelId := panelIDByTitle(o.Dashboard, alert.Annotations["panel_title"])
					// we can clean it up as it was only used to get the panelId
					delete(alert.Annotations, "panel_title")
					if panelId != "" {
						// Both or none should be set
						alert.Annotations["__panelId__"] = panelId
						alert.Annotations["__dashboardUid__"] = *newDashboard.UID
					}
				}
			} else {
				if alert.RuleGroup == "" {
					return fmt.Errorf("you must create at one rule group and set it to your alerts")
				}
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

	// Update alert groups
	if folder != nil {
		for _, alertGroup := range o.AlertGroups {
			_, _, errPostAlertGroup := grafanaClient.UpdateAlertRuleGroup(folder.UID, alertGroup)
			if errPostAlertGroup != nil {
				return errPostAlertGroup
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
	if len(o.ContactPoints) > 0 {
		for _, contactPoint := range o.ContactPoints {
			errCreateOrUpdateContactPoint := grafanaClient.CreateOrUpdateContactPoint(contactPoint)
			if errCreateOrUpdateContactPoint != nil {
				return errCreateOrUpdateContactPoint
			}
		}
	}

	// Create notification policies for the alerts
	if len(o.NotificationPolicies) > 0 {
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
