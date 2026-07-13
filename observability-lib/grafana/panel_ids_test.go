package grafana_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

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
	builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title:    grafana.Pointer("Top Level Pinned"),
			StableID: 20110,
		},
	}))

	o, err := builder.Build()
	require.NoError(t, err)

	id, ok := grafana.PanelIDByTitle(o.Dashboard, "Inside Row")
	require.True(t, ok)
	require.Equal(t, uint32(20127), id)

	id, ok = grafana.PanelIDByTitle(o.Dashboard, "Top Level Pinned")
	require.True(t, ok)
	require.Equal(t, uint32(20110), id)

	id, ok = grafana.PanelIDByTitle(o.Dashboard, "Auto Panel")
	require.True(t, ok)
	require.NotEqual(t, uint32(20127), id)
	require.NotEqual(t, uint32(20110), id)
}

func TestBuilderAutoPanelIDSkipsStableID(t *testing.T) {
	builder := grafana.NewBuilder(&grafana.BuilderOptions{Name: "Skip Stable ID"})
	builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title:    grafana.Pointer("Pinned Low"),
			StableID: 5,
		},
	}))
	for i := 1; i <= 5; i++ {
		builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
			PanelOptions: &grafana.PanelOptions{
				Title: grafana.Pointer("Auto " + strconv.Itoa(i)),
			},
		}))
	}

	o, err := builder.Build()
	require.NoError(t, err)

	pinned, ok := grafana.PanelIDByTitle(o.Dashboard, "Pinned Low")
	require.True(t, ok)
	require.Equal(t, uint32(5), pinned)

	for i := 1; i <= 5; i++ {
		title := "Auto " + strconv.Itoa(i)
		id, found := grafana.PanelIDByTitle(o.Dashboard, title)
		require.True(t, found, "panel %q", title)
		require.NotEqual(t, uint32(5), id, "auto panel %q should skip reserved stable id", title)
	}

	sixth, found := grafana.PanelIDByTitle(o.Dashboard, "Auto 5")
	require.True(t, found)
	require.Equal(t, uint32(6), sixth)
}
