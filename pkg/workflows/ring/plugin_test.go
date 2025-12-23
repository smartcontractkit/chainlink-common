package ring

import (
	"testing"
	"time"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/ring/pb"
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
	intialSeqNr := uint64(42)
	outcomeCtx := ocr3types.OutcomeContext{SeqNr: intialSeqNr}

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
			Now:    timestamppb.Now(),
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
	require.NotNil(t, outcomeProto.State)
	require.Equal(t, intialSeqNr+1, outcomeProto.State.Id, "ID should match SeqNr")
	t.Logf("Outcome - ID: %d, HealthyShards: %v", outcomeProto.State.Id, outcomeProto.State.GetRoutableShards())
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

func TestPlugin_StateTransitions(t *testing.T) {
	lggr := logger.Test(t)
	store := NewStore()

	config := ocr3types.ReportingPluginConfig{
		N: 4, F: 1,
	}

	// Use short time to sync for testing
	plugin, err := NewPlugin(store, config, lggr, &ConsensusConfig{
		MinShardCount: 1,
		MaxShardCount: 10,
		BatchSize:     100,
		TimeToSync:    1 * time.Second,
	})
	require.NoError(t, err)

	ctx := t.Context()
	now := time.Now()

	// Test 1: Initial state with no previous outcome
	t.Run("initial_state", func(t *testing.T) {
		outcomeCtx := ocr3types.OutcomeContext{
			SeqNr:           1,
			PreviousOutcome: nil,
		}

		// Only 1 healthy shard in observations to match minShardCount
		aos := makeObservations(t, []map[uint32]bool{
			{0: true},
			{0: true},
			{0: true},
		}, []string{"wf-1"}, now)

		outcome, err := plugin.Outcome(ctx, outcomeCtx, nil, aos)
		require.NoError(t, err)

		outcomeProto := &pb.Outcome{}
		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)

		// Should be in stable state with min shard count
		require.NotNil(t, outcomeProto.State.GetRoutableShards())
		require.Equal(t, uint32(1), outcomeProto.State.GetRoutableShards())
		t.Logf("Initial state: %d routable shards", outcomeProto.State.GetRoutableShards())
	})

	// Test 2: Transition triggered when shard count changes
	t.Run("transition_triggered", func(t *testing.T) {
		// Start with 1 shard in stable state
		priorOutcome := &pb.Outcome{
			State: &pb.RoutingState{
				Id: 1,
				State: &pb.RoutingState_RoutableShards{
					RoutableShards: 1,
				},
			},
			Routes: map[string]*pb.WorkflowRoute{},
		}
		priorBytes, err := proto.Marshal(priorOutcome)
		require.NoError(t, err)

		outcomeCtx := ocr3types.OutcomeContext{
			SeqNr:           2,
			PreviousOutcome: priorBytes,
		}

		// Observations show 2 healthy shards now
		aos := makeObservations(t, []map[uint32]bool{
			{0: true, 1: true},
			{0: true, 1: true},
			{0: true, 1: true},
		}, []string{"wf-1"}, now)

		outcome, err := plugin.Outcome(ctx, outcomeCtx, nil, aos)
		require.NoError(t, err)

		outcomeProto := &pb.Outcome{}
		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)

		// Should transition to Transition state
		transition := outcomeProto.State.GetTransition()
		require.NotNil(t, transition, "should be in transition state")
		require.Equal(t, uint32(2), transition.WantShards, "want 2 shards")
		require.Equal(t, uint32(1), transition.LastStableCount, "was at 1 shard")
		require.True(t, transition.ChangesSafeAfter.AsTime().After(now), "safety period should be in future")
		t.Logf("Transition: %d → %d, safe after %v", transition.LastStableCount, transition.WantShards, transition.ChangesSafeAfter.AsTime())
	})

	// Test 3: Stay in transition during safety period
	t.Run("stay_in_transition", func(t *testing.T) {
		safeAfter := now.Add(1 * time.Hour)
		priorOutcome := &pb.Outcome{
			State: &pb.RoutingState{
				Id: 2,
				State: &pb.RoutingState_Transition{
					Transition: &pb.Transition{
						WantShards:       2,
						LastStableCount:  1,
						ChangesSafeAfter: timestamppb.New(safeAfter),
					},
				},
			},
			Routes: map[string]*pb.WorkflowRoute{},
		}
		priorBytes, err := proto.Marshal(priorOutcome)
		require.NoError(t, err)

		outcomeCtx := ocr3types.OutcomeContext{
			SeqNr:           3,
			PreviousOutcome: priorBytes,
		}

		// Still showing 2 healthy shards, but safety period not elapsed
		aos := makeObservations(t, []map[uint32]bool{
			{0: true, 1: true},
			{0: true, 1: true},
			{0: true, 1: true},
		}, []string{"wf-1"}, now)

		outcome, err := plugin.Outcome(ctx, outcomeCtx, nil, aos)
		require.NoError(t, err)

		outcomeProto := &pb.Outcome{}
		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)

		// Should still be in transition state
		transition := outcomeProto.State.GetTransition()
		require.NotNil(t, transition, "should still be in transition")
		require.Equal(t, uint32(2), transition.WantShards)
		t.Logf("Still in transition, waiting for safety period")
	})

	// Test 4: Complete transition after safety period
	t.Run("complete_transition", func(t *testing.T) {
		safeAfter := now.Add(-1 * time.Second) // Safety period already passed
		priorOutcome := &pb.Outcome{
			State: &pb.RoutingState{
				Id: 2,
				State: &pb.RoutingState_Transition{
					Transition: &pb.Transition{
						WantShards:       2,
						LastStableCount:  1,
						ChangesSafeAfter: timestamppb.New(safeAfter),
					},
				},
			},
			Routes: map[string]*pb.WorkflowRoute{},
		}
		priorBytes, err := proto.Marshal(priorOutcome)
		require.NoError(t, err)

		outcomeCtx := ocr3types.OutcomeContext{
			SeqNr:           3,
			PreviousOutcome: priorBytes,
		}

		aos := makeObservations(t, []map[uint32]bool{
			{0: true, 1: true},
			{0: true, 1: true},
			{0: true, 1: true},
		}, []string{"wf-1"}, now)

		outcome, err := plugin.Outcome(ctx, outcomeCtx, nil, aos)
		require.NoError(t, err)

		outcomeProto := &pb.Outcome{}
		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)

		// Should now be in stable state with 2 shards
		require.NotNil(t, outcomeProto.State.GetRoutableShards(), "should be in stable state")
		require.Equal(t, uint32(2), outcomeProto.State.GetRoutableShards())
		require.Equal(t, uint64(3), outcomeProto.State.Id, "state ID should increment")
		t.Logf("Transition complete: now at %d routable shards", outcomeProto.State.GetRoutableShards())
	})

	// Test 5: Stay stable when shard count doesn't change
	t.Run("stay_stable", func(t *testing.T) {
		priorOutcome := &pb.Outcome{
			State: &pb.RoutingState{
				Id: 3,
				State: &pb.RoutingState_RoutableShards{
					RoutableShards: 2,
				},
			},
			Routes: map[string]*pb.WorkflowRoute{},
		}
		priorBytes, err := proto.Marshal(priorOutcome)
		require.NoError(t, err)

		outcomeCtx := ocr3types.OutcomeContext{
			SeqNr:           4,
			PreviousOutcome: priorBytes,
		}

		// Same 2 healthy shards
		aos := makeObservations(t, []map[uint32]bool{
			{0: true, 1: true},
			{0: true, 1: true},
			{0: true, 1: true},
		}, []string{"wf-1"}, now)

		outcome, err := plugin.Outcome(ctx, outcomeCtx, nil, aos)
		require.NoError(t, err)

		outcomeProto := &pb.Outcome{}
		err = proto.Unmarshal(outcome, outcomeProto)
		require.NoError(t, err)

		// Should stay in stable state, ID unchanged
		require.NotNil(t, outcomeProto.State.GetRoutableShards())
		require.Equal(t, uint32(2), outcomeProto.State.GetRoutableShards())
		require.Equal(t, uint64(3), outcomeProto.State.Id, "state ID should not change when stable")
		t.Logf("Staying stable at %d routable shards", outcomeProto.State.GetRoutableShards())
	})
}

// Helper function to create observations
func makeObservations(t *testing.T, shardHealths []map[uint32]bool, workflows []string, now time.Time) []types.AttributedObservation {
	aos := make([]types.AttributedObservation, 0, len(shardHealths))
	for i, health := range shardHealths {
		pbObs := &pb.Observation{
			Status: health,
			Hashes: workflows,
			Now:    timestamppb.New(now),
		}
		rawObs, err := proto.Marshal(pbObs)
		require.NoError(t, err)

		aos = append(aos, types.AttributedObservation{
			Observation: rawObs,
			Observer:    commontypes.OracleID(i),
		})
	}
	return aos
}
