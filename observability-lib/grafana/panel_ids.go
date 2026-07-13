package grafana

import (
	"errors"
	"fmt"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
)

// ValidateStablePanelIDs checks that pinned titles are non-empty and panel IDs are unique.
func ValidateStablePanelIDs(byTitle map[string]uint32) error {
	if len(byTitle) == 0 {
		return nil
	}
	seen := make(map[uint32]string, len(byTitle))
	for title, id := range byTitle {
		if title == "" {
			return errors.New("stable panel ID map has empty title")
		}
		if id == 0 {
			return fmt.Errorf("stable panel ID for %q must be non-zero", title)
		}
		if prev, ok := seen[id]; ok {
			return fmt.Errorf("duplicate stable panel ID %d for titles %q and %q", id, prev, title)
		}
		seen[id] = title
	}
	return nil
}

// ApplyStablePanelIDs assigns fixed Grafana panel IDs to panels matched by title.
// Walks top-level panels and panels nested inside row panels.
func ApplyStablePanelIDs(db *dashboard.Dashboard, byTitle map[string]uint32) error {
	if db == nil || len(byTitle) == 0 {
		return nil
	}
	if err := ValidateStablePanelIDs(byTitle); err != nil {
		return err
	}
	for i := range db.Panels {
		applyStablePanelIDToOrRow(&db.Panels[i], byTitle)
	}
	return nil
}

func applyStablePanelIDToOrRow(item *dashboard.PanelOrRowPanel, byTitle map[string]uint32) {
	if item.Panel != nil {
		patchStablePanelID(item.Panel, byTitle)
	}
	if item.RowPanel != nil {
		for j := range item.RowPanel.Panels {
			patchStablePanelID(&item.RowPanel.Panels[j], byTitle)
		}
	}
}

func patchStablePanelID(panel *dashboard.Panel, byTitle map[string]uint32) {
	if panel == nil || panel.Title == nil {
		return
	}
	if id, ok := byTitle[*panel.Title]; ok {
		panel.Id = cog.ToPtr(id)
	}
}

// PanelIDByTitle returns the panel ID for a dashboard panel title, including row panels.
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
