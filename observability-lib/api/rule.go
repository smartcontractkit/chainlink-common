package api

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
)

type GetAllAlertRulesResponse []alerting.Rule

// GetAlertRulesByDashboardUID Get alert rules by dashboard UID
func (c *Client) GetAlertRulesByDashboardUID(dashboardUID string) (GetAllAlertRulesResponse, error) {
	var alerts []alerting.Rule

	alertsRule, _, err := c.GetAlertRules()
	if err != nil {
		return nil, err
	}
	for _, rule := range alertsRule {
		if rule.Annotations["__dashboardUid__"] == dashboardUID {
			alerts = append(alerts, rule)
		}
	}
	return alerts, nil
}

// GetAlertRules Get all alert rules
func (c *Client) GetAlertRules() (GetAllAlertRulesResponse, *resty.Response, error) {
	var grafanaResp GetAllAlertRulesResponse

	resp, err := c.resty.R().
		SetHeader("Accept", "application/json").
		SetResult(&grafanaResp).
		Get("/api/v1/provisioning/alert-rules")

	if err != nil {
		return GetAllAlertRulesResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 {
		return GetAllAlertRulesResponse{}, resp, fmt.Errorf("error fetching alert rules, received unexpected status code %d: %s", statusCode, resp.String())
	}
	return grafanaResp, resp, nil
}

type PostAlertRuleResponse struct{}

// PostAlertRule Create a new alert rule
func (c *Client) PostAlertRule(alertRule alerting.Rule) (PostAlertRuleResponse, *resty.Response, error) {
	var grafanaResp PostAlertRuleResponse

	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Disable-Provenance", "true").
		SetBody(alertRule).
		SetResult(&grafanaResp).
		Post("/api/v1/provisioning/alert-rules")

	if err != nil {
		return PostAlertRuleResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 201 {
		return PostAlertRuleResponse{}, resp, fmt.Errorf("error creating alert rule, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return grafanaResp, resp, nil
}

type DeleteAlertRuleResponse struct{}

// DeleteAlertRule Delete a specific alert rule by UID
func (c *Client) DeleteAlertRule(uid string) (DeleteAlertRuleResponse, *resty.Response, error) {
	var grafanaResp DeleteAlertRuleResponse

	resp, err := c.resty.R().
		SetResult(&grafanaResp).
		Delete(fmt.Sprintf("/api/v1/provisioning/alert-rules/%s", uid))

	if err != nil {
		return DeleteAlertRuleResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 204 {
		return DeleteAlertRuleResponse{}, resp, fmt.Errorf("error deleting alert rule, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return grafanaResp, resp, nil
}
