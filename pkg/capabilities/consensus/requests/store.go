package requests

import (
	"errors"
	"fmt"
	"sync"
)

type statsCollector interface {
	RequestAdded()
	RequestEvicted()
}

type noopsStatsCollector struct{}

func (n *noopsStatsCollector) RequestAdded()   {}
func (n *noopsStatsCollector) RequestEvicted() {}

type StoredRequest[T any] interface {
	ID() string
	Copy() T
}

// Store is a generic store for ongoing consensus requests.
// It is thread-safe and uses a map to store requests.
type Store[T StoredRequest[T]] struct {
	requestIDs []string
	requests   map[string]T

	mu sync.RWMutex

	statsCollector statsCollector
}

func NewStore[T StoredRequest[T]]() *Store[T] {
	return &Store[T]{
		requestIDs:     []string{},
		requests:       map[string]T{},
		statsCollector: &noopsStatsCollector{},
	}
}

func NewStoreWithStatsCollector[T StoredRequest[T]](statsCollector statsCollector) *Store[T] {
	return &Store[T]{
		requestIDs:     []string{},
		requests:       map[string]T{},
		statsCollector: statsCollector,
	}
}

// GetByIDs retrieves requests by their IDs.
// The method deep-copies requests before returning them.
func (s *Store[T]) GetByIDs(requestIDs []string) []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	o := []T{}
	for _, r := range requestIDs {
		gr, ok := s.requests[r]
		if ok {
			o = append(o, gr.Copy())
		}
	}

	return o
}

// FirstN retrieves up to `batchSize` requests.
// The method deep-copies requests before returning them.
func (s *Store[T]) FirstN(batchSize int) ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if batchSize == 0 {
		return nil, errors.New("batchsize cannot be 0")
	}
	got := []T{}
	if len(s.requestIDs) == 0 {
		return got, nil
	}

	for _, r := range s.requestIDs {
		gr, ok := s.requests[r]
		if !ok {
			continue
		}

		got = append(got, gr.Copy())
		if len(got) == batchSize {
			break
		}
	}

	return got, nil
}

// RangeN retrieves up to `batchSize` requests starting at index `start`.
// It deep-copies each request before returning.
func (s *Store[T]) RangeN(start, batchSize int) ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if start < 0 {
		return nil, errors.New("start must be non-negative")
	}
	if batchSize <= 0 {
		return nil, errors.New("batchSize must be greater than 0")
	}
	if start >= len(s.requestIDs) {
		return nil, fmt.Errorf("start index out of bounds: start=%d, len=%d", start, len(s.requestIDs))
	}

	end := start + batchSize
	if end > len(s.requestIDs) {
		end = len(s.requestIDs)
	}

	got := make([]T, 0, end-start)
	for _, r := range s.requestIDs[start:end] {
		gr, ok := s.requests[r]
		if !ok {
			continue
		}
		got = append(got, gr.Copy())
	}
	return got, nil
}

func (s *Store[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.requestIDs)
}

// Add adds a new request to the store.
func (s *Store[T]) Add(req T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.requests[req.ID()]; ok {
		return fmt.Errorf("request with id %s already exists", req.ID())
	}
	s.requestIDs = append(s.requestIDs, req.ID())
	s.requests[req.ID()] = req
	s.statsCollector.RequestAdded()
	return nil
}

// Get retrieves a request by its ID.
// The method deep-copies the request before returning it.
func (s *Store[T]) Get(requestID string) T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rid, ok := s.requests[requestID]
	if ok {
		return rid.Copy()
	}
	var zero T
	return zero
}

// Evict removes a request from the store by its ID.
func (s *Store[T]) Evict(requestID string) (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var found bool

	r, ok := s.requests[requestID]
	if ok {
		found = true
		delete(s.requests, requestID)
		s.statsCollector.RequestEvicted()
	}

	newRequestIDs := make([]string, 0, len(s.requestIDs))
	for _, rid := range s.requestIDs {
		if rid != requestID {
			newRequestIDs = append(newRequestIDs, rid)
		} else {
			found = true
		}
	}

	s.requestIDs = newRequestIDs
	return r, found
}
