package api

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
)

type UpdateAlertRuleGroupResponse struct{}

// UpdateAlertRuleGroup Update a specific alert rule group
func (c *Client) UpdateAlertRuleGroup(folderUID string, alertRuleGroup alerting.RuleGroup) (UpdateAlertRuleGroupResponse, *resty.Response, error) {
	var grafanaResp UpdateAlertRuleGroupResponse

	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Disable-Provenance", "true").
		SetBody(alertRuleGroup).
		SetResult(&grafanaResp).
		Put(fmt.Sprintf("/api/v1/provisioning/folder/%s/rule-groups/%s", folderUID, *alertRuleGroup.Title))

	if err != nil {
		return UpdateAlertRuleGroupResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 {
		return UpdateAlertRuleGroupResponse{}, resp, fmt.Errorf("error updating alert rule group, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return grafanaResp, resp, nil
}
