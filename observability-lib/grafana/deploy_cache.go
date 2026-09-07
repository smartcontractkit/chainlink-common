package grafana

import (
	"sync"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"

	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
)

type deployCache interface {
	folder(key string) (*api.Folder, bool)
	setFolder(key string, f *api.Folder)
	alertRules() ([]alerting.Rule, bool)
	setAlertRules(rules []alerting.Rule)
}

type noopDeployCache struct{}

func (noopDeployCache) folder(string) (*api.Folder, bool)   { return nil, false }
func (noopDeployCache) setFolder(string, *api.Folder)       {}
func (noopDeployCache) alertRules() ([]alerting.Rule, bool) { return nil, false }
func (noopDeployCache) setAlertRules([]alerting.Rule)       {}

// DeployCache memoizes values that are invariant across the dashboard deploys
// of a single run — folder resolution and the full alert-rule list — so a
// composite deploy pays those Grafana lookups once instead of once per
// dashboard.//
// Scoping rules:
//   - A DeployCache is bound to one Grafana instance: do not share it across
//     DeployOptions with different GrafanaURL/GrafanaToken.
//   - The alert-rule snapshot is taken on first use and is not invalidated by
//     the deploys' own writes. That is safe when each deploy in the run touches
//     a disjoint set of rules (e.g. one rule group per dashboard, the default
//     with RuleGroupFromDashboard); do not reuse a cache for a repeated deploy
//     of the same dashboard.
//
// Safe for concurrent use.
type DeployCache struct {
	mu sync.Mutex
	// folders by resolution key: "uid_"+FolderUID or "name_"+FolderName.
	folders map[string]*api.Folder
	// rules holds the full alert-rule list for the instance; rulesReady
	// distinguishes "fetched, empty" from "not fetched yet".
	rules      []alerting.Rule
	rulesReady bool
}

// folder returns the cached folder for key, if any.
func (c *DeployCache) folder(key string) (*api.Folder, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	f, ok := c.folders[key]
	return f, ok
}

func (c *DeployCache) setFolder(key string, f *api.Folder) {
	if f == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.folders == nil {
		c.folders = make(map[string]*api.Folder)
	}
	c.folders[key] = f
}

// alertRules returns the cached full alert-rule list; the second return value
// reports whether it has been fetched (an empty list is a valid snapshot).
func (c *DeployCache) alertRules() ([]alerting.Rule, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.rules, c.rulesReady
}

func (c *DeployCache) setAlertRules(rules []alerting.Rule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules = rules
	c.rulesReady = true
}
