package requests_test

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// testRequest is a fixture implementation of the Store/Handler request
// contract. It mirrors the ReportRequest type previously defined in the
// removed v1 ocr3 consensus capability.
type testRequest struct {
	Observations            *values.List
	OverriddenEncoderName   string
	OverriddenEncoderConfig *values.Map
	ExpiresAt               time.Time

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
		Observations:            r.Observations.CopyList(),
		OverriddenEncoderConfig: r.OverriddenEncoderConfig.CopyMap(),

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
