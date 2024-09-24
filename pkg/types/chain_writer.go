package types

import (
	"context"
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

const (
	ErrSettingTransactionGasLimitNotSupported = InvalidArgumentError("setting transaction gas limit is not supported")
)

type ChainWriter interface {
	services.Service

	// SubmitTransaction packs and broadcasts a transaction to the underlying chain.
	//
	// - `args` should be any object which maps a set of method param into the contract and method specific method params.
	// - `transactionID` will be used by the underlying TXM as an idempotency key, and unique reference to track transaction attempts.
	SubmitTransaction(ctx context.Context, contractName, method string, args any, transactionID string, toAddress string, meta *TxMeta, value *big.Int) error

	// GetTransactionStatus returns the current status of a transaction in the underlying chain's TXM.
	GetTransactionStatus(ctx context.Context, transactionID string) (TransactionStatus, error)

	// GetFeeComponents retrieves the associated gas costs for executing a transaction.
	GetFeeComponents(ctx context.Context) (*ChainFeeComponents, error)
}

// TxMeta contains metadata fields for a transaction.
type TxMeta struct {
	// Used for Keystone Workflows
	WorkflowExecutionID *string
	// An optional maximum gas limit for the transaction. If not set the ChainWriter implementation will be responsible for
	// setting a gas limit for the transaction.  If it is set and the ChainWriter implementation does not support setting
	// this value per transaction it will return ErrSettingTransactionGasLimitNotSupported
	GasLimit *big.Int
}

// TransactionStatus are the status we expect every TXM to support and that can be returned by StatusForUUID.
type TransactionStatus int

const (
	Unknown TransactionStatus = iota
	Pending
	Unconfirmed
	Finalized
	Failed
	Fatal
)

// ChainFeeComponents contains the different cost components of executing a transaction.
type ChainFeeComponents struct {
	// The cost of executing transaction in the chain's EVM (or the L2 environment).
	ExecutionFee *big.Int

	// The cost associated with an L2 posting a transaction's data to the L1.
	DataAvailabilityFee *big.Int
}
