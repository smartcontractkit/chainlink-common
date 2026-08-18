package ocrcommon_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/ocrcommon"
)

// accountlessTransmitter is a plugin's own transmitter: it knows where a report
// goes and not who it is transmitting as.
type accountlessTransmitter struct {
	transmitted int
}

var _ ocr3types.ContractTransmitter[[]byte] = (*accountlessTransmitter)(nil)

func (t *accountlessTransmitter) Transmit(
	context.Context, ocrtypes.ConfigDigest, uint64, ocr3types.ReportWithInfo[[]byte], []ocrtypes.AttributedOnchainSignature,
) error {
	t.transmitted++
	return nil
}

func (t *accountlessTransmitter) FromAccount(context.Context) (ocrtypes.Account, error) {
	return "", nil
}

func TestTransmitterWithAccount(t *testing.T) {
	t.Parallel()

	const account = ocrtypes.Account("5994a5155e9b81ab7794b79bfbf076ef5ef7c437")

	inner := &accountlessTransmitter{}
	wrapped := ocrcommon.TransmitterWithAccount(account, inner)

	got, err := wrapped.FromAccount(t.Context())
	require.NoError(t, err)
	assert.Equal(t, account, got)

	// Only FromAccount is answered here; transmitting is still the plugin's.
	require.NoError(t, wrapped.Transmit(t.Context(), ocrtypes.ConfigDigest{}, 1, ocr3types.ReportWithInfo[[]byte]{}, nil))
	assert.Equal(t, 1, inner.transmitted)
}
