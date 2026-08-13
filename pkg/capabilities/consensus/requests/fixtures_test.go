package requests_test

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// testReportRequest/testReportResponse stand in for a real ConsensusRequest/ConsensusResponse
// implementation (e.g. ocr3.ReportRequest/ReportResponse) so the generic Store/Handler can be
// exercised without depending on a specific capability.
type testReportRequest struct {
	Observations            *values.List
	OverriddenEncoderName   string
	OverriddenEncoderConfig *values.Map
	ExpiresAt               time.Time

	// CallbackCh is a channel to send a response back to the requester
	// after the request has been processed or timed out.
	CallbackCh chan testReportResponse
	StopCh     chan struct{}

	WorkflowExecutionID      string
	WorkflowID               string
	WorkflowOwner            string
	WorkflowName             string
	WorkflowDonID            uint32
	WorkflowDonConfigVersion uint32
	ReportID                 string

	KeyID string
}

func (r *testReportRequest) ID() string {
	return r.WorkflowExecutionID
}

func (r *testReportRequest) ExpiryTime() time.Time {
	return r.ExpiresAt
}

func (r *testReportRequest) SendResponse(ctx context.Context, resp testReportResponse) {
	select {
	case <-ctx.Done():
		return
	case r.CallbackCh <- resp:
		close(r.CallbackCh)
	}
}

func (r *testReportRequest) SendTimeout(ctx context.Context) {
	r.SendResponse(ctx, testReportResponse{
		WorkflowExecutionID: r.WorkflowExecutionID,
		Err:                 fmt.Errorf("timeout exceeded: could not process request before expiry, workflowExecutionID %s", r.WorkflowExecutionID),
	})
}

func (r *testReportRequest) Copy() *testReportRequest {
	return &testReportRequest{
		Observations:            r.Observations.CopyList(),
		OverriddenEncoderConfig: r.OverriddenEncoderConfig.CopyMap(),

		// No need to copy these, they're value types.
		OverriddenEncoderName:    r.OverriddenEncoderName,
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

type testReportResponse struct {
	WorkflowExecutionID string
	Value               *values.Map
	Err                 error
}

func (r testReportResponse) RequestID() string {
	return r.WorkflowExecutionID
}
