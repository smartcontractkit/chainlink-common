package api

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
)

type PutNotificationTemplateResponse struct{}

// PutNotificationTemplate Create or update a notification template
func (c *Client) PutNotificationTemplate(notificationTemplate alerting.NotificationTemplate) (PutNotificationPolicyResponse, *resty.Response, error) {
	var grafanaResp PutNotificationPolicyResponse

	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(notificationTemplate).
		SetResult(&grafanaResp).
		Put(fmt.Sprintf("/api/v1/provisioning/templates/%s", *notificationTemplate.Name))

	if err != nil {
		return PutNotificationPolicyResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 202 {
		return PutNotificationPolicyResponse{}, resp, fmt.Errorf("error creating/updating notification template, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return grafanaResp, resp, nil
}
