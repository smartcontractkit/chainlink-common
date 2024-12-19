package types

import (
	"encoding/hex"
	"strings"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const MetadataFieldName = "INTERNAL_METADATA"

type Metadata struct {
	Version          uint32 //  1 byte
	ExecutionID      string // 32 hex bytes
	Timestamp        uint32 //  4 bytes
	DONID            uint32 //  4 bytes
	DONConfigVersion uint32 //  4 bytes
	WorkflowID       string // 32 hex bytes
	WorkflowName     string // 10 hex bytes
	WorkflowOwner    string // 20 hex bytes
	ReportID         string //  2 hex bytes
}

// the contract requires exactly 10 bytes for the workflow name
// the resulting workflow name should be up to 10 bytes long
// so pad accordingly to meet the contract requirements
func (m *Metadata) padWorkflowName() {
	b, err := hex.DecodeString(m.WorkflowName)
	if err == nil && len(b) < 10 {
		// Each byte is 2 characters, so we need to pad with 0s
		neededBytes := append(b, make([]byte, 10-len(b))...)
		m.WorkflowName = hex.EncodeToString(neededBytes)
	} else if len(m.WorkflowName) < 10 {
		// Pad with spaces
		suffix := strings.Repeat(" ", 10-len(m.WorkflowName))
		m.WorkflowName += suffix
	}
}

type Aggregator interface {
	// Called by the Outcome() phase of OCR reporting.
	// The inner array of observations corresponds to elements listed in "inputs.observations" section.
	Aggregate(lggr logger.Logger, previousOutcome *AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*AggregationOutcome, error)
}

func AppendMetadata(outcome *AggregationOutcome, meta *Metadata) (*AggregationOutcome, error) {
	meta.padWorkflowName()
	metaWrapped, err := values.Wrap(meta)
	if err != nil {
		return nil, err
	}
	outcome.EncodableOutcome.Fields[MetadataFieldName] = values.Proto(metaWrapped)
	return outcome, nil
}

type AggregatorFactory func(name string, config values.Map, lggr logger.Logger) (Aggregator, error)
