package requests_test

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// testRequest is a minimal implementation of requests.ConsensusRequest used to
// exercise the store and the handler.
type testRequest struct {
	Observations *values.List
	ExpiresAt    time.Time

	// CallbackCh is a channel to send a response back to the requester
	// after the request has been processed or timed out.
	CallbackCh chan testResponse
	StopCh     services.StopChan

	WorkflowExecutionID      string
	WorkflowID               string
	WorkflowOwner            string
	WorkflowName             string
	WorkflowDonID            uint32
	WorkflowDonConfigVersion uint32
	ReportID                 string

	KeyID string
}

func (r *testRequest) ID() string {
	return r.WorkflowExecutionID
}

func (r *testRequest) ExpiryTime() time.Time {
	return r.ExpiresAt
}

func (r *testRequest) SendResponse(ctx context.Context, resp testResponse) {
	select {
	case <-ctx.Done():
		return
	case r.CallbackCh <- resp:
		close(r.CallbackCh)
	}
}

func (r *testRequest) SendTimeout(ctx context.Context) {
	timeoutResponse := testResponse{
		WorkflowExecutionID: r.WorkflowExecutionID,
		Err:                 fmt.Errorf("timeout exceeded: could not process request before expiry, workflowExecutionID %s", r.WorkflowExecutionID),
	}
	r.SendResponse(ctx, timeoutResponse)
}

func (r *testRequest) Copy() *testRequest {
	return &testRequest{
		Observations: r.Observations.CopyList(),

		// No need to copy these, they're value types.
		ExpiresAt:                r.ExpiresAt,
		WorkflowExecutionID:      r.WorkflowExecutionID,
		WorkflowID:               r.WorkflowID,
		WorkflowName:             r.WorkflowName,
		WorkflowOwner:            r.WorkflowOwner,
		WorkflowDonID:            r.WorkflowDonID,
		WorkflowDonConfigVersion: r.WorkflowDonConfigVersion,
		ReportID:                 r.ReportID,
		KeyID:                    r.KeyID,

		// Intentionally not copied, but are thread-safe.
		CallbackCh: r.CallbackCh,
		StopCh:     r.StopCh,
	}
}

type testResponse struct {
	WorkflowExecutionID string
	Value               *values.Map
	Err                 error
}

func (r testResponse) RequestID() string {
	return r.WorkflowExecutionID
}
