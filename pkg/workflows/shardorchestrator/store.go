package shardorchestrator

import (
	"slices"
	"strconv"
	"sync"

	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash/v2"
)

// xxhashHasher implements the consistent.Hasher interface using xxhash
type xxhashHasher struct{}

func (h xxhashHasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

// ShardMember implements consistent.Member for shard IDs
type ShardMember string

func (m ShardMember) String() string {
	return string(m)
}

// Store manages shard routing state and workflow mappings
type Store struct {
	routingState   map[string]uint32      // workflow_id -> shard_id
	shardHealth    map[uint32]bool        // shard_id -> is_healthy
	consistentHash *consistent.Consistent // Consistent hash ring for routing
	healthyShards  []uint32               // Sorted list of healthy shards
	mu             sync.Mutex
}

func NewStore() *Store {
	return &Store{
		routingState:  make(map[string]uint32),
		shardHealth:   make(map[uint32]bool),
		healthyShards: make([]uint32, 0),
		mu:            sync.Mutex{},
	}
}

// consistentHashConfig returns the configuration for consistent hashing
// Matches prototype: PartitionCount=997 (prime), ReplicationFactor=50, Load=1.1
func consistentHashConfig() consistent.Config {
	return consistent.Config{
		PartitionCount:    997, // Prime number for better distribution
		ReplicationFactor: 50,  // Number of replicas per node
		Load:              1.1, // Load factor for bounded loads
		Hasher:            xxhashHasher{},
	}
}

// updateConsistentHash rebuilds the consistent hash ring based on healthy shards
func (s *Store) updateConsistentHash() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get list of healthy shards and create members
	healthyMembers := make([]consistent.Member, 0)
	s.healthyShards = make([]uint32, 0)

	for shardID, healthy := range s.shardHealth {
		if healthy {
			healthyMembers = append(healthyMembers, ShardMember(strconv.FormatUint(uint64(shardID), 10)))
			s.healthyShards = append(s.healthyShards, shardID)
		}
	}

	// Sort for determinism
	slices.Sort(s.healthyShards)

	// If no healthy shards, add shard 0 as fallback
	if len(healthyMembers) == 0 {
		healthyMembers = append(healthyMembers, ShardMember("0"))
		s.healthyShards = []uint32{0}
	}

	// Create consistent hash ring
	s.consistentHash = consistent.New(healthyMembers, consistentHashConfig())
}

// GetShardForWorkflow deterministically assigns a workflow to a shard using consistent hashing.
// The assignment uses the same algorithm as the prototype: xxhash + consistent hashing ring.
func (s *Store) GetShardForWorkflow(workflowID string) uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.consistentHash == nil {
		// Fallback if hash ring not initialized
		return 0
	}

	// Use consistent hashing to find the member for this workflow
	member := s.consistentHash.LocateKey([]byte(workflowID))
	if member == nil {
		return 0
	}

	// Parse shard ID from member name
	shardID, err := strconv.ParseUint(member.String(), 10, 32)
	if err != nil {
		return 0
	}

	return uint32(shardID)
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

	// Rebuild consistent hash ring when shard health changes
	s.updateConsistentHash()
}

func (s *Store) SetAllShardHealth(health map[uint32]bool) {
	s.mu.Lock()
	s.shardHealth = make(map[uint32]bool)
	for k, v := range health {
		s.shardHealth[k] = v
	}
	s.mu.Unlock()

	// Rebuild consistent hash ring
	s.updateConsistentHash()
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
