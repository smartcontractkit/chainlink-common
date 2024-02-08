package ccip

import (
	"context"
)

type USDCReader interface {
	// GetLastUSDCMessagePriorToLogIndexInTx returns the last USDC message that was sent
	// before the provided log index in the given transaction.
	GetLastUSDCMessagePriorToLogIndexInTx(ctx context.Context, logIndex int64, txHash Hash) ([]byte, error)
}
