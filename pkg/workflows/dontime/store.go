package dontime

import (
	"fmt"
	"sync"
	"time"
)

var DefaultRequestTimeout = 20 * time.Minute

type Store struct {
	requests       map[string]*Request // Maps workflow execution ID to request
	requestTimeout time.Duration

	// donTimes holds ordered sequence timestamps generated for consecutive workflow requests
	// i.e. ExecutionID --> [timestamp-0, timestamp-1 , ...]
	donTimes            map[string][]int64
	lastObservedDonTime int64
	mu                  sync.Mutex
}

func NewStore(requestTimeout time.Duration) *Store {
	return &Store{
		requests:            make(map[string]*Request),
		requestTimeout:      requestTimeout,
		donTimes:            make(map[string][]int64),
		lastObservedDonTime: 0,
		mu:                  sync.Mutex{},
	}
}

func (s *Store) GetRequest(executionID string) *Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.requests[executionID]
}

func (s *Store) GetRequests() map[string]*Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make(map[string]*Request)
	for k, v := range s.requests {
		copied[k] = v
	}
	return copied
}

// RequestDonTime adds a don time request to the queue or return the dontime if we have it yet.
func (s *Store) RequestDonTime(executionID string, seqNum int) <-chan Response {
	ch := make(chan Response, 1)
	dontime := s.GetDonTimeForSeqNum(executionID, seqNum)
	if dontime != nil {
		ch <- Response{
			WorkflowExecutionID: executionID,
			SeqNum:              seqNum,
			Timestamp:           *dontime,
			Err:                 nil,
		}
		close(ch)
		return ch
	}

	// Submit request and return channel
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, alreadyExists := s.requests[executionID]; alreadyExists {
		ch <- Response{
			WorkflowExecutionID: executionID,
			SeqNum:              seqNum,
			Timestamp:           0,
			Err: fmt.Errorf(
				"DON Time request for executionID=%s already exists (sequenceNumber=%d)",
				executionID, seqNum),
		}
		close(ch)
		return ch
	}

	s.requests[executionID] = &Request{
		ExpiresAt:           time.Now().Add(s.requestTimeout),
		CallbackCh:          ch,
		WorkflowExecutionID: executionID,
		SeqNum:              seqNum,
	}
	return ch
}

func (s *Store) RemoveRequest(executionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.requests, executionID)
}

func (s *Store) GetDonTimeForSeqNum(executionID string, seqNum int) *int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if times, ok := s.donTimes[executionID]; ok {
		if len(times) > seqNum {
			return &times[seqNum]
		}
	}
	return nil
}

func (s *Store) GetDonTimes(executionID string) ([]int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if times, ok := s.donTimes[executionID]; ok {
		return times, nil
	}
	return []int64{}, fmt.Errorf("no don time for executionID %s", executionID)
}

func (s *Store) setDonTimes(executionID string, donTimes []int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.donTimes[executionID] = donTimes
}

func (s *Store) GetLastObservedDonTime() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastObservedDonTime
}

func (s *Store) setLastObservedDonTime(observedDonTime int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastObservedDonTime = observedDonTime
}

func (s *Store) deleteExecutionID(executionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.donTimes, executionID)
	delete(s.requests, executionID)
}

func (s *Store) deleteExecutionIDs(executionIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range executionIDs {
		delete(s.donTimes, id)
		delete(s.requests, id)
	}
}
