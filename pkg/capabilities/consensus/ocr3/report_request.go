package ocr3

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type ReportRequest struct {
	Observations            *values.List `mapstructure:"-"`
	OverriddenEncoderName   string
	OverriddenEncoderConfig *values.Map
	ExpiresAt               time.Time

	// CallbackCh is a channel to send a response back to the requester
	// after the request has been processed or timed out.
	CallbackCh chan ReportResponse
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

func (r *ReportRequest) ID() string {
	return r.WorkflowExecutionID
}

func (r *ReportRequest) ExpiryTime() time.Time {
	return r.ExpiresAt
}

func (r *ReportRequest) SendResponse(ctx context.Context, resp ReportResponse) {
	select {
	case <-ctx.Done():
		return
	case r.CallbackCh <- resp:
		close(r.CallbackCh)
	}
}

func (r *ReportRequest) SendTimeout(ctx context.Context) {
	timeoutResponse := ReportResponse{
		WorkflowExecutionID: r.WorkflowExecutionID,
		Err:                 fmt.Errorf("timeout exceeded: could not process request before expiry, workflowExecutionID %s", r.WorkflowExecutionID),
	}
	r.SendResponse(ctx, timeoutResponse)
}

func (r *ReportRequest) Copy() *ReportRequest {
	return &ReportRequest{
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

type ReportResponse struct {
	WorkflowExecutionID string
	Value               *values.Map
	Err                 error
}

func (r ReportResponse) RequestID() string {
	return r.WorkflowExecutionID
}
