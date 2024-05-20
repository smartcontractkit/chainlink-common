package types

import (
	"context"
	"math/big"

	"github.com/google/uuid"
)

type ChainWriter interface {
	// SubmiSignedTransaction packs and broadcasts a transaction to the underlying chain.
	//
	// The `transactionID` will be used by the underlying TXM as an idempotency key, and unique reference to track transaction attempts.
	SubmitSignedTransaction(ctx context.Context, payload []byte, signature map[string]any, transactionID uuid.UUID, toAddress string, meta *TxMeta, value big.Int) (int64, error)

	// StatusForUUID returns the current status of a transaction in the underlying chain's TXM.
	StatusForUUID(ctx context.Context, transactionID uuid.UUID) (TransactionStatus, error)

	// GetFeeComponents retrieves the associated gas costs for executing a transaction.
	GetFeeComponents() (ChainFeeComponents, error)
}

// TxMeta contains metadata fields for a transaction.
//
// Eventually this will replace, or be replaced by (via a move), the `TxMeta` in core:
// https://github.com/smartcontractkit/chainlink/blob/dfc399da715f16af1fcf6441ea5fc47b71800fa1/common/txmgr/types/tx.go#L121
type TxMeta map[string]string

// TransactionStatus are the status we expect every TXM to support and that can be returned by StatusForUUID.
type TransactionStatus int

const (
	Unknown TransactionStatus = iota
	Unconfirmed
	Finalized
	Failed
	Fatal
)

// ChainFeeComponents contains the different cost components of executing a transaction.
type ChainFeeComponents struct {
	// The cost of executing transaction in the chain's EVM (or the L2 environment).
	ExecutionPrice big.Int

	// The cost associated with an L2 posting a transaction's data to the L1.
	DataAvailabilityPrice big.Int
}
