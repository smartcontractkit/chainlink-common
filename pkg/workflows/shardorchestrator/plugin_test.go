package shardorchestrator

import (
	"testing"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/shardorchestrator/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

func TestPlugin_OutcomeWithMultiNodeObservations(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore()
	store.SetAllShardHealth(map[uint32]bool{0: true, 1: true, 2: true})

	config := ocr3types.ReportingPluginConfig{
		N: 4, F: 1,
		OffchainConfig:                          []byte{},
		MaxDurationObservation:                  0,
		MaxDurationShouldAcceptAttestedReport:   0,
		MaxDurationShouldTransmitAcceptedReport: 0,
	}

	plugin, err := NewPlugin(store, config, lggr, nil)
	require.NoError(t, err)

	ctx := t.Context()
	outcomeCtx := ocr3types.OutcomeContext{PreviousOutcome: []byte("")}

	// Observations from 4 NOPs reporting health and workflows
	observations := []struct {
		name        string
		shardHealth map[uint32]bool
		workflows   []string
	}{
		{
			name:        "NOP 0",
			shardHealth: map[uint32]bool{0: true, 1: true, 2: true},
			workflows:   []string{"wf-A", "wf-B", "wf-C"},
		},
		{
			name:        "NOP 1",
			shardHealth: map[uint32]bool{0: true, 1: true, 2: true},
			workflows:   []string{"wf-B", "wf-C", "wf-D"},
		},
		{
			name:        "NOP 2",
			shardHealth: map[uint32]bool{0: true, 1: true, 2: false}, // shard 2 unhealthy
			workflows:   []string{"wf-A", "wf-C"},
		},
		{
			name:        "NOP 3",
			shardHealth: map[uint32]bool{0: true, 1: true, 2: true},
			workflows:   []string{"wf-A", "wf-B", "wf-D"},
		},
	}

	// Build attributed observations
	aos := make([]types.AttributedObservation, 0)
	for _, obs := range observations {
		pbObs := &pb.Observation{
			Status: obs.shardHealth,
			Hashes: obs.workflows,
		}
		rawObs, err := proto.Marshal(pbObs)
		require.NoError(t, err)

		aos = append(aos, types.AttributedObservation{
			Observation: rawObs,
			Observer:    commontypes.OracleID(len(aos)),
		})
	}

	// Execute Outcome phase
	outcome, err := plugin.Outcome(ctx, outcomeCtx, nil, aos)
	require.NoError(t, err)
	require.NotNil(t, outcome)

	// Verify outcome
	outcomeProto := &pb.Outcome{}
	err = proto.Unmarshal(outcome, outcomeProto)
	require.NoError(t, err)

	// Check consensus results
	t.Logf("Outcome - ID: %d, HealthyShards: %d", outcomeProto.State.Id, outcomeProto.State.GetRoutableShards())
	t.Logf("Workflows assigned: %d", len(outcomeProto.Routes))

	// Verify all workflows are assigned
	expectedWorkflows := map[string]bool{"wf-A": true, "wf-B": true, "wf-C": true, "wf-D": true}
	require.Equal(t, len(expectedWorkflows), len(outcomeProto.Routes))
	for wf := range expectedWorkflows {
		route, exists := outcomeProto.Routes[wf]
		require.True(t, exists, "workflow %s should be assigned", wf)
		require.True(t, route.Shard <= 2, "shard should be healthy (0-2)")
		t.Logf("  %s → shard %d", wf, route.Shard)
	}

	// Verify determinism: run again, should get same assignments
	outcome2, err := plugin.Outcome(ctx, outcomeCtx, nil, aos)
	require.NoError(t, err)

	outcomeProto2 := &pb.Outcome{}
	err = proto.Unmarshal(outcome2, outcomeProto2)
	require.NoError(t, err)

	// Same workflows → same shards
	for wf, route1 := range outcomeProto.Routes {
		route2, exists := outcomeProto2.Routes[wf]
		require.True(t, exists)
		require.Equal(t, route1.Shard, route2.Shard, "workflow %s should assign to same shard", wf)
	}
}
