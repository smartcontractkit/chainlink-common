package dontime

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequest_SendResponseAndTimeoutOnce(t *testing.T) {
	ch := make(chan Response, 1)
	req := &Request{
		CallbackCh:          ch,
		WorkflowExecutionID: "exec-1",
		SeqNum:              0,
	}
	resp := Response{WorkflowExecutionID: "exec-1", SeqNum: 0, Timestamp: 123}

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			req.SendResponse(resp)
		})
		wg.Go(func() {
			req.SendTimeout()
		})
	}
	wg.Wait()

	got, ok := <-ch
	require.True(t, ok)
	require.Equal(t, "exec-1", got.WorkflowExecutionID)
	require.Equal(t, 0, got.SeqNum)

	_, ok = <-ch
	require.False(t, ok, "channel should be closed after first response")
}
