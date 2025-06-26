package workflowLib

import (
	"fmt"
	"sync"
	"time"

	consensusRequests "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
)

type DonTimeStore struct {
	Requests       *consensusRequests.Store[*DonTimeRequest, DonTimeResponse]
	requestTimeout time.Duration

	finishedExecutionIDs map[string]bool
	donTimes             map[string][]int64 // ExecutionID --> [timestamp-0, timestamp-1 , ...]
	lastObservedDonTime  int64
	mu                   sync.Mutex
}

func NewDonTimeStore(requestTimeout time.Duration) *DonTimeStore {
	return &DonTimeStore{
		Requests:             consensusRequests.NewStore[*DonTimeRequest, DonTimeResponse](),
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
func (s *DonTimeStore) ExecutionFinished(executionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.finishedExecutionIDs[executionID] = true
}

func (s *DonTimeStore) GetFinishedExecutionIDs() map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.finishedExecutionIDs
}

// RequestDonTime adds a don time request to the queue or return the dontime if we have it yet.
func (s *DonTimeStore) RequestDonTime(executionID string, seqNum int) <-chan DonTimeResponse {
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
	err := s.Requests.Add(&DonTimeRequest{
		ExpiresAt:           time.Now().Add(s.requestTimeout),
		CallbackCh:          ch,
		WorkflowExecutionID: executionID,
		SeqNum:              seqNum,
	})
	if err != nil {
		DonTimeRequestErr := fmt.Errorf(
			"failed to queue DON Time request (executionID=%s, sequenceNumber=%d): %w",
			executionID, seqNum, err)
		ch <- DonTimeResponse{
			WorkflowExecutionID: executionID,
			SeqNum:              seqNum,
			Timestamp:           0,
			Err:                 DonTimeRequestErr,
		}
	}
	return ch
}

func (s *DonTimeStore) GetDonTimeForSeqNum(executionID string, seqNum int) *int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if times, ok := s.donTimes[executionID]; ok {
		if len(times) > seqNum {
			return &times[seqNum]
		}
	}
	return nil
}

func (s *DonTimeStore) GetDonTimes(executionID string) ([]int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if times, ok := s.donTimes[executionID]; ok {
		return times, nil
	}
	return []int64{}, fmt.Errorf("no don time for executionID %s", executionID)
}

func (s *DonTimeStore) SetDonTimes(executionID string, donTimes []int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.donTimes[executionID] = donTimes
}

func (s *DonTimeStore) GetLastObservedDonTime() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastObservedDonTime
}

func (s *DonTimeStore) SetLastObservedDonTime(observedDonTime int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastObservedDonTime = observedDonTime
}

func (s *DonTimeStore) deleteExecutionID(executionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.donTimes, executionID)
	delete(s.finishedExecutionIDs, executionID)
}
