package stellar

import (
	"context"
)

type ScErrorType int32

const (
	ScErrorTypeContract ScErrorType = 0
	ScErrorTypeWasmVM   ScErrorType = 1
	ScErrorTypeContext  ScErrorType = 2
	ScErrorTypeStorage  ScErrorType = 3
	ScErrorTypeObject   ScErrorType = 4
	ScErrorTypeCrypto   ScErrorType = 5
	ScErrorTypeEvents   ScErrorType = 6
	ScErrorTypeBudget   ScErrorType = 7
	ScErrorTypeValue    ScErrorType = 8
	ScErrorTypeAuth     ScErrorType = 9
)

type ScErrorCode int32

const (
	ScErrorCodeArithDomain    ScErrorCode = 0
	ScErrorCodeIndexBounds    ScErrorCode = 1
	ScErrorCodeInvalidInput   ScErrorCode = 2
	ScErrorCodeMissingValue   ScErrorCode = 3
	ScErrorCodeExistingValue  ScErrorCode = 4
	ScErrorCodeExceededLimit  ScErrorCode = 5
	ScErrorCodeInvalidAction  ScErrorCode = 6
	ScErrorCodeInternalError  ScErrorCode = 7
	ScErrorCodeUnexpectedType ScErrorCode = 8
	ScErrorCodeUnexpectedSize ScErrorCode = 9
)

// Client wraps native Stellar RPC calls via the type/chains/stellar domain types.
// Methods map 1:1 to the Stellar RPC API.
type Client interface {
	// GetLedgerEntries fetches ledger entries by XDR key (used for sequence number lookups).
	GetLedgerEntries(ctx context.Context, req GetLedgerEntriesRequest) (GetLedgerEntriesResponse, error)
	// GetLatestLedger returns current ledger info (used for timeout detection).
	GetLatestLedger(ctx context.Context) (GetLatestLedgerResponse, error)
	// ReadContract simulates a read-only Soroban contract function call.
	// Each element of req.Args is a domain ScVal value.
	ReadContract(ctx context.Context, req ReadContractRequest) (ReadContractResponse, error)
}

// GetLedgerEntriesRequest fetches ledger entries by XDR-encoded keys.
type GetLedgerEntriesRequest struct {
	// Keys is a slice of base64-encoded XDR ledger keys.
	Keys []string
}

// LedgerEntryResult is a single ledger entry returned from GetLedgerEntries.
type LedgerEntryResult struct {
	// KeyXDR is the base64-encoded XDR ledger key matching the request.
	KeyXDR string
	// DataXDR is the base64-encoded XDR ledger entry data.
	DataXDR string
	// LastModifiedLedger is the ledger sequence of the last modification.
	LastModifiedLedger uint32
	// LiveUntilLedgerSeq is the ledger until which the entry is live; nil if not applicable.
	LiveUntilLedgerSeq *uint32
	// ExtensionXDR is the base64-encoded XDR ledger entry extension; empty if absent.
	ExtensionXDR string
}

// GetLedgerEntriesResponse contains the requested ledger entries.
type GetLedgerEntriesResponse struct {
	// Entries holds all found ledger entries (may be fewer than keys requested).
	Entries []LedgerEntryResult
	// LatestLedger is the latest ledger sequence number at query time.
	LatestLedger uint32
}

// ReadContractRequest is the domain representation of a Soroban read-only call.
type ReadContractRequest struct {
	// ContractID is the Stellar contract address in C… StrKey encoding.
	ContractID string
	// Function is the Soroban function name to call.
	Function string
	// Args holds one ScVal per contract argument.
	// An empty slice is valid for zero-argument functions.
	Args []ScVal
	// SourceAccount is the G… account to simulate the call as (the invoker).
	// It is required for contracts whose result depends on the caller, e.g. that
	// call require_auth or branch on the invoker. Leave empty for source-insensitive reads.
	SourceAccount string
}

// ReadContractResponse is the domain representation of a Soroban simulation result.
type ReadContractResponse struct {
	// Result is a serialized base64 string - return value of the Host Function call.
	Result string
	// LedgerSequence is the ledger that was used for the simulation.
	LedgerSequence uint32
	// Error is non-empty when the call failed.
	Error string
}

// SubmitTransactionRequest invokes a Soroban contract via the chain's TXM pipeline.
// The TXM handles simulation, sequence management, signing, fee bumping, and on-chain confirmation;
// callers only need to supply the logical contract invocation parameters.
type SubmitTransactionRequest struct {
	// IdempotencyKey optionally identifies the transaction for deduplication and status look-up.
	IdempotencyKey string
	// FromAddress is the source/signer account (G… StrKey).
	// Leave empty to use the TXM's default keystore account.
	FromAddress string
	// ContractID is the Soroban contract address to invoke (C… StrKey).
	ContractID string
	// Function is the Soroban function name to call.
	Function string
	// Args holds the typed Soroban function arguments.
	Args []ScVal
	// LedgerBoundsOffset overrides the TXM's configured ledger bounds for this transaction.
	// Zero means use the TXM default.
	LedgerBoundsOffset uint32
}

