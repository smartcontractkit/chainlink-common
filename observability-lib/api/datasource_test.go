package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDataSourceByName(t *testing.T) {
	expectedName := "VictoriaMetrics - MetricsQL"
	expectedUID := "123abc"
	expectedType := "victoriametrics-metrics-datasource"
	expectedID := uint(1)

	// Create a test HTTP server that mimics Grafana's API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/datasources/name/%s", expectedName), r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		t.Logf("received request: %s %s", r.Method, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{
			"id": %d,
			"uid": "%s",
			"name": "%s",
			"type": "%s"
		}`, expectedID, expectedUID, expectedName, expectedType)
	}))
	defer ts.Close()

	client := &Client{
		resty: resty.New().SetBaseURL(ts.URL),
	}

	ds, resp, err := client.GetDataSourceByName(expectedName)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.NotNil(t, ds)
	assert.Equal(t, expectedID, ds.ID)
	assert.Equal(t, expectedUID, ds.UID)
	assert.Equal(t, expectedName, ds.Name)
	assert.Equal(t, expectedType, ds.Type)
}
