package dontime

import (
	"fmt"
	"sync"
	"time"
)

type Request struct {
	ExpiresAt time.Time

	// CallbackCh is a channel to send a response back to the requester
	// after the request has been processed or timed out.
	CallbackCh chan Response

	WorkflowExecutionID string
	SeqNum              int

	expiryTimer *time.Timer
	respondOnce sync.Once
}

func (r *Request) ID() string {
	return r.WorkflowExecutionID
}

func (r *Request) stopExpiry() {
	if r.expiryTimer != nil {
		r.expiryTimer.Stop()
	}
}

func (r *Request) SendResponse(resp Response) {
	r.respondOnce.Do(func() {
		r.stopExpiry()
		select {
		case r.CallbackCh <- resp:
		default: // Don't block trying to send
		}
		close(r.CallbackCh)
	})
}

func (r *Request) SendTimeout() {
	timeoutResponse := Response{
		WorkflowExecutionID: r.WorkflowExecutionID,
		SeqNum:              r.SeqNum,
		Err:                 fmt.Errorf("timeout exceeded: could not process request before expiry, workflowExecutionID %s", r.WorkflowExecutionID),
	}
	r.SendResponse(timeoutResponse)
}

type Response struct {
	WorkflowExecutionID string
	SeqNum              int
	Timestamp           int64
	Err                 error
}

func (r Response) RequestID() string {
	return r.WorkflowExecutionID
}