// TransactionStatus is the outcome of a submitted transaction.
type TransactionStatus int

const (
	// TxFatal indicates submission failed before reaching the network (RPC, signing, validation).
	TxFatal TransactionStatus = iota
	// TxFailed indicates the transaction was accepted but failed on-chain.
	TxFailed
	// TxSuccess indicates the transaction was accepted and succeeded on-chain.
	TxSuccess
)

// SubmitTransactionResponse carries the result of SubmitTransaction.
type SubmitTransactionResponse struct {
	TxStatus         TransactionStatus
	TxHash           string
	TxIdempotencyKey string
	// ResultXDR is the base64-encoded transaction result XDR when available.
	ResultXDR string
	// ResultMetaXDR is the base64-encoded result meta XDR when available.
	ResultMetaXDR string
	// Error is non-empty when the transaction was accepted but failed on-chain.
	Error string
	// TransactionFee is the total fee charged in stroops (FeeCharged), when available.
	TransactionFee *uint64
}

// GetLatestLedgerResponse holds the current ledger state.
type GetLatestLedgerResponse struct {
	// Hash is the hex-encoded latest ledger hash.
	Hash string
	// ProtocolVersion is the Stellar Core protocol version associated with the ledger.
	ProtocolVersion uint32
	// Sequence is the latest ledger sequence number.
	Sequence uint32
	// LedgerCloseTime is the unix timestamp when the latest ledger closed.
	LedgerCloseTime int64
	// LedgerHeaderXDR is the base64-encoded LedgerHeader XDR for the latest ledger.
	LedgerHeaderXDR string
	// LedgerMetadataXDR is the base64-encoded LedgerCloseMetaV2 XDR for the latest ledger.
	LedgerMetadataXDR string
}

type ScVal struct {
	Type ScValType

	Bool      *bool
	Void      *Void
	Error     *ScError
	U32       *uint32
	I32       *int32
	U64       *uint64
	I64       *int64
	Timepoint *uint64
	Duration  *uint64

	U128 *UInt128Parts
	I128 *Int128Parts
	U256 *UInt256Parts
	I256 *Int256Parts

	Bytes  []byte
	String *string
	Symbol *string

	Vec                       *ScVec
	Map                       *ScMap
	Address                   *ScAddress
	ContractInstance          *ScContractInstance
	LedgerKeyContractInstance *Void
	NonceKey                  *ScNonceKey
}

// ===== Integer Parts =====

type UInt128Parts struct {
	Hi uint64
	Lo uint64
}

type Int128Parts struct {
	Hi int64
	Lo uint64
}

type UInt256Parts struct {
	HiHi uint64
	HiLo uint64
	LoHi uint64
	LoLo uint64
}

type Int256Parts struct {
	HiHi int64
	HiLo uint64
	LoHi uint64
	LoLo uint64
}

// ===== Error =====

type ScError struct {
	Type ScErrorType

	// only one should be set
	ContractCode *uint32
	Code         *ScErrorCode
}

// ===== Accounts / Addresses =====

type MuxedEd25519Account struct {
	ID      uint64
	Ed25519 []byte // 32-byte pubkey
}

type ClaimableBalanceID struct {
	V0 []byte // 32-byte SHA256 hash
}

type ScAddressType int

const (
	ScAddressTypeAccountID ScAddressType = iota
	ScAddressTypeContractID
	ScAddressTypeMuxedAccount
	ScAddressTypeClaimableBalanceID
	ScAddressTypeLiquidityPoolID
)

type ScAddress struct {
	Type ScAddressType

	AccountID        []byte
	ContractID       []byte
	MuxedAccount     *MuxedEd25519Account
	ClaimableBalance *ClaimableBalanceID
	LiquidityPoolID  []byte
}

// ===== Contract Executable =====

type ContractExecutableType int

const (
	ContractExecutableTypeWasmHash ContractExecutableType = iota
	ContractExecutableTypeStellarAsset
)

type ContractExecutable struct {
	Type ContractExecutableType

	WasmHash     []byte
	StellarAsset bool
}

// ===== Contract Instance =====

type ScContractInstance struct {
	Executable *ContractExecutable
	Storage    []ScMapEntry
}

// ===== Misc =====

type ScNonceKey struct {
	Nonce int64
}

type Void struct{}

// ===== Collections =====

type ScMapEntry struct {
	Key *ScVal
	Val *ScVal
}

type ScVec struct {
	Values []*ScVal
}

type ScMap struct {
	Entries []ScMapEntry
}

// ===== SCVal =====

type ScValType int

const (
	ScValTypeBool ScValType = iota
	ScValTypeVoid
	ScValTypeError
	ScValTypeU32
	ScValTypeI32
	ScValTypeU64
	ScValTypeI64
	ScValTypeTimepoint
	ScValTypeDuration
	ScValTypeU128
	ScValTypeI128
	ScValTypeU256
	ScValTypeI256
	ScValTypeBytes
	ScValTypeString
	ScValTypeSymbol
	ScValTypeVec
	ScValTypeMap
	ScValTypeAddress
	ScValTypeContractInstance
	ScValTypeLedgerKeyContractInstance
	ScValTypeNonceKey
)
