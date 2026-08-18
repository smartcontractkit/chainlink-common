package api

import (
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
)

type Datasource struct {
	ID   uint   `json:"id"`
	UID  string `json:"uid"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetDataSourceByName Get a datasource by name
func (c *Client) GetDataSourceByName(name string) (*Datasource, *resty.Response, error) {
	var grafanaResp Datasource

	// URL-encode the name to handle special characters and spaces
	escapedName := url.PathEscape(name)

	resp, err := c.resty.R().
		SetHeader("Accept", "application/json").
		SetResult(&grafanaResp).
		Get(fmt.Sprintf("/api/datasources/name/%s", escapedName))

	if err != nil {
		return nil, resp, err
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 {
		return nil, resp, fmt.Errorf("error fetching datasource %s, received unexpected status code %d: %s", name, statusCode, resp.String())
	}
	return &grafanaResp, resp, nil
}
