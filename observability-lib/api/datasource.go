package api

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type Datasource struct {
	ID   uint   `json:"id"`
	UID  string `json:"uid"`
	Name string `json:"name"`
}

// GetDataSourceByName Get a datasource by name
func (c *Client) GetDataSourceByName(name string) (*Datasource, *resty.Response, error) {
	var grafanaResp Datasource

	resp, err := c.resty.R().
		SetHeader("Accept", "application/json").
		SetResult(&grafanaResp).
		Get(fmt.Sprintf("/api/datasources/name/%s", name))

	if err != nil {
		return nil, resp, err
	}

	statusCode := resp.StatusCode()
	if statusCode != 200 {
		return nil, resp, fmt.Errorf("error fetching datasource, received unexpected status code %d: %s", statusCode, resp.String())
	}
	return &grafanaResp, resp, nil
}
