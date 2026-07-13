package grafana_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

func TestValidateStablePanelIDs(t *testing.T) {
	t.Run("accepts valid map", func(t *testing.T) {
		err := grafana.ValidateStablePanelIDs(map[string]uint32{
			"Panel A": 20101,
			"Panel B": 20102,
		})
		require.NoError(t, err)
	})

	t.Run("rejects duplicate IDs", func(t *testing.T) {
		err := grafana.ValidateStablePanelIDs(map[string]uint32{
			"Panel A": 20101,
			"Panel B": 20101,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate stable panel ID")
	})

	t.Run("rejects zero ID", func(t *testing.T) {
		err := grafana.ValidateStablePanelIDs(map[string]uint32{
			"Panel A": 0,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be non-zero")
	})
}

func TestApplyStablePanelIDs(t *testing.T) {
	builder := grafana.NewBuilder(&grafana.BuilderOptions{Name: "Stable IDs"})
	builder.AddRow("Metrics")
	builder.AddPanelToRow("Metrics",
		grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{Title: grafana.Pointer("Pinned Panel")},
		}),
		grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{Title: grafana.Pointer("Auto Panel")},
		}),
	)
	builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{Title: grafana.Pointer("Top Level Pinned")},
	}))

	o, err := builder.Build()
	require.NoError(t, err)

	err = grafana.ApplyStablePanelIDs(o.Dashboard, map[string]uint32{
		"Pinned Panel":      20105,
		"Top Level Pinned":  20110,
	})
	require.NoError(t, err)

	id, ok := grafana.PanelIDByTitle(o.Dashboard, "Pinned Panel")
	require.True(t, ok)
	require.Equal(t, uint32(20105), id)

	id, ok = grafana.PanelIDByTitle(o.Dashboard, "Top Level Pinned")
	require.True(t, ok)
	require.Equal(t, uint32(20110), id)

	id, ok = grafana.PanelIDByTitle(o.Dashboard, "Auto Panel")
	require.True(t, ok)
	require.NotEqual(t, uint32(20105), id)
	require.NotEqual(t, uint32(20110), id)
}

func TestBuilderPanelOptionsStableID(t *testing.T) {
	builder := grafana.NewBuilder(&grafana.BuilderOptions{Name: "PanelOptions StableID"})
	builder.AddRow("Row")
	builder.AddPanelToRow("Row", grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title:    grafana.Pointer("Inside Row"),
			StableID: 20127,
		},
	}))
	builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{Title: grafana.Pointer("Auto Panel")},
	}))

	o, err := builder.Build()
	require.NoError(t, err)

	id, ok := grafana.PanelIDByTitle(o.Dashboard, "Inside Row")
	require.True(t, ok)
	require.Equal(t, uint32(20127), id)

	id, ok = grafana.PanelIDByTitle(o.Dashboard, "Auto Panel")
	require.True(t, ok)
	require.NotEqual(t, uint32(20127), id)
}

func TestBuilderPanelOptionsStableIDDuplicate(t *testing.T) {
	builder := grafana.NewBuilder(&grafana.BuilderOptions{Name: "Duplicate StableID"})
	builder.AddPanel(
		grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{Title: grafana.Pointer("A"), StableID: 20101},
		}),
		grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{Title: grafana.Pointer("B"), StableID: 20101},
		}),
	)

	_, err := builder.Build()
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate StableID")
}

func TestBuilderWithStablePanelIDs(t *testing.T) {
	builder := grafana.NewBuilder(&grafana.BuilderOptions{Name: "Builder Stable IDs"})
	builder.WithStablePanelIDs(map[string]uint32{
		"Inside Row": 20127,
	})
	builder.AddRow("Row")
	builder.AddPanelToRow("Row", grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{Title: grafana.Pointer("Inside Row")},
	}))

	o, err := builder.Build()
	require.NoError(t, err)

	id, ok := grafana.PanelIDByTitle(o.Dashboard, "Inside Row")
	require.True(t, ok)
	require.Equal(t, uint32(20127), id)
}

func TestBuilderWithStablePanelIDsInvalidMap(t *testing.T) {
	builder := grafana.NewBuilder(&grafana.BuilderOptions{Name: "Invalid Stable IDs"})
	builder.WithStablePanelIDs(map[string]uint32{
		"A": 1,
		"B": 1,
	})
	builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{Title: grafana.Pointer("A")},
	}))

	_, err := builder.Build()
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate stable panel ID")
}
