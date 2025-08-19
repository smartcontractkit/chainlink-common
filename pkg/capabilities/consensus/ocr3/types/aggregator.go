package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
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

// Encode serializes Metadata in contract order:
// 1B Version, 32B ExecutionID, 4B Timestamp, 4B DONID, 4B DONConfigVersion,
// 32B WorkflowID, 10B WorkflowName, 20B WorkflowOwner, 2B ReportID
func (m Metadata) Encode() ([]byte, error) {
	m.padWorkflowName()
	buf := new(bytes.Buffer)

	// 1) Version as a single byte
	if err := buf.WriteByte(byte(m.Version)); err != nil {
		return nil, err
	}

	// 2) Helper to decode a hex string and ensure length
	writeHex := func(field string, expectedBytes int) error {
		s := strings.TrimPrefix(field, "0x")
		b, err := hex.DecodeString(s)
		if err != nil {
			return fmt.Errorf("invalid hex in field: %w", err)
		}
		if len(b) != expectedBytes {
			return fmt.Errorf("wrong length: expected %d bytes, got %d", expectedBytes, len(b))
		}
		_, err = buf.Write(b)
		return err
	}

	// ExecutionID: 32 bytes
	if err := writeHex(m.ExecutionID, 32); err != nil {
		return nil, fmt.Errorf("ExecutionID: %w", err)
	}

	// Timestamp, DONID, DONConfigVersion—all 4‐byte big endian
	for _, v := range []uint32{m.Timestamp, m.DONID, m.DONConfigVersion} {
		if err := binary.Write(buf, binary.BigEndian, v); err != nil {
			return nil, err
		}
	}

	// WorkflowID: 32 bytes
	if err := writeHex(m.WorkflowID, 32); err != nil {
		return nil, fmt.Errorf("WorkflowID: %w", err)
	}

	// Workflow Name: 10 bytes
	if err := writeHex(m.WorkflowName, 10); err != nil {
		return nil, fmt.Errorf("WorkflowName: %w", err)
	}

	// WorkflowOwner: 20 bytes
	if err := writeHex(m.WorkflowOwner, 20); err != nil {
		return nil, fmt.Errorf("WorkflowOwner: %w", err)
	}

	// ReportID: 2 bytes
	if err := writeHex(m.ReportID, 2); err != nil {
		return nil, fmt.Errorf("ReportID: %w", err)
	}

	return buf.Bytes(), nil
}

// 1B Version, 32B ExecutionID, 4B Timestamp, 4B DONID, 4B DONConfigVersion,
// 32B WorkflowID, 10B WorkflowName, 20B WorkflowOwner, 2B ReportID
const MetadataLen = 1 + 32 + 4 + 4 + 4 + 32 + 10 + 20 + 2 // =109

// Decode parses exactly MetadataLen bytes from raw, returns a Metadata struct
// and any trailing data.
func Decode(raw []byte) (Metadata, []byte, error) {
	m := Metadata{}

	if len(raw) < MetadataLen {
		return m, nil, fmt.Errorf("metadata: raw too short, want ≥%d, got %d", MetadataLen, len(raw))
	}

	buf := bytes.NewReader(raw[:MetadataLen])

	// 1) Version (1 byte)
	var vb byte
	if err := binary.Read(buf, binary.BigEndian, &vb); err != nil {
		return m, nil, err
	}
	m.Version = uint32(vb)

	// helper to read N bytes and hex-decode
	readHex := func(n int) (string, error) {
		tmp := make([]byte, n)
		if _, err := io.ReadFull(buf, tmp); err != nil {
			return "", err
		}
		return hex.EncodeToString(tmp), nil
	}

	// 2) ExecutionID (32 bytes hex)
	var err error
	if m.ExecutionID, err = readHex(32); err != nil {
		return m, nil, fmt.Errorf("ExecutionID: %w", err)
	}

	// 3) Timestamp, DONID, DONConfigVersion (each 4 bytes BE)
	for _, ptr := range []*uint32{&m.Timestamp, &m.DONID, &m.DONConfigVersion} {
		if err := binary.Read(buf, binary.BigEndian, ptr); err != nil {
			return m, nil, err
		}
	}

	// 4) WorkflowID (32 bytes hex)
	if m.WorkflowID, err = readHex(32); err != nil {
		return m, nil, fmt.Errorf("WorkflowID: %w", err)
	}

	nameBytes := make([]byte, 10)
	if _, err := io.ReadFull(buf, nameBytes); err != nil {
		return m, nil, err
	}
	// hex-encode those 10 bytes into a 20-char string
	m.WorkflowName = hex.EncodeToString(nameBytes)

	// 6) WorkflowOwner (20 bytes hex)
	if m.WorkflowOwner, err = readHex(20); err != nil {
		return m, nil, fmt.Errorf("WorkflowOwner: %w", err)
	}

	// 7) ReportID (2 bytes hex)
	if m.ReportID, err = readHex(2); err != nil {
		return m, nil, fmt.Errorf("ReportID: %w", err)
	}

	// strip any stray "0x" prefixes just in case
	m.ExecutionID = strings.TrimPrefix(m.ExecutionID, "0x")
	m.WorkflowID = strings.TrimPrefix(m.WorkflowID, "0x")
	m.WorkflowOwner = strings.TrimPrefix(m.WorkflowOwner, "0x")
	m.ReportID = strings.TrimPrefix(m.ReportID, "0x")

	// the rest is payload
	tail := raw[MetadataLen:]
	return m, tail, nil
}

func (m Metadata) Length() int {
	return MetadataLen
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
