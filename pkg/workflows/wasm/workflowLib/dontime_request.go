package workflowLib

import (
	"context"
	"fmt"
	"time"
)

type DonTimeRequest struct {
	ExpiresAt time.Time

	// CallbackCh is a channel to send a response back to the requester
	// after the request has been processed or timed out.
	CallbackCh chan DonTimeResponse

	WorkflowExecutionID string
	SeqNum              int
}

func (r *DonTimeRequest) ID() string {
	return r.WorkflowExecutionID
}

func (r *DonTimeRequest) ExpiryTime() time.Time {
	return r.ExpiresAt
}

func (r *DonTimeRequest) SendResponse(ctx context.Context, resp DonTimeResponse) {
	select {
	case <-ctx.Done():
		return
	case r.CallbackCh <- resp:
		close(r.CallbackCh)
	}
}

func (r *DonTimeRequest) SendTimeout(ctx context.Context) {
	timeoutResponse := DonTimeResponse{
		WorkflowExecutionID: r.WorkflowExecutionID,
		seqNum:              r.SeqNum,
		Err:                 fmt.Errorf("timeout exceeded: could not process request before expiry, workflowExecutionID %s", r.WorkflowExecutionID),
	}
	r.SendResponse(ctx, timeoutResponse)
}

func (r *DonTimeRequest) Copy() *DonTimeRequest {
	return &DonTimeRequest{
		ExpiresAt:           r.ExpiresAt,
		WorkflowExecutionID: r.WorkflowExecutionID,
		SeqNum:              r.SeqNum,

		// Intentionally not copied, but are thread-safe.
		CallbackCh: r.CallbackCh,
	}
}

type DonTimeResponse struct {
	WorkflowExecutionID string
	seqNum              int
	timestamp           int64
	Err                 error
}

func (r DonTimeResponse) RequestID() string {
	return r.WorkflowExecutionID
}
