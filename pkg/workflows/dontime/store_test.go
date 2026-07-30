package dontime

import (
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
