package ocr3

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	consensustypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type PassthroughEncoder struct{}

// Brought in from CL core.
type ReportV1Metadata struct {
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

func (rm ReportV1Metadata) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, rm)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (rm ReportV1Metadata) Length() int {
	bytes, err := rm.Encode()
	if err != nil {
		return 0
	}
	return len(bytes)
}

func (v PassthroughEncoder) Encode(ctx context.Context, input values.Map) ([]byte, error) {
	metaMap, ok := input.Underlying[consensustypes.MetadataFieldName]
	if !ok {
		return nil, fmt.Errorf("expected metadata field to be present: %s", consensustypes.MetadataFieldName)
	}

	dataMap, ok := input.Underlying["0"]
	if !ok {
		return nil, fmt.Errorf("expected metadata field to be present: %s", "0")
	}

	var meta consensustypes.Metadata
	err := metaMap.UnwrapTo(&meta)
	if err != nil {
		return nil, err
	}

	var data []byte
	err = dataMap.UnwrapTo(&data)
	if err != nil {
		return nil, err
	}

	return prependMetadataFields(meta, data)
}

func prependMetadataFields(meta consensustypes.Metadata, userPayload []byte) ([]byte, error) {
	var err error
	var result []byte

	// 1. Version (1 byte)
	if meta.Version > 255 {
		return nil, fmt.Errorf("version must be between 0 and 255")
	}
	result = append(result, byte(meta.Version))

	// 2. Execution ID (32 bytes)
	if result, err = decodeAndAppend(meta.ExecutionID, 32, result, "ExecutionID"); err != nil {
		return nil, err
	}

	// 3. Timestamp (4 bytes)
	tsBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(tsBytes, meta.Timestamp)
	result = append(result, tsBytes...)

	// 4. DON ID (4 bytes)
	donIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(donIDBytes, meta.DONID)
	result = append(result, donIDBytes...)

	// 5. DON config version (4 bytes)
	cfgVersionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(cfgVersionBytes, meta.DONConfigVersion)
	result = append(result, cfgVersionBytes...)

	// 6. Workflow ID / spec hash (32 bytes)
	if result, err = decodeAndAppend(meta.WorkflowID, 32, result, "WorkflowID"); err != nil {
		return nil, err
	}

	// 7. Workflow Name (7 bytes but we pad to 10)
	if result, err = decodeAndAppend(meta.WorkflowName, 10, result, "WorkflowName"); err != nil {
		return nil, err
	}

	// 8. Workflow Owner (20 bytes)
	if result, err = decodeAndAppend(meta.WorkflowOwner, 20, result, "WorkflowOwner"); err != nil {
		return nil, err
	}

	// 9. Report ID (2 bytes)
	if result, err = decodeAndAppend(meta.ReportID, 2, result, "ReportID"); err != nil {
		return nil, err
	}

	return append(result, userPayload...), nil
}

func decodeAndAppend(id string, expectedLen int, prevResult []byte, logName string) ([]byte, error) {
	b, err := hex.DecodeString(id)
	if err != nil {
		return nil, fmt.Errorf("failed to hex-decode %s (%s): %w", logName, id, err)
	}
	if len(b) > expectedLen {
		return nil, fmt.Errorf("incorrect length for id %s (%s), expected at most %d bytes, got %d", logName, id, expectedLen, len(b))
	}
	if len(b) < expectedLen {
		padding := make([]byte, expectedLen-len(b))
		b = append(b, padding...)
	}
	return append(prevResult, b...), nil
}

var _ consensustypes.Encoder = (*PassthroughEncoder)(nil)
