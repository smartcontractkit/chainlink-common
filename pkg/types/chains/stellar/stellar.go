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
type Client interface {
	// GetLedgerEntries fetches ledger entries by XDR key (used for sequence number lookups).
	GetLedgerEntries(ctx context.Context, req GetLedgerEntriesRequest) (GetLedgerEntriesResponse, error)
	// GetLatestLedger returns current ledger info (used for timeout detection).
	GetLatestLedger(ctx context.Context) (GetLatestLedgerResponse, error)
	// GetEvents fetches contract events matching the provided ledger range, filters, and pagination.
	GetEvents(ctx context.Context, req GetEventsRequest) (GetEventsResponse, error)
	// GetTransaction fetches an on-chain transaction by hash.
	GetTransaction(ctx context.Context, req GetTransactionRequest) (GetTransactionResponse, error)
	// SimulateTransaction builds a synthetic single-operation Soroban InvokeContract
	// transaction and simulates it without submitting it.
	SimulateTransaction(ctx context.Context, req SimulateTransactionRequest) (SimulateTransactionResponse, error)
}

// GetSigningAccountResponse is the relayer's default TXM signing account (G... StrKey).
//
// Some Soroban contracts (e.g. CRE forwarder report()) take the transmitter as an
// explicit Address argument checked via require_auth(). That contract argument is
// separate from the transaction source account (FromAddress) used for signing.
// Callers that must encode such arguments should query this address from the relayer
// keystore rather than hard-coding capability config.
type GetSigningAccountResponse struct {
	AccountAddress string
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

// SimulateAuthMode controls how Soroban authorization is handled during simulation.
type SimulateAuthMode string

const (
	// SimulateAuthModeRecord records required authorization entries.
	// This is the recommended default when AuthMode is empty.
	SimulateAuthModeRecord SimulateAuthMode = "record"

	// SimulateAuthModeEnforce enforces authorization entries already present on the invocation.
	SimulateAuthModeEnforce SimulateAuthMode = "enforce"

	// SimulateAuthModeRecordAllowNonroot records non-root authorization entries where supported.
	SimulateAuthModeRecordAllowNonroot SimulateAuthMode = "record_allow_nonroot"
)

// SimulateResourceConfig carries optional resource configuration for simulation.
type SimulateResourceConfig struct {
	// InstructionLeeway is the extra instruction budget leeway requested for simulation.
	InstructionLeeway uint64
}

// SimulateTransactionRequest is the domain representation of a Soroban contract-call simulation.
//
// It builds a synthetic single-operation InvokeContract transaction from ContractID,
// Function, and Args, then simulates that transaction without submitting it.
type SimulateTransactionRequest struct {
	// ContractID is the Stellar contract address in C… StrKey encoding.
	ContractID string

	// Function is the Soroban function name to call.
	Function string

	// Args holds one ScVal per contract argument.
	// An empty slice is valid for zero-argument functions.
	Args []ScVal

	// SourceAccount is the G… account used as the synthetic transaction and operation source.
	//
	// This is not necessarily the same as any Address argument that the contract
	// authorizes via require_auth. Leave empty to use the service default source.
	SourceAccount string

	// AuthMode controls authorization behavior during simulation.
	// Empty means the implementation default, which should be record.
	AuthMode SimulateAuthMode

	// ResourceConfig optionally customizes simulation resource behavior.
	ResourceConfig *SimulateResourceConfig
}

// SimulateRestorePreamble carries restore transaction data returned by simulation
// when archived ledger entries must be restored before the invocation can be submitted.
type SimulateRestorePreamble struct {
	// TransactionDataXDR is the base64-encoded SorobanTransactionData for restore.
	TransactionDataXDR string

	// MinResourceFee is the minimum resource fee for the restore preamble.
	MinResourceFee int64
}

// SimulateTransactionResponse is the domain representation of a Soroban simulation result.
type SimulateTransactionResponse struct {
	// LedgerSequence is the ledger that was used for the simulation.
	LedgerSequence uint32

	// Success is true when transport succeeded and the simulation itself did not
	// return a host or contract error.
	Success bool

	// Error is non-empty when simulation failed at the host or contract layer.
	Error string

	// ReturnValueXDR is the base64-encoded ScVal return value, when present.
	// Empty is valid for void/unit-returning contract calls.
	ReturnValueXDR string

	// RequiredAuthXDR contains base64-encoded SorobanAuthorizationEntry values
	// returned by simulation, typically when AuthMode is record.
	RequiredAuthXDR []string

	// EventsXDR contains base64-encoded diagnostic/event XDR values returned by simulation.
	EventsXDR []string

	// TransactionDataXDR is the base64-encoded SorobanTransactionData returned by simulation.
	TransactionDataXDR string

	// MinResourceFee is the minimum resource fee returned by simulation.
	MinResourceFee int64

	// RestorePreamble is set when archived ledger entries must be restored before submission.
	RestorePreamble *SimulateRestorePreamble
}

// SubmitTransactionRequest invokes a Soroban contract via the chain's TXM pipeline.
// The TXM handles simulation, sequence management, signing, fee bumping, and on-chain confirmation;
// callers only need to supply the logical contract invocation parameters.
type SubmitTransactionRequest struct {
	// IdempotencyKey optionally identifies the transaction for TXM deduplication.
	// Leave empty to let the relayer TXM assign one; the assigned key is returned
	// as TxIdempotencyKey in SubmitTransactionResponse.
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
	// BlockTimestamp is the ledger close time in microseconds, when available.
	BlockTimestamp *uint64
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

type EventType int32

const (
	EventTypeSystem EventType = iota
	EventTypeContract
)

type TopicSegment struct {
	Wildcard *string
	Value    *ScVal
}

type TopicFilter struct {
	Segments []TopicSegment
}

type EventFilter struct {
	EventTypes  []EventType
	ContractIDs []string
	Topics      []TopicFilter
}

type PaginationOptions struct {
	Cursor string
	Limit  uint32
}

type GetEventsRequest struct {
	StartLedger uint32
	EndLedger   uint32

	Filters    []EventFilter
	Pagination *PaginationOptions
}

type EventInfo struct {
	EventType EventType

	Ledger         uint32
	LedgerClosedAt string

	ContractID string
	ID         string

	OperationIndex   uint32
	TransactionIndex uint32
	TransactionHash  string

	Topics []ScVal
	Value  ScVal
}

type GetEventsResponse struct {
	Events []EventInfo

	Cursor string

	LatestLedger          uint32
	OldestLedger          uint32
	LatestLedgerCloseTime int64
	OldestLedgerCloseTime int64
}

// GetTransactionRequest fetches a transaction by hash.
type GetTransactionRequest struct {
	TxHash string
}

// GetTransactionResponse carries fee and ledger metadata for a confirmed transaction.
type GetTransactionResponse struct {
	FeeStroops      uint64
	LedgerSequence  uint32
	LedgerCloseTime int64 // unix seconds
}
