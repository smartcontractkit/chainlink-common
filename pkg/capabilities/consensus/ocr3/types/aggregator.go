package types

import (
	"strings"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const MetadataFieldName = "INTERNAL_METADATA"

type Metadata struct {
	Version          uint32 //  1 byte
	ExecutionID      string // 32 hex bytes (string len  = 64)
	Timestamp        uint32 //  4 bytes
	DONID            uint32 //  4 bytes
	DONConfigVersion uint32 //  4 bytes
	WorkflowID       string // 32 hex bytes (string len = 64)
	WorkflowName     string // 10 hex bytes (string len = 20)
	WorkflowOwner    string // 20 hex bytes (string len = 40)
	ReportID         string //  2 hex bytes (string len = 4)
}

// the contract requires exactly 10 bytes for the workflow name
// the resulting workflow name should be up to 10 bytes long
// so pad accordingly to meet the contract requirements
func (m *Metadata) padWorkflowName() {
	// it should have 10 hex bytes, so 20 characters total
	if len(m.WorkflowName) < 20 {
		suffix := strings.Repeat("0", 20-len(m.WorkflowName))
		m.WorkflowName += suffix
	}
}

// Aggregator is the interface that enables a hook to the Outcome() phase of OCR reporting.
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
