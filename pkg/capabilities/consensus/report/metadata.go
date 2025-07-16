package report

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	consensustypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
)

func PrependMetadataFields(meta consensustypes.Metadata, userPayload []byte) ([]byte, error) {
	var err error
	var result []byte

	// 1. Version (1 byte)
	if meta.Version > 255 {
		return nil, errors.New("version must be between 0 and 255")
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

	// 7. Workflow Name (10 bytes)
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
	if len(b) != expectedLen {
		return nil, fmt.Errorf("incorrect length for id %s (%s), expected %d bytes, got %d", logName, id, expectedLen, len(b))
	}
	return append(prevResult, b...), nil
}
