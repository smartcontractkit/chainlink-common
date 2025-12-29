package ring

import (
	"context"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/ring/pb"
	"github.com/stretchr/testify/require"
)

func TestStore_DeterministicHashing(t *testing.T) {
	store := NewStore()

	// Set up healthy shards
	store.SetAllShardHealth(map[uint32]bool{
		0: true,
		1: true,
		2: true,
	})

	ctx := context.Background()

	// Test determinism: same workflow always gets same shard
	shard1, err := store.GetShardForWorkflow(ctx, "workflow-123")
	require.NoError(t, err)
	shard2, err := store.GetShardForWorkflow(ctx, "workflow-123")
	require.NoError(t, err)
	shard3, err := store.GetShardForWorkflow(ctx, "workflow-123")
	require.NoError(t, err)

	require.Equal(t, shard1, shard2, "Same workflow should get same shard (call 2)")
	require.Equal(t, shard2, shard3, "Same workflow should get same shard (call 3)")
	require.True(t, shard1 >= 0 && shard1 <= 2, "Shard should be in healthy set")
}

func TestStore_ConsistentRingConsistency(t *testing.T) {
	store1 := NewStore()
	store2 := NewStore()
	store3 := NewStore()

	// All stores with same healthy shards
	healthyShards := map[uint32]bool{0: true, 1: true, 2: true}
	store1.SetAllShardHealth(healthyShards)
	store2.SetAllShardHealth(healthyShards)
	store3.SetAllShardHealth(healthyShards)

	ctx := context.Background()

	// All compute same assignments
	workflows := []string{"workflow-A", "workflow-B", "workflow-C", "workflow-D"}
	for _, wf := range workflows {
		s1, err := store1.GetShardForWorkflow(ctx, wf)
		require.NoError(t, err)
		s2, err := store2.GetShardForWorkflow(ctx, wf)
		require.NoError(t, err)
		s3, err := store3.GetShardForWorkflow(ctx, wf)
		require.NoError(t, err)

		require.Equal(t, s1, s2, "All nodes should agree on %s assignment", wf)
		require.Equal(t, s2, s3, "All nodes should agree on %s assignment", wf)
	}
}

func TestStore_Rebalancing(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	// Start with 3 healthy shards
	store.SetAllShardHealth(map[uint32]bool{0: true, 1: true, 2: true})
	assignments1 := make(map[string]uint32)
	for i := 1; i <= 10; i++ {
		wfID := "workflow-" + string(rune(i))
		shard, err := store.GetShardForWorkflow(ctx, wfID)
		require.NoError(t, err)
		assignments1[wfID] = shard
	}

	// Shard 1 fails
	store.SetShardHealth(1, false)
	assignments2 := make(map[string]uint32)
	for i := 1; i <= 10; i++ {
		wfID := "workflow-" + string(rune(i))
		shard, err := store.GetShardForWorkflow(ctx, wfID)
		require.NoError(t, err)
		assignments2[wfID] = shard
	}

	// Check that rebalancing occurred (some workflows moved)
	healthyShards := store.GetHealthyShards()
	require.Equal(t, 2, len(healthyShards), "Should have 2 healthy shards")
	require.NotContains(t, healthyShards, uint32(1), "Shard 1 should not be healthy")

	// Verify that workflows on healthy shards did not move
	for wfID, originalShard := range assignments1 {
		if originalShard == 0 || originalShard == 2 {
			require.Equal(t, originalShard, assignments2[wfID],
				"Workflow %s on healthy shard %d should not have moved", wfID, originalShard)
		}
	}
}

func TestStore_GetHealthyShards(t *testing.T) {
	store := NewStore()

	store.SetAllShardHealth(map[uint32]bool{
		3: true,
		1: true,
		2: true,
	})

	healthyShards := store.GetHealthyShards()
	require.Len(t, healthyShards, 3)
	// Should be sorted
	require.Equal(t, []uint32{1, 2, 3}, healthyShards)
}

func TestStore_NilHashRingFallback(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	// Should not panic, should return 0 as fallback (no healthy shards set)
	shard, err := store.GetShardForWorkflow(ctx, "workflow-123")
	require.NoError(t, err)
	require.Equal(t, uint32(0), shard)
}

func TestStore_DistributionAcrossShards(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	store.SetAllShardHealth(map[uint32]bool{
		0: true,
		1: true,
		2: true,
	})

	// Generate many workflows and check distribution
	distribution := make(map[uint32]int)
	for i := 0; i < 100; i++ {
		wfID := "workflow-" + string(rune(i))
		shard, err := store.GetShardForWorkflow(ctx, wfID)
		require.NoError(t, err)
		distribution[shard]++
	}

	require.Equal(t, 38, distribution[0], "Should have 29 workflows on shard 0")
	require.Equal(t, 29, distribution[1], "Should have 33 workflows on shard 1")
	require.Equal(t, 33, distribution[2], "Should have 33 workflows on shard 2")
	require.Equal(t, 100, sum(distribution), "Should have 100 workflows")
}

func sum(distribution map[uint32]int) int {
	total := 0
	for _, count := range distribution {
		total += count
	}
	return total
}

func TestStore_PendingAllocsDuringTransition(t *testing.T) {
	store := NewStore()
	store.SetAllShardHealth(map[uint32]bool{0: true, 1: true})

	// Put store in transition state
	store.SetRoutingState(&pb.RoutingState{
		State: &pb.RoutingState_Transition{
			Transition: &pb.Transition{WantShards: 3},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start a goroutine that requests allocation (will block)
	resultCh := make(chan uint32)
	go func() {
		shard, _ := store.GetShardForWorkflow(ctx, "workflow-X")
		resultCh <- shard
	}()

	// Give goroutine time to enqueue request
	time.Sleep(10 * time.Millisecond)

	// Verify request is pending
	pending := store.GetPendingAllocations()
	require.Contains(t, pending, "workflow-X")

	// Fulfill the allocation (simulates transmitter receiving OCR outcome)
	store.SetShardForWorkflow("workflow-X", 2)

	// Blocked goroutine should now receive result
	select {
	case shard := <-resultCh:
		require.Equal(t, uint32(2), shard)
	case <-time.After(50 * time.Millisecond):
		t.Fatal("allocation was not fulfilled")
	}
}
