package report

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	consensustypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
)

var (
	reportA = []byte{0x01, 0x02, 0x03}

	workflowID       = "15c631d295ef5e32deb99a10ee6804bc4af1385568f9b3363f6552ac6dbb2cef"
	workflowName     = "aabbccddeeaabbccddee"
	donID            = uint32(2)
	donIDHex         = "00000002"
	executionID      = "8d4e66421db647dd916d3ec28d56188c8d7dae5f808e03d03339ed2562f13bb0"
	workflowOwnerID  = "0000000000000000000000000000000000000000"
	reportID         = "9988"
	timestampInt     = uint32(1234567890)
	timestampHex     = "499602d2"
	configVersionInt = uint32(1)
	configVersionHex = "00000001"
)

func TestPrependMetadataFields(t *testing.T) {
	result, err := PrependMetadataFields(getMetadata(workflowID), reportA)
	require.NoError(t, err)

	resultAsHex := hex.EncodeToString(result)

	reportAAsString := hex.EncodeToString(reportA)
	expectedResult := getHexMetadata() + reportAAsString

	assert.Equal(t, expectedResult, resultAsHex)
}

func getHexMetadata() string {
	return "01" + executionID + timestampHex + donIDHex + configVersionHex + workflowID + workflowName + workflowOwnerID + reportID
}

func getMetadata(cid string) consensustypes.Metadata {
	return consensustypes.Metadata{
		Version:          1,
		ExecutionID:      executionID,
		Timestamp:        timestampInt,
		DONID:            donID,
		DONConfigVersion: configVersionInt,
		WorkflowID:       cid,
		WorkflowName:     workflowName,
		WorkflowOwner:    workflowOwnerID,
		ReportID:         reportID,
	}
}
