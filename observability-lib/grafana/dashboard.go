package grafana

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"

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

// defaultConcurrency is the default bound on in-flight HTTP calls for
// alert-rule writes when DeployOptions.Concurrency is unset.
const defaultConcurrency = 8

type DeployOptions struct {
	GrafanaURL             string
	GrafanaToken           string
	FolderName             string
	FolderUID              string // when set, deploy to this folder instead of resolving FolderName by title
	EnableAlerts           bool
	RuleGroupFromDashboard bool // if true, set the alert rule group to the dashboard title on all alerts
	NotificationTemplates  string
	// Concurrency bounds in-flight HTTP calls for alert-rule writes (each rule
	// is addressed by UID, so rules deploy independently). 0 uses
	// defaultConcurrency; 1 restores the previous serial behavior.
	Concurrency int
	// Cache, when set, memoizes folder resolution and the full alert-rule
	// fetch across the DeployToGrafana calls sharing it, so composite deploys
	// pay those lookups once instead of once per dashboard. See DeployCache
	// for scoping rules.
	Cache *DeployCache
}

func (o *DeployOptions) concurrency() int {
	if o.Concurrency <= 0 {
		return defaultConcurrency
	}
	return o.Concurrency
}

func resolveDeployFolder(client *api.Client, cache deployCache, options *DeployOptions) (*api.Folder, error) {
	if options.FolderUID != "" {
		key := "uid_" + options.FolderUID
		if folder, ok := cache.folder(key); ok {
			return folder, nil
		}
		folder, err := client.GetFolderByUID(options.FolderUID)
		if err != nil {
			return nil, err
		}
		if folder == nil {
			return nil, fmt.Errorf("folder with UID %q not found", options.FolderUID)
		}
		cache.setFolder(key, folder)
		return folder, nil
	}
	if options.FolderName != "" {
		key := "name_" + options.FolderName
		if folder, ok := cache.folder(key); ok {
			return folder, nil
		}
		folder, err := client.FindOrCreateFolder(options.FolderName)
		if err != nil {
			return nil, err
		}
		cache.setFolder(key, folder)
		return folder, nil
	}
	return nil, nil
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

func getAlertRules(grafanaClient *api.Client, cache deployCache, dashboardUID *string, folderUID string, alertGroups []alerting.RuleGroup) ([]alerting.Rule, error) {
	// Fetch the full rule list exactly once to amortize the cost of
	// fetching alert rules.
	allRules, cached := cache.alertRules()
	if !cached {
		var errGetAlertRules error
		allRules, _, errGetAlertRules = grafanaClient.GetAlertRules()
		if errGetAlertRules != nil {
			return nil, errGetAlertRules
		}
		cache.setAlertRules(allRules)
	}

	var alertsRule []alerting.Rule

	// check for alert rules by dashboard UID
	if dashboardUID != nil {
		for _, rule := range allRules {
			if rule.Annotations["__dashboardUid__"] == *dashboardUID {
				alertsRule = append(alertsRule, rule)
			}
		}
	}

	// check for alert rules by folder UID and group name
	if len(alertGroups) > 0 {
		groupNames := make(map[string]bool, len(alertGroups))
		for _, alertGroup := range alertGroups {
			if alertGroup.Title != nil {
				groupNames[*alertGroup.Title] = true
			}
		}
		for _, rule := range allRules {
			if rule.FolderUID != "" && rule.FolderUID == folderUID && groupNames[rule.RuleGroup] {
				alertsRule = append(alertsRule, rule)
			}
		}
	}

	return alertsRule, nil
}

func (o *Observability) DeployToGrafana(options *DeployOptions) error {
	grafanaClient := api.NewClient(
		options.GrafanaURL,
		options.GrafanaToken,
	)

	// Substitute a no-op cache when the caller didn't configure one, so the
	// rest of the deploy path never handles a nil cache.
	cache := deployCache(noopDeployCache{})
	if options.Cache != nil {
		cache = options.Cache
	}

	// Create or update folder
	folder, errFolder := resolveDeployFolder(grafanaClient, cache, options)
	if errFolder != nil {
		return errFolder
	}

	// Create or update dashboard
	var newDashboard api.PostDashboardResponse
	var errPostDashboard error

	if o.Dashboard != nil && folder != nil {
		dashboardFound, _, err := grafanaClient.GetDashboardByNameFolderUID(*o.Dashboard.Title, folder.UID)
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
				FolderUID: folder.UID,
			})
			if errPostDashboard != nil {
				return errPostDashboard
			}
			if newDashboard.URL != nil && o.Dashboard.Title != nil {
				fmt.Fprintf(os.Stderr, "deployed dashboard %q to folder uid=%s: %s\n", *o.Dashboard.Title, folder.UID, *newDashboard.URL)
			}
		}
	}

	// If disabling alerts delete alerts for the folder and alert groups scope
	if folder != nil && !options.EnableAlerts && o.Alerts != nil && len(o.Alerts) > 0 {
		alertsRule, errGetAlertRules := getAlertRules(grafanaClient, cache, newDashboard.UID, folder.UID, o.AlertGroups)
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
		alertsRule, errGetAlertRules := getAlertRules(grafanaClient, cache, newDashboard.UID, folder.UID, o.AlertGroups)
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
		errUpsertAlerts := parallelFor(o.Alerts, options.concurrency(), func(alert alerting.Rule) error {
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
					_, updateResp, errPutAlertRule := grafanaClient.UpdateAlertRule(*alertToUpdate.Uid, alert)
					if errPutAlertRule != nil {
						// A 409 means a provenance mismatch: the stored rule was created with a
						// different provenance (e.g. "api") than what we send (""). Migrate by
						// deleting the old rule and recreating it with the current provenance.
						if updateResp == nil || updateResp.StatusCode() != 409 {
							return errPutAlertRule
						}

						if _, _, errDelete := grafanaClient.DeleteAlertRule(*alertToUpdate.Uid); errDelete != nil {
							return fmt.Errorf("provenance migration: delete alert rule %q: %w", alert.Title, errDelete)
						}
						if _, _, errPost := grafanaClient.PostAlertRule(alert); errPost != nil {
							return fmt.Errorf("provenance migration: recreate alert rule %q: %w", alert.Title, errPost)
						}
					}
				}
			} else {
				// create alert rule if it doesn't exist
				_, _, errPostAlertRule := grafanaClient.PostAlertRule(alert)
				if errPostAlertRule != nil {
					return errPostAlertRule
				}
			}
			return nil
		})
		if errUpsertAlerts != nil {
			return errUpsertAlerts
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
	if id, ok := PanelIDByTitle(db, title); ok {
		return strconv.FormatUint(uint64(id), 10)
	}
	return ""
}

type DeleteOptions struct {
	Name         string
	GrafanaURL   string
	GrafanaToken string
	FolderUID    string
}

func DeleteDashboard(options *DeleteOptions) error {
	grafanaClient := api.NewClient(
		options.GrafanaURL,
		options.GrafanaToken,
	)

	db, _, errGetDashboard := grafanaClient.GetDashboardByNameFolderUID(options.Name, options.FolderUID)
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
