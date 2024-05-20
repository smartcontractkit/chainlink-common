package types

import (
	"context"
	"math/big"

	"github.com/google/uuid"
)

type ChainWriter interface {
	// SubmitTransaction packs and broadcasts a transaction to the underlying chain.
	//
	// The `payload` array must contain the encoded transaction payload, and the related signatures.
	// The `transactionID` will be used by the underlying TXM as an idempotency key, and unique reference to track transaction attempts.
	SubmitTransaction(ctx context.Context, payload []any, transactionID uuid.UUID, toAddress string, meta *TxMeta, value big.Int) (int64, error)

	// StatusForUUID returns the current status of a transaction in the underlying chain's TXM.
	StatusForUUID(ctx context.Context, transactionID uuid.UUID) (TransactionStatus, error)

	// GetFeeComponents retrieves the associated gas costs for executing a transaction.
	GetFeeComponents() (ChainFeeComponents, error)
}

// TODO(nickcorin): Intentionally left blank for now.
type TxMeta struct{}

// TransactionStatus are the status we expect every TXM to support and that can be returned by StatusForUUID.
type TransactionStatus int

const (
	Unconfirmed TransactionStatus = iota
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

	AssetType ChainFeeType
}

// ChainFeeType describes the asset type the underlying chain uses to estimate transaction costs.
type ChainFeeType int

const (
	Wei ChainFeeType = iota
	Sun
	Octa
)
