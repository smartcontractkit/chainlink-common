package grafana_test

import (
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
