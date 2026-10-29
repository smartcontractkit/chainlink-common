package dontime

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStore_RequestExpiresWithoutPlugin(t *testing.T) {
	store := NewStore(50 * time.Millisecond)
	executionID := "workflow-123"

	timeRequest := store.RequestDonTime(executionID, 0)

	select {
	case resp := <-timeRequest:
		require.Equal(t, executionID, resp.WorkflowExecutionID)
		require.Equal(t, 0, resp.SeqNum)
		require.ErrorContains(t, resp.Err, "timeout exceeded: could not process request before expiry")
	case <-time.After(time.Second):
		t.Fatal("request did not expire")
	}

	require.Nil(t, store.GetRequest(executionID))
}

func (s *Store) GetDonTimes(executionID string) (map[int64]int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if times, ok := s.donTimes[executionID]; ok {
		return times, nil
	}
	return map[int64]int64{}, fmt.Errorf("no don time for executionID %s", executionID)
}

func (s *Store) setDonTimes(executionID string, donTimes map[int64]int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.donTimes[executionID] = donTimes
}
