package api

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
)

type GetNotificationPolicyResponse alerting.NotificationPolicy

// GetNotificationPolicy Get the notification policy tree
func (c *Client) GetNotificationPolicy() (GetNotificationPolicyResponse, *resty.Response, error) {
	var grafanaResp GetNotificationPolicyResponse

	resp, err := c.resty.R().
		SetHeader("Accept", "application/json").
		SetResult(&grafanaResp).
		Get("/api/v1/provisioning/policies")

	if err != nil {
		return GetNotificationPolicyResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 {
		return GetNotificationPolicyResponse{}, resp, fmt.Errorf("error fetching notification policy tree, received unexpected status code %d: %s", statusCode, resp.String())
	}
	return grafanaResp, resp, nil
}

type DeleteNotificationPolicyResponse struct{}

// DeleteNotificationPolicy Clears the notification policy tree
func (c *Client) DeleteNotificationPolicy() (DeleteNotificationPolicyResponse, *resty.Response, error) {
	var grafanaResp DeleteNotificationPolicyResponse

	resp, err := c.resty.R().
		SetResult(&grafanaResp).
		Delete("/api/v1/provisioning/policies")

	if err != nil {
		return DeleteNotificationPolicyResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 202 {
		return DeleteNotificationPolicyResponse{}, resp, fmt.Errorf("error deleting notification policy tree, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return grafanaResp, resp, nil
}

type PutNotificationPolicyResponse struct{}

// PutNotificationPolicy Sets the notification policy tree
func (c *Client) PutNotificationPolicy(notificationPolicy alerting.NotificationPolicy) (PutNotificationPolicyResponse, *resty.Response, error) {
	var grafanaResp PutNotificationPolicyResponse

	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(notificationPolicy).
		SetResult(&grafanaResp).
		Put("/api/v1/provisioning/policies")

	if err != nil {
		return PutNotificationPolicyResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 202 {
		return PutNotificationPolicyResponse{}, resp, fmt.Errorf("error setting notification policy tree, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return grafanaResp, resp, nil
}
