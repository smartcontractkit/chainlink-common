package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
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

func (m Metadata) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	// Write fixed-size fields using binary.Write
	if err := binary.Write(buf, binary.BigEndian, m.Version); err != nil {
		return nil, err
	}

	// Encode ExecutionID: convert hex string to bytes.
	execID, err := hex.DecodeString(m.ExecutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ExecutionID: %w", err)
	}
	if len(execID) != 32 {
		return nil, fmt.Errorf("expected ExecutionID to be 32 bytes after decoding, got %d", len(execID))
	}
	buf.Write(execID)

	// Write Timestamp, DONID, DONConfigVersion.
	if err := binary.Write(buf, binary.BigEndian, m.Timestamp); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, m.DONID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, m.DONConfigVersion); err != nil {
		return nil, err
	}

	// Encode WorkflowID.
	wID, err := hex.DecodeString(m.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode WorkflowID: %w", err)
	}
	if len(wID) != 32 {
		return nil, fmt.Errorf("expected WorkflowID to be 32 bytes after decoding, got %d", len(wID))
	}
	buf.Write(wID)

	// Encode WorkflowName: pad if necessary then decode (should be 20 hex characters = 10 bytes).
	m.padWorkflowName()
	wName, err := hex.DecodeString(m.WorkflowName)
	if err != nil {
		return nil, fmt.Errorf("failed to decode WorkflowName: %w", err)
	}
	if len(wName) != 10 {
		return nil, fmt.Errorf("expected WorkflowName to be 10 bytes after decoding, got %d", len(wName))
	}
	buf.Write(wName)

	// Encode WorkflowOwner (40 hex characters = 20 bytes).
	wOwner, err := hex.DecodeString(m.WorkflowOwner)
	if err != nil {
		return nil, fmt.Errorf("failed to decode WorkflowOwner: %w", err)
	}
	if len(wOwner) != 20 {
		return nil, fmt.Errorf("expected WorkflowOwner to be 20 bytes after decoding, got %d", len(wOwner))
	}
	buf.Write(wOwner)

	// Encode ReportID (4 hex characters = 2 bytes).
	reportID, err := hex.DecodeString(m.ReportID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ReportID: %w", err)
	}
	if len(reportID) != 2 {
		return nil, fmt.Errorf("expected ReportID to be 2 bytes after decoding, got %d", len(reportID))
	}
	buf.Write(reportID)

	return buf.Bytes(), nil
}

func (m Metadata) Length() int {
	bytes, err := m.Encode()
	if err != nil {
		return 0
	}
	return len(bytes)
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
