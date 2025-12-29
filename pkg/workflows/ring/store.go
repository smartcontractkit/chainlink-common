package ring

import (
	"slices"
	"sync"
)

// Store manages shard routing state and workflow mappings
type Store struct {
	routingState  map[string]uint32 // workflow_id -> shard_id
	shardHealth   map[uint32]bool   // shard_id -> is_healthy
	healthyShards []uint32          // Sorted list of healthy shards
	mu            sync.Mutex
}

func NewStore() *Store {
	return &Store{
		routingState:  make(map[string]uint32),
		shardHealth:   make(map[uint32]bool),
		healthyShards: make([]uint32, 0),
		mu:            sync.Mutex{},
	}
}

// updateHealthyShards rebuilds the sorted list of healthy shards
func (s *Store) updateHealthyShards() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.healthyShards = make([]uint32, 0)

	for shardID, healthy := range s.shardHealth {
		if healthy {
			s.healthyShards = append(s.healthyShards, shardID)
		}
	}

	// Sort for determinism
	slices.Sort(s.healthyShards)

	// If no healthy shards, add shard 0 as fallback
	if len(s.healthyShards) == 0 {
		s.healthyShards = []uint32{0}
	}
}

// GetShardForWorkflow deterministically assigns a workflow to a shard using consistent hashing.
func (s *Store) GetShardForWorkflow(workflowID string) uint32 {
	s.mu.Lock()
	shardCount := uint32(len(s.healthyShards))
	s.mu.Unlock()

	return getShardForWorkflow(workflowID, shardCount)
}

func (s *Store) SetShardForWorkflow(workflowID string, shardID uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routingState[workflowID] = shardID
}

func (s *Store) GetShardHealth() map[uint32]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make(map[uint32]bool)
	for k, v := range s.shardHealth {
		copied[k] = v
	}
	return copied
}

func (s *Store) SetShardHealth(shardID uint32, healthy bool) {
	s.mu.Lock()
	s.shardHealth[shardID] = healthy
	s.mu.Unlock()

	// Rebuild healthy shards list when shard health changes
	s.updateHealthyShards()
}

func (s *Store) SetAllShardHealth(health map[uint32]bool) {
	s.mu.Lock()
	s.shardHealth = make(map[uint32]bool)
	for k, v := range health {
		s.shardHealth[k] = v
	}
	s.mu.Unlock()

	// Rebuild healthy shards list
	s.updateHealthyShards()
}

func (s *Store) GetAllRoutingState() map[string]uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make(map[string]uint32)
	for k, v := range s.routingState {
		copied[k] = v
	}
	return copied
}

func (s *Store) DeleteWorkflow(workflowID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.routingState, workflowID)
}

func (s *Store) GetHealthyShardCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, healthy := range s.shardHealth {
		if healthy {
			count++
		}
	}
	return count
}

// GetHealthyShards returns a sorted list of healthy shards for inspection
func (s *Store) GetHealthyShards() []uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.healthyShards)
}
