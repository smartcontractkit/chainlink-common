package api

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
)

// CreateOrUpdateContactPoint create or update a contact point
func (c *Client) CreateOrUpdateContactPoint(contactPoint alerting.ContactPoint) error {
	cPoint, err := c.GetContactPointByName(*contactPoint.Name)
	if err != nil {
		return fmt.Errorf("could not create or update contact point: %w", err)
	}
	if cPoint == nil {
		_, _, err = c.PostContactPoint(contactPoint)
		if err != nil {
			return fmt.Errorf("could not create or update contact point: %w", err)
		}
	} else {
		_, _, err = c.PutContactPoint(*cPoint.Uid, contactPoint)
		if err != nil {
			return fmt.Errorf("could not create or update contact point: %w", err)
		}
	}

	return nil
}

// GetContactPointByName Get a contact point by name
func (c *Client) GetContactPointByName(name string) (*alerting.ContactPoint, error) {
	contactPoints, _, err := c.GetContactPoints()
	if err != nil {
		return nil, err
	}
	for _, contactPoint := range contactPoints {
		if *contactPoint.Name == name {
			return &contactPoint, nil
		}
	}

	return nil, nil
}

type GetContactPointsResponse []alerting.ContactPoint

// GetContactPoints Get all the contact points
func (c *Client) GetContactPoints() (GetContactPointsResponse, *resty.Response, error) {
	var grafanaResp GetContactPointsResponse

	resp, err := c.resty.R().
		SetHeader("Accept", "application/json").
		SetResult(&grafanaResp).
		Get("/api/v1/provisioning/contact-points")

	if err != nil {
		return GetContactPointsResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 {
		return GetContactPointsResponse{}, resp, fmt.Errorf("error fetching contact points, received unexpected status code %d: %s", statusCode, resp.String())
	}
	return grafanaResp, resp, nil
}

type DeleteContactPointResponse struct{}

// DeleteContactPoint Delete a contact point
func (c *Client) DeleteContactPoint(uid string) (DeleteContactPointResponse, *resty.Response, error) {
	var grafanaResp DeleteContactPointResponse

	resp, err := c.resty.R().
		SetResult(&grafanaResp).
		Delete(fmt.Sprintf("/api/v1/provisioning/contact-points/%s", uid))

	if err != nil {
		return DeleteContactPointResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 204 {
		return DeleteContactPointResponse{}, resp, fmt.Errorf("error deleting contact point, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return grafanaResp, resp, nil
}

type PostContactPointResponse struct{}

// PostContactPoint Create a new contact point
func (c *Client) PostContactPoint(contactPoint alerting.ContactPoint) (PostContactPointResponse, *resty.Response, error) {
	var grafanaResp PostContactPointResponse

	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(contactPoint).
		SetResult(&grafanaResp).
		Post("/api/v1/provisioning/contact-points")

	if err != nil {
		return PostContactPointResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 202 {
		return PostContactPointResponse{}, resp, fmt.Errorf("error creating contact point, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return grafanaResp, resp, nil
}

type PutContactPointResponse struct{}

// PutContactPoint Update an existing contact point
func (c *Client) PutContactPoint(uid string, contactPoint alerting.ContactPoint) (PutContactPointResponse, *resty.Response, error) {
	var grafanaResp PutContactPointResponse

	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(contactPoint).
		SetResult(&grafanaResp).
		Put(fmt.Sprintf("/api/v1/provisioning/contact-points/%s", uid))

	if err != nil {
		return PutContactPointResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 202 {
		return PutContactPointResponse{}, resp, fmt.Errorf("error updating contact point, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return grafanaResp, resp, nil
}
