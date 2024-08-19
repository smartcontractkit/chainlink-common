package requests

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Store struct {
	requestIDs []string
	requests   map[string]*Request

	mu sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		requestIDs: []string{},
		requests:   map[string]*Request{},
	}
}

// GetByIDs is best-effort, doesn't return requests that are not in store
func (s *Store) GetByIDs(ctx context.Context, requestIDs []string) []*Request {
	s.mu.RLock()
	defer s.mu.RUnlock()

	o := []*Request{}
	for _, r := range requestIDs {
		gr, ok := s.requests[r]
		if ok {
			o = append(o, gr)
		}
	}

	return o
}

func (s *Store) FirstN(ctx context.Context, batchSize int) ([]*Request, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if batchSize == 0 {
		return nil, errors.New("batchsize cannot be 0")
	}
	got := []*Request{}
	if len(s.requestIDs) == 0 {
		return got, nil
	}

	for _, r := range s.requestIDs {
		gr, ok := s.requests[r]
		if !ok {
			continue
		}

		got = append(got, gr)
		if len(got) == batchSize {
			break
		}
	}

	return got, nil
}

func (s *Store) Add(req *Request) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.requests[req.WorkflowExecutionID]; ok {
		return fmt.Errorf("request with id %s already exists", req.WorkflowExecutionID)
	}
	s.requestIDs = append(s.requestIDs, req.WorkflowExecutionID)
	s.requests[req.WorkflowExecutionID] = req
	return nil
}

func (s *Store) Get(requestID string) *Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.requests[requestID]
}

func (s *Store) evict(requestID string) (*Request, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var found bool

	r, ok := s.requests[requestID]
	if ok {
		found = true
		delete(s.requests, requestID)
	}

	newRequestIDs := []string{}
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
