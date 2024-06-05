package ocr3

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type config struct {
	AggregationMethod string      `mapstructure:"aggregation_method" json:"aggregation_method" jsonschema:"enum=data_feeds"`
	AggregationConfig *values.Map `mapstructure:"aggregation_config" json:"aggregation_config"`
	Encoder           string      `mapstructure:"encoder" json:"encoder"`
	EncoderConfig     *values.Map `mapstructure:"encoder_config" json:"encoder_config"`
	ReportID          string      `mapstructure:"report_id" json:"report_id" jsonschema:"required, pattern=^[a-f0-9]{4}$"`
}

// TODO: KS-83 remove this once we have a proper tests for the json schema validation
func (c *config) validate() error {
	_, err := hex.DecodeString(c.ReportID)
	if err != nil {
		return fmt.Errorf("report_id must be a hex string that represents 2 bytes: %w", err)
	}
	if len(c.ReportID) != 4 {
		return fmt.Errorf("report_id must be a hex string that represents 2 bytes: expected 4 characters, got %d", len(c.ReportID))
	}
	return nil
}

type inputs struct {
	Observations *values.List `json:"observations"`
}

type outputs struct {
	WorkflowExecutionID string
	capabilities.CapabilityResponse
}

type request struct {
	Observations *values.List `mapstructure:"-"`
	ExpiresAt    time.Time

	// CallbackCh is a channel to send a response back to the requester
	// after the request has been processed or timed out.
	CallbackCh chan capabilities.CapabilityResponse
	StopCh     services.StopChan

	WorkflowExecutionID string
	WorkflowID          string
	WorkflowOwner       string
	WorkflowName        string
	WorkflowDonID       string
	ReportID            string
}
