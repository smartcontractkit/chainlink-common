package shardorchestrator

import (
	"maps"
	"sync"
	"time"
)

var DefaultRequestTimeout = 20 * time.Minute

// Store manages shard routing state and workflow mappings
type Store struct {
	routingState map[string]uint32 // workflow_id -> shard_id
	shardHealth  map[uint32]bool   // shard_id -> is_healthy
	mu           sync.Mutex
}

func NewStore() *Store {
	return &Store{
		routingState: make(map[string]uint32),
		shardHealth:  make(map[uint32]bool),
		mu:           sync.Mutex{},
	}
}

func (s *Store) GetShardForWorkflow(workflowID string) (uint32, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	shardID, ok := s.routingState[workflowID]
	return shardID, ok
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
	maps.Copy(copied, s.shardHealth)
	return copied
}

func (s *Store) SetShardHealth(shardID uint32, healthy bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shardHealth[shardID] = healthy
}

func (s *Store) SetAllShardHealth(health map[uint32]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shardHealth = make(map[uint32]bool)
	maps.Copy(s.shardHealth, health)
}

func (s *Store) GetAllRoutingState() map[string]uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make(map[string]uint32)
	maps.Copy(copied, s.routingState)
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
