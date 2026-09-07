package grafana_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

// fakeGrafana stubs the Grafana endpoints used by DeployToGrafana and records
// how the alert-rule endpoints are called.
type fakeGrafana struct {
	mu sync.Mutex

	folderGets     int
	alertRuleGets  int
	alertRulePosts int
	inFlightPosts  int
	maxInFlight    int
}

func (f *fakeGrafana) handler(t *testing.T) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/folders", func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		f.folderGets++
		f.mu.Unlock()
		writeJSON(t, w, []map[string]any{{"id": 1, "uid": "folder-uid", "title": "Folder"}})
	})
	mux.HandleFunc("GET /api/search", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, []map[string]any{})
	})
	mux.HandleFunc("POST /api/dashboards/db", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"uid": "dash-uid", "url": "/d/dash-uid"})
	})
	mux.HandleFunc("GET /api/v1/provisioning/alert-rules", func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		f.alertRuleGets++
		f.mu.Unlock()
		writeJSON(t, w, []map[string]any{})
	})
	mux.HandleFunc("POST /api/v1/provisioning/alert-rules", func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		f.alertRulePosts++
		f.inFlightPosts++
		if f.inFlightPosts > f.maxInFlight {
			f.maxInFlight = f.inFlightPosts
		}
		f.mu.Unlock()

		// Hold the request so overlapping POSTs are observable.
		time.Sleep(50 * time.Millisecond)

		f.mu.Lock()
		f.inFlightPosts--
		f.mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		writeJSON(t, w, map[string]any{})
	})
	mux.HandleFunc("PUT /api/v1/provisioning/folder/{folderUID}/rule-groups/{group}", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{})
	})

	return mux
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(v))
}

func TestDeployToGrafanaFetchesRulesOnceAndWritesConcurrently(t *testing.T) {
	fake := &fakeGrafana{}
	server := httptest.NewServer(fake.handler(t))
	t.Cleanup(server.Close)

	const numAlerts = 20

	title := "Test Dashboard"
	o := &grafana.Observability{
		Dashboard: &dashboard.Dashboard{
			Title: &title,
		},
		Alerts:      make([]alerting.Rule, numAlerts),
		AlertGroups: nil,
	}
	for i := range o.Alerts {
		o.Alerts[i] = alerting.Rule{
			Title:     fmt.Sprintf("alert-%d", i),
			RuleGroup: "group",
			Condition: "A",
			Data:      []alerting.Query{},
		}
	}

	// Alert groups exercise the folder+group lookup in getAlertRules; with the
	// dashboard UID lookup it must still fetch the full rule list only once.
	group, err := grafana.NewAlertGroup(&grafana.AlertGroupOptions{Title: "group", Interval: 60}).Build()
	require.NoError(t, err)
	o.AlertGroups = []alerting.RuleGroup{group}

	err = o.DeployToGrafana(&grafana.DeployOptions{
		GrafanaURL:             server.URL,
		GrafanaToken:           "test-token",
		FolderName:             "Folder",
		EnableAlerts:           true,
		RuleGroupFromDashboard: true,
	})
	require.NoError(t, err)

	require.Equal(t, 1, fake.alertRuleGets, "full alert rule list must be fetched exactly once per deploy")
	require.Equal(t, numAlerts, fake.alertRulePosts)
	require.Greater(t, fake.maxInFlight, 1, "alert rule writes should overlap")
	require.LessOrEqual(t, fake.maxInFlight, 8, "alert rule writes must respect the default concurrency bound")
}

// deployOne deploys a single-dashboard observability with the given options,
// failing the test on any error.
func deployOne(t *testing.T, title string, opts *grafana.DeployOptions) {
	t.Helper()
	o := &grafana.Observability{
		Dashboard: &dashboard.Dashboard{Title: &title},
		Alerts: []alerting.Rule{{
			Title:     "alert-" + title,
			RuleGroup: "group",
			Condition: "A",
			Data:      []alerting.Query{},
		}},
	}
	group, err := grafana.NewAlertGroup(&grafana.AlertGroupOptions{Title: "group", Interval: 60}).Build()
	require.NoError(t, err)
	o.AlertGroups = []alerting.RuleGroup{group}

	require.NoError(t, o.DeployToGrafana(opts))
}

func TestDeployToGrafanaSharesCacheAcrossDeploys(t *testing.T) {
	fake := &fakeGrafana{}
	server := httptest.NewServer(fake.handler(t))
	t.Cleanup(server.Close)

	cache := &grafana.DeployCache{}
	for _, title := range []string{"Dashboard A", "Dashboard B"} {
		deployOne(t, title, &grafana.DeployOptions{
			GrafanaURL:             server.URL,
			GrafanaToken:           "test-token",
			FolderName:             "Folder",
			EnableAlerts:           true,
			RuleGroupFromDashboard: true,
			Cache:                  cache,
		})
	}

	require.Equal(t, 1, fake.folderGets, "shared cache must resolve the folder once across deploys")
	require.Equal(t, 1, fake.alertRuleGets, "shared cache must fetch the full alert rule list once across deploys")
}

func TestDeployToGrafanaWithoutCacheFetchesPerDeploy(t *testing.T) {
	fake := &fakeGrafana{}
	server := httptest.NewServer(fake.handler(t))
	t.Cleanup(server.Close)

	for _, title := range []string{"Dashboard A", "Dashboard B"} {
		deployOne(t, title, &grafana.DeployOptions{
			GrafanaURL:             server.URL,
			GrafanaToken:           "test-token",
			FolderName:             "Folder",
			EnableAlerts:           true,
			RuleGroupFromDashboard: true,
		})
	}

	require.Equal(t, 2, fake.folderGets, "without a cache each deploy resolves the folder")
	require.Equal(t, 2, fake.alertRuleGets, "without a cache each deploy fetches the full alert rule list")
}
