package dontime

import (
	"fmt"
	"maps"
	"sync"
	"time"
)

// This timeout is used to remove the request from the pending queue inside the Store object.
// Having a short timeout here comes with a risk of affecting all future sequence numbers
// for the same executionID, if the plugin is not able to process the request fast enough.
// For that reason, we keep this timeout relatively long but allow the Engine to time out
// much faster when waiting on the response channel.
var DefaultRequestTimeout = 10 * time.Minute

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
	maps.Copy(copied, s.requests)
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

	expiresAt := time.Now().Add(s.requestTimeout)
	s.requests[executionID] = &Request{
		ExpiresAt:           expiresAt,
		CallbackCh:          ch,
		WorkflowExecutionID: executionID,
		SeqNum:              seqNum,
		expiryTimer: time.AfterFunc(time.Until(expiresAt), func() {
			s.expireRequest(executionID)
		}),
	}
	return ch
}

func (s *Store) expireRequest(executionID string) {
	s.mu.Lock()
	req := s.removeRequestLocked(executionID)
	s.mu.Unlock()
	if req == nil {
		return
	}
	req.SendTimeout()
}

func (s *Store) removeRequestLocked(executionID string) *Request {
	req, ok := s.requests[executionID]
	if !ok {
		return nil
	}
	delete(s.requests, executionID)
	req.stopExpiry()
	return req
}

func (s *Store) RemoveRequest(executionID string) {
	s.mu.Lock()
	s.removeRequestLocked(executionID)
	s.mu.Unlock()
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

func (s *Store) replaceDonTimes(donTimes map[string][]int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	maps.Copy(s.donTimes, donTimes)

	for executionID := range s.donTimes {
		if _, ok := donTimes[executionID]; !ok {
			delete(s.donTimes, executionID)
			s.removeRequestLocked(executionID)
		}
	}
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
	s.removeRequestLocked(executionID)
}
