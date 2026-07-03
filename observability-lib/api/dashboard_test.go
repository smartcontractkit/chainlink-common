package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDashboardPanelsByUID(t *testing.T) {
	dashboardUID := "dashboard-uid"
	expectedResponse := getDashboardByUIDResponse{
		Dashboard: struct {
			ID     uint                 `json:"id"`
			UID    string               `json:"uid"`
			Title  string               `json:"title"`
			Panels []dashboardPanelJSON `json:"panels"`
		}{
			ID:    10,
			UID:   dashboardUID,
			Title: "Test Dashboard",
			Panels: []dashboardPanelJSON{
				{ID: 1, Title: "Uptime", Type: "stat"},
				{
					ID:    2,
					Title: "Resource Usage",
					Type:  "row",
					Panels: []dashboardPanelJSON{
						{ID: 3, Title: "CPU Usage", Type: "timeseries"},
						{ID: 4, Title: "Memory Usage", Type: "stat"},
					},
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/dashboards/uid/"+dashboardUID, r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(expectedResponse); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer ts.Close()

	client := &Client{
		resty: resty.New().SetBaseURL(ts.URL),
	}

	panels, resp, err := client.GetDashboardPanelsByUID(dashboardUID)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, []DashboardPanel{
		{ID: 1, Title: "Uptime", Type: "stat"},
		{ID: 2, Title: "Resource Usage", Type: "row"},
		{ID: 3, Title: "CPU Usage", Type: "timeseries"},
		{ID: 4, Title: "Memory Usage", Type: "stat"},
	}, panels)
}

func TestGetDashboardPanelsByUID_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := &Client{
		resty: resty.New().SetBaseURL(ts.URL),
	}

	panels, resp, err := client.GetDashboardPanelsByUID("missing-dashboard")

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
	assert.Nil(t, panels)
}
