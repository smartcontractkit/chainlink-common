package ring

import (
	"context"
	"slices"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/ring/pb"
)

// AllocationRequest represents a pending workflow allocation request during transition
type AllocationRequest struct {
	WorkflowID string
	Result     chan uint32
}

// Store manages shard routing state and workflow mappings
type Store struct {
	routingState  map[string]uint32 // workflow_id -> shard_id (cache of allocated workflows)
	shardHealth   map[uint32]bool   // shard_id -> is_healthy
	healthyShards []uint32          // Sorted list of healthy shards
	currentState  *pb.RoutingState  // Current routing state (steady or transition)

	pendingAllocs map[string][]chan uint32 // workflow_id -> waiting channels
	allocRequests chan AllocationRequest   // Channel for new allocation requests

	mu sync.Mutex
}

func NewStore() *Store {
	return &Store{
		routingState:  make(map[string]uint32),
		shardHealth:   make(map[uint32]bool),
		healthyShards: make([]uint32, 0),
		pendingAllocs: make(map[string][]chan uint32),
		allocRequests: make(chan AllocationRequest, 1000),
		mu:            sync.Mutex{},
	}
}

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

// GetShardForWorkflow returns the shard for a workflow.
// In steady state, it uses consistent hashing from cache.
// In transition state, it enqueues an allocation request and waits for OCR to process it.
func (s *Store) GetShardForWorkflow(ctx context.Context, workflowID string) (uint32, error) {
	s.mu.Lock()

	// Check if already allocated in cache
	if shard, ok := s.routingState[workflowID]; ok {
		s.mu.Unlock()
		return shard, nil
	}

	// In steady state, compute locally using consistent hashing
	if s.currentState == nil || s.isInSteadyState() {
		healthyShards := slices.Clone(s.healthyShards)
		s.mu.Unlock()
		return getShardForWorkflow(workflowID, healthyShards), nil
	}

	// In transition state, enqueue request and wait for allocation
	resultCh := make(chan uint32, 1)
	s.pendingAllocs[workflowID] = append(s.pendingAllocs[workflowID], resultCh)
	s.mu.Unlock()

	select {
	case s.allocRequests <- AllocationRequest{WorkflowID: workflowID, Result: resultCh}:
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	select {
	case shard := <-resultCh:
		return shard, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (s *Store) isInSteadyState() bool {
	if s.currentState == nil {
		return true
	}
	_, ok := s.currentState.State.(*pb.RoutingState_RoutableShards)
	return ok
}

func (s *Store) SetShardForWorkflow(workflowID string, shardID uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.routingState[workflowID] = shardID

	// Signal any waiting allocation requests
	if waiters, ok := s.pendingAllocs[workflowID]; ok {
		for _, ch := range waiters {
			select {
			case ch <- shardID:
			default:
			}
		}
		delete(s.pendingAllocs, workflowID)
	}
}

func (s *Store) SetRoutingState(state *pb.RoutingState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentState = state
}

func (s *Store) GetRoutingState() *pb.RoutingState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentState
}

// GetPendingAllocations returns workflow IDs that need allocation (non-blocking)
func (s *Store) GetPendingAllocations() []string {
	var pending []string
	for {
		select {
		case req := <-s.allocRequests:
			pending = append(pending, req.WorkflowID)
		default:
			return pending
		}
	}
}

func (s *Store) IsInTransition() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return !s.isInSteadyState()
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

	s.updateHealthyShards()
}

func (s *Store) SetAllShardHealth(health map[uint32]bool) {
	s.mu.Lock()
	s.shardHealth = make(map[uint32]bool)
	for k, v := range health {
		s.shardHealth[k] = v
	}
	s.mu.Unlock()

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

func (s *Store) GetHealthyShards() []uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.healthyShards)
}
