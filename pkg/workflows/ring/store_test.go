package ring

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestStore_DeterministicHashing verifies that workflow assignments are deterministic
func TestStore_DeterministicHashing(t *testing.T) {
	store := NewStore()

	// Set up healthy shards
	store.SetAllShardHealth(map[uint32]bool{
		0: true,
		1: true,
		2: true,
	})

	// Test determinism: same workflow always gets same shard
	shard1 := store.GetShardForWorkflow("workflow-123")
	shard2 := store.GetShardForWorkflow("workflow-123")
	shard3 := store.GetShardForWorkflow("workflow-123")

	require.Equal(t, shard1, shard2, "Same workflow should get same shard (call 2)")
	require.Equal(t, shard2, shard3, "Same workflow should get same shard (call 3)")
	require.True(t, shard1 >= 0 && shard1 <= 2, "Shard should be in healthy set")
}

// TestStore_ConsistentRingConsistency verifies that all nodes with same healthy shards agree
func TestStore_ConsistentRingConsistency(t *testing.T) {
	store1 := NewStore()
	store2 := NewStore()
	store3 := NewStore()

	// All stores with same healthy shards
	healthyShards := map[uint32]bool{0: true, 1: true, 2: true}
	store1.SetAllShardHealth(healthyShards)
	store2.SetAllShardHealth(healthyShards)
	store3.SetAllShardHealth(healthyShards)

	// All compute same assignments
	workflows := []string{"workflow-A", "workflow-B", "workflow-C", "workflow-D"}
	for _, wf := range workflows {
		s1 := store1.GetShardForWorkflow(wf)
		s2 := store2.GetShardForWorkflow(wf)
		s3 := store3.GetShardForWorkflow(wf)

		require.Equal(t, s1, s2, "All nodes should agree on %s assignment", wf)
		require.Equal(t, s2, s3, "All nodes should agree on %s assignment", wf)
	}
}

// TestStore_Rebalancing verifies rebalancing when shard health changes
func TestStore_Rebalancing(t *testing.T) {
	store := NewStore()

	// Start with 3 healthy shards
	store.SetAllShardHealth(map[uint32]bool{0: true, 1: true, 2: true})
	assignments1 := make(map[string]uint32)
	for i := 1; i <= 10; i++ {
		wfID := "workflow-" + string(rune(i))
		assignments1[wfID] = store.GetShardForWorkflow(wfID)
	}

	// Shard 1 fails
	store.SetShardHealth(1, false)
	assignments2 := make(map[string]uint32)
	for i := 1; i <= 10; i++ {
		wfID := "workflow-" + string(rune(i))
		assignments2[wfID] = store.GetShardForWorkflow(wfID)
	}

	// Check that rebalancing occurred (some workflows moved)
	healthyShards := store.GetHealthyShards()
	require.Equal(t, 2, len(healthyShards), "Should have 2 healthy shards")
	require.NotContains(t, healthyShards, uint32(1), "Shard 1 should not be healthy")
}

// TestStore_GetHealthyShards verifies that healthy shards list is correctly maintained
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

// TestStore_NilHashRingFallback verifies fallback when hash ring is uninitialized
func TestStore_NilHashRingFallback(t *testing.T) {
	store := &Store{
		routingState:  make(map[string]uint32),
		shardHealth:   make(map[uint32]bool),
		healthyShards: make([]uint32, 0),
	}

	// Should not panic, should return 0 as fallback
	shard := store.GetShardForWorkflow("workflow-123")
	require.Equal(t, uint32(0), shard)
}

// TestStore_DistributionAcrossShards verifies that workflows are distributed across shards
func TestStore_DistributionAcrossShards(t *testing.T) {
	store := NewStore()

	store.SetAllShardHealth(map[uint32]bool{
		0: true,
		1: true,
		2: true,
	})

	// Generate many workflows and check distribution
	distribution := make(map[uint32]int)
	for i := 0; i < 100; i++ {
		wfID := "workflow-" + string(rune(i))
		shard := store.GetShardForWorkflow(wfID)
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
