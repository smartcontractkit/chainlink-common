package dontime

import (
	"fmt"
	"sync"
	"time"

	consensusRequests "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
)

var (
	// donTimeStore is a singleton which can be accessed by anyone who needs it
	donTimeStore          *Store
	once                  sync.Once
	defaultRequestTimeout = 20 * time.Minute
)

func GetDonTimeStore() *Store {
	once.Do(func() {
		donTimeStore = NewDonTimeStore(defaultRequestTimeout)
	})
	return donTimeStore
}

type Store struct {
	requests       *consensusRequests.Store[*Request, DonTimeResponse]
	requestTimeout time.Duration

	finishedExecutionIDs map[string]bool
	donTimes             map[string][]int64 // ExecutionID --> [timestamp-0, timestamp-1 , ...]
	lastObservedDonTime  int64
	mu                   sync.Mutex
}

func NewDonTimeStore(requestTimeout time.Duration) *Store {
	return &Store{
		requests:             consensusRequests.NewStore[*Request, DonTimeResponse](),
		requestTimeout:       requestTimeout,
		finishedExecutionIDs: make(map[string]bool),
		donTimes:             make(map[string][]int64),
		lastObservedDonTime:  0,
		mu:                   sync.Mutex{},
	}
}

// ExecutionFinished marks a workflow execution as finished for this node
// Once consensus is reached that the execution has finished, the executionID
// will be marked for deletion after some time.
func (s *Store) ExecutionFinished(executionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.finishedExecutionIDs[executionID] = true
}

func (s *Store) GetFinishedExecutionIDs() map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.finishedExecutionIDs
}

// RequestDonTime adds a don time request to the queue or return the dontime if we have it yet.
func (s *Store) RequestDonTime(executionID string, seqNum int) <-chan DonTimeResponse {
	ch := make(chan DonTimeResponse, 1)
	dontime := s.GetDonTimeForSeqNum(executionID, seqNum)
	if dontime != nil {
		ch <- DonTimeResponse{
			WorkflowExecutionID: executionID,
			SeqNum:              seqNum,
			Timestamp:           *dontime,
			Err:                 nil,
		}
		close(ch)
		return ch
	}

	// Submit request and return channel
	err := s.requests.Add(&Request{
		ExpiresAt:           time.Now().Add(s.requestTimeout),
		CallbackCh:          ch,
		WorkflowExecutionID: executionID,
		SeqNum:              seqNum,
	})
	if err != nil {
		ch <- DonTimeResponse{
			WorkflowExecutionID: executionID,
			SeqNum:              seqNum,
			Timestamp:           0,
			Err: fmt.Errorf(
				"failed to queue DON Time request (executionID=%s, sequenceNumber=%d): %w",
				executionID, seqNum, err),
		}
		close(ch)
	}
	return ch
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

func (s *Store) SetDonTimes(executionID string, donTimes []int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.donTimes[executionID] = donTimes
}

func (s *Store) GetLastObservedDonTime() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastObservedDonTime
}

func (s *Store) SetLastObservedDonTime(observedDonTime int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastObservedDonTime = observedDonTime
}

func (s *Store) deleteExecutionID(executionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.donTimes, executionID)
	delete(s.finishedExecutionIDs, executionID)
}
