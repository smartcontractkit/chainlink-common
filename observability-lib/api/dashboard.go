package api

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

type GetDashboardResponse struct {
	ID          *uint   `json:"id"`
	UID         *string `json:"uid"`
	Title       *string `json:"title"`
	FolderTitle *string `json:"folderTitle"`
	FolderUID   *string `json:"folderUid"`
	FolderID    *uint   `json:"folderId"`
}

func (c *Client) GetDashboardByNameFolderUID(name string, folderUID string) (GetDashboardResponse, *resty.Response, error) {
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
			if dashboard.Title == nil || dashboard.FolderUID == nil || dashboard.UID == nil {
				continue
			}
			if strings.EqualFold(*dashboard.Title, name) && strings.EqualFold(*dashboard.FolderUID, folderUID) {
				return dashboard, resp, nil
			}
		}
	}

	return GetDashboardResponse{}, resp, nil
}

type PostDashboardRequest struct {
	Dashboard interface{} `json:"dashboard"`
	FolderID  int         `json:"folderId,omitempty"`
	FolderUID string      `json:"folderUid,omitempty"`
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

type DashboardPanel struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

type dashboardPanelJSON struct {
	ID     int                  `json:"id"`
	Title  string               `json:"title"`
	Type   string               `json:"type"`
	Panels []dashboardPanelJSON `json:"panels"`
}

type getDashboardByUIDResponse struct {
	Dashboard struct {
		ID     uint                 `json:"id"`
		UID    string               `json:"uid"`
		Title  string               `json:"title"`
		Panels []dashboardPanelJSON `json:"panels"`
	} `json:"dashboard"`
}

func collectDashboardPanels(panels []dashboardPanelJSON) []DashboardPanel {
	result := make([]DashboardPanel, 0, len(panels))
	for _, panel := range panels {
		result = append(result, DashboardPanel{
			ID:    panel.ID,
			Title: panel.Title,
			Type:  panel.Type,
		})
		if len(panel.Panels) > 0 {
			result = append(result, collectDashboardPanels(panel.Panels)...)
		}
	}
	return result
}

// GetDashboardPanelsByUID returns all panels (including nested row panels) with their IDs and titles.
func (c *Client) GetDashboardPanelsByUID(uid string) ([]DashboardPanel, *resty.Response, error) {
	var grafanaResp getDashboardByUIDResponse

	resp, err := c.resty.R().
		SetHeader("Accept", "application/json").
		SetResult(&grafanaResp).
		Get(fmt.Sprintf("/api/dashboards/uid/%s", uid))

	if err != nil {
		return nil, resp, fmt.Errorf("error making API request: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode == 404 {
		return nil, resp, nil
	}
	if statusCode != 200 {
		return nil, resp, fmt.Errorf("error fetching dashboard %q, received unexpected status code %d: %s", uid, statusCode, resp.String())
	}

	return collectDashboardPanels(grafanaResp.Dashboard.Panels), resp, nil
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
