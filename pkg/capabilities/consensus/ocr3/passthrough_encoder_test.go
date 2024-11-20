package ocr3_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	consensustypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

func Test_PassthroughEncoder_Encode(t *testing.T) {
	t.Parallel()
	input := map[string]any{
		"INTERNAL_METADATA": consensustypes.Metadata{
			Version:          1,
			ExecutionID:      "9916550055460393550154715607595304997184366551966969077087191279",
			Timestamp:        1000000000,
			DONID:            1,
			DONConfigVersion: 1,
			WorkflowID:       "7558258820343239239009088930242061323966339210999574814231444361",
			WorkflowName:     "11111111111111",
			WorkflowOwner:    "5c7cb2d007218404a2f38ade9738735faf56a8c6",
			ReportID:         "0001",
		},
		"0": []byte{0x01, 0x02, 0x03},
	}
	inputWrapped, err := values.NewMap(input)
	require.NoError(t, err)

	encoder := ocr3.PassthroughEncoder{}
	actual, err := encoder.Encode(tests.Context(t), *inputWrapped)
	require.NoError(t, err)

	assert.True(t, bytes.Equal([]byte{0x01, 0x02, 0x03}, actual[len(actual)-3:]))

	metadata, err := decodeReportMetadata(t, actual)
	require.NoError(t, err)

	assert.Equal(t, uint8(1), metadata.Version)
	assert.Equal(t, "9916550055460393550154715607595304997184366551966969077087191279", hex.EncodeToString(metadata.WorkflowExecutionID[:]))
	assert.Equal(t, uint32(1000000000), metadata.Timestamp)
	assert.Equal(t, uint32(1), metadata.DonID)
	assert.Equal(t, uint32(1), metadata.DonConfigVersion)
	assert.Equal(t, "7558258820343239239009088930242061323966339210999574814231444361", hex.EncodeToString(metadata.WorkflowCID[:]))
	assert.Equal(t, "11111111111111000000", hex.EncodeToString(metadata.WorkflowName[:])) // padded by 3 bytes
	assert.Equal(t, "5c7cb2d007218404a2f38ade9738735faf56a8c6", hex.EncodeToString(metadata.WorkflowOwner[:]))
	assert.Equal(t, []byte{0x00, 0x01}, metadata.ReportID[:])
}

func decodeReportMetadata(t *testing.T, data []byte) (metadata ocr3.ReportV1Metadata, err error) {
	require.True(t, len(data) >= metadata.Length())
	return metadata, binary.Read(bytes.NewReader(data[:metadata.Length()]), binary.BigEndian, &metadata)
}
