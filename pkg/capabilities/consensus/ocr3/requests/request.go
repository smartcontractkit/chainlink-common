package requests

import (
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type Request struct {
	Observations            *values.List `mapstructure:"-"`
	OverriddenEncoderName   string
	OverriddenEncoderConfig *values.Map
	ExpiresAt               time.Time

	// CallbackCh is a channel to send a response back to the requester
	// after the request has been processed or timed out.
	CallbackCh chan Response
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

func (r *Request) Copy() *Request {
	return &Request{
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

type Response struct {
	WorkflowExecutionID string
	Value               *values.Map
	Err                 error
}
