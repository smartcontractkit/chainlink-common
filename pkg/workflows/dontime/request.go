package dontime

import (
	"context"
	"fmt"
	"time"
)

type Request struct {
	ExpiresAt time.Time

	// CallbackCh is a channel to send a response back to the requester
	// after the request has been processed or timed out.
	CallbackCh chan Response

	WorkflowExecutionID string
	SeqNum              int
}

func (r *Request) ID() string {
	return r.WorkflowExecutionID
}

func (r *Request) ExpiryTime() time.Time {
	return r.ExpiresAt
}

func (r *Request) SendResponse(ctx context.Context, resp Response) {
	select {
	case r.CallbackCh <- resp:
		close(r.CallbackCh)
	default:
		// Don't block if receiver not ready, but check if context is actually expired
		select {
		case <-ctx.Done():
			// Context cancelled or deadline exceeded before send
		default:
			// Try once more without blocking
		}
	}
}

func (r *Request) SendTimeout(ctx context.Context) {
	timeoutResponse := Response{
		WorkflowExecutionID: r.WorkflowExecutionID,
		SeqNum:              r.SeqNum,
		Err:                 fmt.Errorf("timeout exceeded: could not process request before expiry, workflowExecutionID %s", r.WorkflowExecutionID),
	}
	r.SendResponse(ctx, timeoutResponse)
}

func (r *Request) Copy() *Request {
	return &Request{
		ExpiresAt:           r.ExpiresAt,
		WorkflowExecutionID: r.WorkflowExecutionID,
		SeqNum:              r.SeqNum,

		// Intentionally not copied, but are thread-safe.
		CallbackCh: r.CallbackCh,
	}
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
