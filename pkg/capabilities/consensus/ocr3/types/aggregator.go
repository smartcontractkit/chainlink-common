package types

import (
	"bytes"
	"encoding/binary"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const MetadataFieldName = "INTERNAL_METADATA"

func (rm Metadata) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, rm)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (rm Metadata) Length() int {
	bytes, err := rm.Encode()
	if err != nil {
		return 0
	}
	return len(bytes)
}

// Note: This should be a shared type that the OCR3 package validates as well
type Metadata struct {
	Version             uint8
	WorkflowExecutionID [32]byte
	Timestamp           uint32
	DonID               uint32
	DonConfigVersion    uint32
	WorkflowCID         [32]byte
	WorkflowName        [10]byte
	WorkflowOwner       [20]byte
	ReportID            [2]byte
}

// Aggregator is the interface that enables a hook to the Outcome() phase of OCR reporting.
type Aggregator interface {
	// Called by the Outcome() phase of OCR reporting.
	// The inner array of observations corresponds to elements listed in "inputs.observations" section.
	Aggregate(lggr logger.Logger, previousOutcome *AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*AggregationOutcome, error)
}

func AppendMetadata(outcome *AggregationOutcome, meta *Metadata) (*AggregationOutcome, error) {
	metaWrapped, err := values.Wrap(meta)
	if err != nil {
		return nil, err
	}
	outcome.EncodableOutcome.Fields[MetadataFieldName] = values.Proto(metaWrapped)
	return outcome, nil
}

type AggregatorFactory func(name string, config values.Map, lggr logger.Logger) (Aggregator, error)
