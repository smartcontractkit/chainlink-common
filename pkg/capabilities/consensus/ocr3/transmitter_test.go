package ocr3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestContractTransmitter_IsDeprecated(t *testing.T) {
	transmitter := NewContractTransmitter(logger.Test(t), nil, "0xacct")

	err := transmitter.Transmit(t.Context(), types.ConfigDigest{}, 0, ocr3types.ReportWithInfo[[]byte]{}, nil)
	require.ErrorIs(t, err, ErrDeprecated)

	account, err := transmitter.FromAccount(t.Context())
	require.ErrorIs(t, err, ErrDeprecated)
	assert.Empty(t, account)
}
