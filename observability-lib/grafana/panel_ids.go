package grafana

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
)

// PanelIDByTitle returns the panel ID for a dashboard panel matched by title.
// It searches top-level panels and panels nested inside row containers; row
// container panels themselves are not matched.
func PanelIDByTitle(db *dashboard.Dashboard, title string) (uint32, bool) {
	if db == nil || title == "" {
		return 0, false
	}
	var found uint32
	ok := false
	foreachPanel(db, func(panel *dashboard.Panel) bool {
		if panel.Title != nil && *panel.Title == title && panel.Id != nil {
			found = *panel.Id
			ok = true
			return false
		}
		return true
	})
	return found, ok
}

func foreachPanel(db *dashboard.Dashboard, fn func(panel *dashboard.Panel) bool) {
	for i := range db.Panels {
		if !applyToOrRowPanel(&db.Panels[i], fn) {
			return
		}
	}
}

func applyToOrRowPanel(item *dashboard.PanelOrRowPanel, fn func(panel *dashboard.Panel) bool) bool {
	if item.Panel != nil {
		if !fn(item.Panel) {
			return false
		}
	}
	if item.RowPanel != nil {
		for j := range item.RowPanel.Panels {
			if !fn(&item.RowPanel.Panels[j]) {
				return false
			}
		}
	}
	return true
}
