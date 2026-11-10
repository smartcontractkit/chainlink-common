package ocrutil

import (
	"context"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// ForRegistryConfig adjusts lc for config read from the capabilities registry: there are no block confirmations to
// wait for, and config changes should be picked up quickly. It also works with config contracts, which then just
// don't wait for confirmations.
func ForRegistryConfig(lc ocrtypes.LocalConfig) ocrtypes.LocalConfig {
	lc.SkipContractConfigConfirmations = true
	lc.ContractConfigTrackerPollInterval = 5 * time.Second
	return lc
}

// TransmitterWithAccount wraps transmitter to report account as the account it transmits from, which is how the oracle
// finds itself among the transmitters in its config.
func TransmitterWithAccount[RI any](account ocrtypes.Account, transmitter ocr3types.ContractTransmitter[RI]) ocr3types.ContractTransmitter[RI] {
	return &accountTransmitter[RI]{ContractTransmitter: transmitter, account: account}
}

type accountTransmitter[RI any] struct {
	ocr3types.ContractTransmitter[RI]
	account ocrtypes.Account
}

func (t *accountTransmitter[RI]) FromAccount(context.Context) (ocrtypes.Account, error) {
	return t.account, nil
}
