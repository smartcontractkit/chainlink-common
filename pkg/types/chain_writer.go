package types

import (
	"context"
	"math/big"

	"github.com/google/uuid"
)

// CallOption can be used to configure a single call to the ChainWriter.
type CallOption interface {
	apply(*CallOptions)
}

// CallOptionFunc is a functional type that implements the CallOption interface.
type callOptionFunc func(*CallOptions)

func (fn callOptionFunc) apply(o *CallOptions) {
	fn(o)
}

// CallOptions contains the values which are configurable per-call to the ChainWriter.
//
// ChainWriter implementations should hold an instance of this struct and apply the options to the struct before making the call.
type CallOptions struct {
	// TransactionID is submitted to the particular chain's TXM as an idempotency key.
	// It should be unique across transaction.
	TransactionID   string
	TransactionMeta *TxMeta
	Value           *big.Int
}

// Apply applies the given options to the CallOptions.
func (o *CallOptions) Apply(opts ...CallOption) {
	for _, opt := range opts {
		opt.apply(o)
	}
}

// WithTransactionID sets the transactionID (idempotency key) for a call to the ChainWriter.
//
// Default: ""
func WithTransactionID(transactionID string) CallOption {
	return callOptionFunc(func(o *CallOptions) {
		o.TransactionID = transactionID
	})
}

// WithTransactionMeta sets the transaction metadata for a call to the ChainWriter.
//
// Default: <nil>
func WithTransactionMeta(meta *TxMeta) CallOption {
	return callOptionFunc(func(o *CallOptions) {
		o.TransactionMeta = meta
	})
}

// WithValue sets the value to be sent with a transaction.
//
// Default: <nil>
func WithValue(value *big.Int) CallOption {
	return callOptionFunc(func(o *CallOptions) {
		o.Value = value
	})
}

// MethodArguments can be any object which maps a set of generic parameters into chain specific parameters defined in RelayConfig.
//
// It must encode as an object via [json.Marshal] and [github.com/fxamacker/cbor/v2.Marshal].
// Typically, would be either a struct with field names mapping to arguments, or anonymous map such as `map[string]any{"baz": 42, "test": true}}`
//
// Example use:
//
//	 type ReportParams struct {
//	 		Receiver	string
//			RawReport 	[]byte
//			Signatures	[][]byte
//	 }
//
//	 func submitReport(ctx context.Context, cw ChainWriter) error {
//			params := ReportParams{ ... }
//			return cw.SubmitTransaction(ctx, "ReportContract", "Report", &params, "0x1234")
//	 }
//
// Note that implementations should ignore extra fields in params that are not expected in the call to allow easier
// use across chains and contract versions.
// Similarly, when using a struct for returnVal, fields in the return value that are not on-chain will not be set.
type MethodArguments any

type ChainWriter interface {
	// SubmitTransaction packs and broadcasts a transaction to the underlying chain.
	//
	// - `contract`, and `method` are the selectors for the contract and method to be called.
	// - `args` should be any object which maps a set of method param into the contract and method specific method params.
	// - `opts` is an optional list of CallOptions that can be used to configure a transaction. These configurations are not
	// persisted across calls.
	SubmitTransaction(ctx context.Context, contract, method string, args MethodArguments, toAddress string, opts ...CallOption) error

	// GetTransactionStatus returns the current status of a transaction in the underlying chain's TXM.
	GetTransactionStatus(ctx context.Context, transactionID uuid.UUID) (TransactionStatus, error)

	// GetFeeComponents retrieves the associated gas costs for executing a transaction.
	GetFeeComponents(ctx context.Context) (*ChainFeeComponents, error)
}

// TxMeta contains metadata fields for a transaction.
type TxMeta struct {
	// Used for Keystone Workflows
	WorkflowExecutionID *string
}

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
	ExecutionFee big.Int

	// The cost associated with an L2 posting a transaction's data to the L1.
	DataAvailabilityFee big.Int
}
