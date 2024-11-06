package api

import (
	"fmt"
	"reflect"

	"github.com/go-resty/resty/v2"
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
)

func objectMatchersEqual(a alerting.ObjectMatchers, b alerting.ObjectMatchers) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		foundMatch := false
		for j := range b {
			if reflect.DeepEqual(a[i], b[j]) {
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			return false
		}
	}

	return true
}

func policyExist(parent alerting.NotificationPolicy, newNotificationPolicy alerting.NotificationPolicy) bool {
	for _, notificationPolicy := range parent.Routes {
		matchersEqual := false
		if notificationPolicy.ObjectMatchers != nil {
			matchersEqual = objectMatchersEqual(*notificationPolicy.ObjectMatchers, *newNotificationPolicy.ObjectMatchers)
		}
		receiversEqual := reflect.DeepEqual(notificationPolicy.Receiver, newNotificationPolicy.Receiver)
		if matchersEqual && receiversEqual {
			return true
		}
		if notificationPolicy.Routes != nil {
			policyExist(notificationPolicy, newNotificationPolicy)
		}
	}
	return false
}

// AddNestedPolicy Add Nested Policy to Notification Policy Tree
func (c *Client) AddNestedPolicy(newNotificationPolicy alerting.NotificationPolicy) error {
	notificationPolicyTree, _, err := c.GetNotificationPolicy()
	if err != nil {
		return err
	}
	if !policyExist(alerting.NotificationPolicy(notificationPolicyTree), newNotificationPolicy) {
		notificationPolicyTree.Routes = append(notificationPolicyTree.Routes, newNotificationPolicy)
		_, _, errPutNotificationPolicy := c.PutNotificationPolicy(alerting.NotificationPolicy(notificationPolicyTree))
		if errPutNotificationPolicy != nil {
			return errPutNotificationPolicy
		}
	}
	return nil
}

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
		SetHeader("X-Disable-Provenance", "true").
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
