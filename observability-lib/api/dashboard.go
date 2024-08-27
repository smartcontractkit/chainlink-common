package api

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

type GetDashboardResponse struct {
	ID    *uint   `json:"id"`
	UID   *string `json:"uid"`
	Title *string `json:"title"`
}

func (c *Client) GetDashboardByName(name string) (GetDashboardResponse, *resty.Response, error) {
	var grafanaResp []GetDashboardResponse

	resp, err := c.resty.R().
		SetResult(&grafanaResp).
		SetQueryParam("type", "dash-db").
		SetQueryParam("query", name).
		Get("/api/search")

	if err != nil {
		return GetDashboardResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 && statusCode != 201 {
		return GetDashboardResponse{}, resp, fmt.Errorf("error getting dashboard, received unexpected status code %d: %s", statusCode, resp.String())
	}

	if len(grafanaResp) > 0 {
		for _, dashboard := range grafanaResp {
			if strings.EqualFold(*dashboard.Title, name) {
				return dashboard, resp, nil
			}
		}
	}

	return GetDashboardResponse{}, resp, nil
}

type PostDashboardRequest struct {
	Dashboard interface{} `json:"dashboard"`
	FolderID  int         `json:"folderId"`
	Overwrite bool        `json:"overwrite"`
}

type PostDashboardResponse struct {
	ID      *uint   `json:"id"`
	OrgID   *uint   `json:"orgId"`
	Message *string `json:"message"`
	Slug    *string `json:"slug"`
	Version *int    `json:"version"`
	Status  *string `json:"status"`
	UID     *string `json:"uid"`
	URL     *string `json:"url"`
}

func (c *Client) PostDashboard(dashboard PostDashboardRequest) (PostDashboardResponse, *resty.Response, error) {
	var grafanaResp PostDashboardResponse

	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(dashboard).
		SetResult(&grafanaResp).
		Post("/api/dashboards/db")

	if err != nil {
		return PostDashboardResponse{}, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 && statusCode != 201 {
		return PostDashboardResponse{}, resp, fmt.Errorf("error creating/updating dashboard, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return grafanaResp, resp, nil
}

func (c *Client) DeleteDashboardByUID(uid string) (*resty.Response, error) {
	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		Delete(fmt.Sprintf("/api/dashboards/uid/%s", uid))

	if err != nil {
		return resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 && statusCode != 201 {
		return resp, fmt.Errorf("error deleting dashboard, received unexpected status code %d: %s", statusCode, resp.String())
	}

	return resp, nil
}
