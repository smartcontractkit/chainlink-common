package solana

import (
	"context"
	"math/big"
)

const (
	PublicKeyLength = 32
	HashLength      = PublicKeyLength
	SignatureLength = 64
)

// represents solana-go PublicKey
type PublicKey [PublicKeyLength]byte

// represents solana-go PublicKeySlice
type PublicKeySlice []PublicKey

// represents solana-go Signature
type Signature [SignatureLength]byte

// represents solana-go Hash
type Hash PublicKey

// represents solana-go AccountsMeta
type AccountMeta struct {
	PublicKey  PublicKey
	IsWritable bool
	IsSigner   bool
}

// represents solana-go AccountMetaSlice
type AccountMetaSlice []*AccountMeta

// represents solana-go EncodingType
type EncodingType string

const (
	EncodingBase58     EncodingType = "base58"      // limited to Account data of less than 129 bytes
	EncodingBase64     EncodingType = "base64"      // will return base64 encoded data for Account data of any size
	EncodingBase64Zstd EncodingType = "base64+zstd" // compresses the Account data using Zstandard and base64-encodes the result

	// attempts to use program-specific state parsers to
	// return more human-readable and explicit account state data.
	// If "jsonParsed" is requested but a parser cannot be found,
	// the field falls back to "base64" encoding, detectable when the data field is type <string>.
	// Cannot be used if specifying dataSlice parameters (offset, length).
	EncodingJSONParsed EncodingType = "jsonParsed"

	EncodingJSON EncodingType = "json" // NOTE: you're probably looking for EncodingJSONParsed
)

// represents solana-go CommitmentType
type CommitmentType string

const (
	// The node will query the most recent block confirmed by supermajority
	// of the cluster as having reached maximum lockout,
	// meaning the cluster has recognized this block as finalized.
	CommitmentFinalized CommitmentType = "finalized"

	// The node will query the most recent block that has been voted on by supermajority of the cluster.
	// - It incorporates votes from gossip and replay.
	// - It does not count votes on descendants of a block, only direct votes on that block.
	// - This confirmation level also upholds "optimistic confirmation" guarantees in release 1.3 and onwards.
	CommitmentConfirmed CommitmentType = "confirmed"

	// The node will query its most recent block. Note that the block may still be skipped by the cluster.
	CommitmentProcessed CommitmentType = "processed"
)

// Client wraps Solana RPC via type/chains/solana types.
type Client interface {
	// GetBalance: lamports for account at commitment.
	// In: ctx, {Addr, Commitment}. Out: {Value}, error.
	GetBalance(ctx context.Context, req GetBalanceRequest) (*GetBalanceReply, error)

	// GetAccountInfoWithOpts: account state + metadata.
	// In: ctx, {Account, Opts(Encoding, Commitment, DataSlice, MinContextSlot)}.
	// Out: {Context.Slot, Value(*Account|nil)}, error.
	GetAccountInfoWithOpts(ctx context.Context, req GetAccountInfoRequest) (*GetAccountInfoReply, error)

	// GetMultipleAccountsWithOpts: batch account fetch.
	// In: ctx, {Accounts, Opts}. Out: {Context.Slot, Value([]*Account)}, error.
	GetMultipleAccountsWithOpts(ctx context.Context, req GetMultipleAccountsRequest) (*GetMultipleAccountsReply, error)

	// GetBlock: block by slot with optional tx details.
	// In: ctx, {Slot, Opts(Encoding, TxDetails, Rewards, Commitment, MaxTxVersion)}.
	// Out: block fields + Transactions/Signatures, error.
	GetBlock(ctx context.Context, req GetBlockRequest) (*GetBlockReply, error)

	// GetSlotHeight: current height at commitment.
	// In: ctx, {Commitment}. Out: {Height}, error.
	GetSlotHeight(ctx context.Context, req GetSlotHeightRequest) (*GetSlotHeightReply, error)

	// GetTransaction: tx by signature with meta.
	// In: ctx, {Signature}. Out: {Slot, BlockTime, Version, Transaction, Meta}, error.
	GetTransaction(ctx context.Context, req GetTransactionRequest) (*GetTransactionReply, error)

	// GetFeeForMessage: fee estimate for base58 message.
	// In: ctx, {Message, Commitment}. Out: {Fee}, error.
	GetFeeForMessage(ctx context.Context, req GetFeeForMessageRequest) (*GetFeeForMessageReply, error)

	// GetSignatureStatuses: confirmation status for sigs.
	// In: ctx, {Sigs}. Out: {Results[Slot, Confirmations, Err, Status]}, error.
	GetSignatureStatuses(ctx context.Context, req GetSignatureStatusesRequest) (*GetSignatureStatusesReply, error)

	// SimulateTX: dry-run a transaction.
	// In: ctx, {Receiver, EncodedTransaction, Opts(SigVerify, Commitment, ReplaceRecentBlockhash, Accounts)}.
	// Out: {Err, Logs, Accounts, UnitsConsumed}, error.
	SimulateTX(ctx context.Context, req SimulateTXRequest) (*SimulateTXReply, error)
}

// represents solana-go DataSlice
type DataSlice struct {
	Offset *uint64
	Length *uint64
}

// represents solana-go GetAccountInfoOpts
type GetAccountInfoOpts struct {
	// Encoding for Account data.
	// Either "base58" (slow), "base64", "base64+zstd", or "jsonParsed".
	// - "base58" is limited to Account data of less than 129 bytes.
	// - "base64" will return base64 encoded data for Account data of any size.
	// - "base64+zstd" compresses the Account data using Zstandard and base64-encodes the result.
	// - "jsonParsed" encoding attempts to use program-specific state parsers to return more
	// 	 human-readable and explicit account state data. If "jsonParsed" is requested but a parser
	//   cannot be found, the field falls back to "base64" encoding,
	//   detectable when the data field is type <string>.
	//
	// This parameter is optional.
	Encoding EncodingType

	// Commitment requirement.
	//
	// This parameter is optional. Default value is Finalized
	Commitment CommitmentType

	// dataSlice parameters for limiting returned account data:
	// Limits the returned account data using the provided offset and length fields;
	// only available for "base58", "base64" or "base64+zstd" encodings.
	//
	// This parameter is optional.
	DataSlice *DataSlice

	// The minimum slot that the request can be evaluated at.
	// This parameter is optional.
	MinContextSlot *uint64
}

// represents solana-go RPCContext
type RPCContext struct {
	Slot uint64
}

// represents solana-go Account
type Account struct {
	// Number of lamports assigned to this account
	Lamports uint64

	// Pubkey of the program this account has been assigned to
	Owner PublicKey

	// Data associated with the account, either as encoded binary data or JSON format {<program>: <state>}, depending on encoding parameter
	Data *DataBytesOrJSON

	// Boolean indicating if the account contains a program (and is strictly read-only)
	Executable bool

	// The epoch at which this account will next owe rent
	RentEpoch *big.Int

	// The amount of storage space required to store the token account
	Space uint64
}

// represents solana-go TransactionDetailsType
type TransactionDetailsType string

const (
	TransactionDetailsFull       TransactionDetailsType = "full"
	TransactionDetailsSignatures TransactionDetailsType = "signatures"
	TransactionDetailsNone       TransactionDetailsType = "none"
	TransactionDetailsAccounts   TransactionDetailsType = "accounts"
)

type TransactionVersion int

const (
	LegacyTransactionVersion TransactionVersion = -1
	legacyVersion                               = `"legacy"`
)

type ConfirmationStatusType string

const (
	ConfirmationStatusProcessed ConfirmationStatusType = "processed"
	ConfirmationStatusConfirmed ConfirmationStatusType = "confirmed"
	ConfirmationStatusFinalized ConfirmationStatusType = "finalized"
)

type UiTokenAmount struct {
	// Raw amount of tokens as a string, ignoring decimals.
	Amount string

	// Number of decimals configured for token's mint.
	Decimals uint8

	// Token amount as a string, accounting for decimals.
	UiAmountString string
}

type TokenBalance struct {
	// Index of the account in which the token balance is provided for.
	AccountIndex uint16

	// Pubkey of token balance's owner.
	Owner *PublicKey
	// Pubkey of token program.
	ProgramId *PublicKey

	// Pubkey of the token's mint.
	Mint          PublicKey
	UiTokenAmount *UiTokenAmount
}

type LoadedAddresses struct {
	ReadOnly PublicKeySlice
	Writable PublicKeySlice
}

type Data struct {
	Content  []byte
	Encoding EncodingType
}

type ReturnData struct {
	ProgramId PublicKey
	Data      Data
}

type TransactionMeta struct {
	// Error if transaction failed, empty if transaction succeeded.
	Err string
	// Fee this transaction was charged
	Fee uint64

	// Array of u64 account balances from before the transaction was processed
	PreBalances []uint64

	// Array of u64 account balances after the transaction was processed
	PostBalances []uint64

	// List of inner instructions or omitted if inner instruction recording
	// was not yet enabled during this transaction
	InnerInstructions []InnerInstruction

	// List of token balances from before the transaction was processed
	// or omitted if token balance recording was not yet enabled during this transaction
	PreTokenBalances []TokenBalance

	// List of token balances from after the transaction was processed
	// or omitted if token balance recording was not yet enabled during this transaction
	PostTokenBalances []TokenBalance

	// Array of string log messages or omitted if log message
	// recording was not yet enabled during this transaction
	LogMessages []string

	LoadedAddresses LoadedAddresses

	// Data returned by transaction (if any
	ReturnData ReturnData

	// Total compute units consumed by transaction execution
	ComputeUnitsConsumed *uint64
}

type InnerInstruction struct {
	Index uint16

	// Ordered list of inner program instructions that were invoked during a single transaction instruction.
	Instructions []CompiledInstruction
}

type CompiledInstruction struct {
	// Index into the message.accountKeys array indicating the program account that executes this instruction.
	ProgramIDIndex uint16

	// List of ordered indices into the message.accountKeys array indicating which accounts to pass to the program.
	Accounts []uint16

	// The program input data encoded in a base-58 string.
	Data []byte

	StackHeight uint16
}

// represents solana-go GetBlockOpts
type GetBlockOpts struct {
	// "processed" is not supported.
	// If parameter not provided, the default is "finalized".
	//
	// This parameter is optional.
	Commitment CommitmentType
}

var (
	MaxSupportedTransactionVersion0 uint64 = 0
	MaxSupportedTransactionVersion1 uint64 = 1
)

// UnixTimeSeconds represents a UNIX second-resolution timestamp.
type UnixTimeSeconds int64

type GetMultipleAccountsOpts GetAccountInfoOpts

type SimulateTransactionAccountsOpts struct {
	// (optional) Encoding for returned Account data,
	// either "base64" (default), "base64+zstd" or "jsonParsed".
	// - "jsonParsed" encoding attempts to use program-specific state parsers
	//   to return more human-readable and explicit account state data.
	//   If "jsonParsed" is requested but a parser cannot be found,
	//   the field falls back to binary encoding, detectable when
	//   the data field is type <string>.
	Encoding EncodingType

	// An array of accounts to return.
	Addresses []PublicKey
}

type SimulateTXOpts struct {
	// If true the transaction signatures will be verified
	// (default: false, conflicts with ReplaceRecentBlockhash)
	SigVerify bool

	// Commitment level to simulate the transaction at.
	// (default: "finalized").
	Commitment CommitmentType

	// If true the transaction recent blockhash will be replaced with the most recent blockhash.
	// (default: false, conflicts with SigVerify)
	ReplaceRecentBlockhash bool

	Accounts *SimulateTransactionAccountsOpts
}

type GetAccountInfoRequest struct {
	Account PublicKey
	Opts    *GetAccountInfoOpts
}

type GetAccountInfoReply struct {
	RPCContext
	Value *Account
}

type GetMultipleAccountsRequest struct {
	Accounts []PublicKey
	Opts     *GetMultipleAccountsOpts
}

type GetMultipleAccountsReply struct {
	RPCContext
	Value []*Account
}

type GetBalanceRequest struct {
	Addr       PublicKey
	Commitment CommitmentType
}

type GetBalanceReply struct {
	// Balance in lamports
	Value uint64
}

type GetBlockRequest struct {
	Slot uint64
	Opts *GetBlockOpts
}

// Will contain a AsJSON if the requested encoding is `solana.EncodingJSON`
// (which is also the default when the encoding is not specified),
// or a `AsDecodedBinary` in case of EncodingBase58, EncodingBase64.
type DataBytesOrJSON struct {
	RawDataEncoding EncodingType
	AsDecodedBinary []byte
	AsJSON          []byte
}

type GetBlockReply struct {
	// The blockhash of this block.
	Blockhash Hash

	// The blockhash of this block's parent;
	// if the parent block is not available due to ledger cleanup,
	// this field will return "11111111111111111111111111111111".
	PreviousBlockhash Hash

	// The slot index of this block's parent.
	ParentSlot uint64

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch).
	// Nil if not available.
	BlockTime *UnixTimeSeconds

	// The number of blocks beneath this block.
	BlockHeight *uint64
}

type GetSlotHeightRequest struct {
	Commitment CommitmentType
}

type GetSlotHeightReply struct {
	Height uint64
}

type MessageHeader struct {
	// The total number of signatures required to make the transaction valid.
	// The signatures must match the first `numRequiredSignatures` of `message.account_keys`.
	NumRequiredSignatures uint8

	// The last numReadonlySignedAccounts of the signed keys are read-only accounts.
	// Programs may process multiple transactions that load read-only accounts within
	// a single PoH entry, but are not permitted to credit or debit lamports or modify
	// account data.
	// Transactions targeting the same read-write account are evaluated sequentially.
	NumReadonlySignedAccounts uint8

	// The last `numReadonlyUnsignedAccounts` of the unsigned keys are read-only accounts.
	NumReadonlyUnsignedAccounts uint8
}

type MessageAddressTableLookupSlice []MessageAddressTableLookup

type MessageAddressTableLookup struct {
	AccountKey      PublicKey // The account key of the address table.
	WritableIndexes []uint8
	ReadonlyIndexes []uint8
}

type Message struct {
	// List of base-58 encoded public keys used by the transaction,
	// including by the instructions and for signatures.
	// The first `message.header.numRequiredSignatures` public keys must sign the transaction.
	AccountKeys PublicKeySlice

	// Details the account types and signatures required by the transaction.
	Header MessageHeader

	// A base-58 encoded hash of a recent block in the ledger used to
	// prevent transaction duplication and to give transactions lifetimes.
	RecentBlockhash Hash

	// List of program instructions that will be executed in sequence
	// and committed in one atomic transaction if all succeed.
	Instructions []CompiledInstruction

	// List of address table lookups used to load additional accounts for this transaction.
	AddressTableLookups MessageAddressTableLookupSlice
}

type Transaction struct {
	Signatures []Signature
	Message    Message
}

type TransactionResultEnvelope struct {
	AsParsedTransaction *Transaction
	AsDecodedBinary     Data
}

// arguments for solana-rpc GetTransaction call
type GetTransactionRequest struct {
	Signature Signature
}

// result of solana-rpc GetTransaction call
type GetTransactionReply struct {
	Slot uint64

	BlockTime *UnixTimeSeconds

	Version TransactionVersion

	Transaction *TransactionResultEnvelope

	Meta *TransactionMeta
}

type GetFeeForMessageRequest struct {
	// base58 encoded message
	Message    string
	Commitment CommitmentType
}

type GetFeeForMessageReply struct {
	// The amount in lamports the network will charge for a particular message
	Fee uint64
}

type GetLatestBlockhashRequest struct {
	Commitment CommitmentType
}

type GetLatestBlockhashReply struct {
	RPCContext
	Hash                 Hash
	LastValidBlockHeight uint64
}

type SimulateTXRequest struct {
	Receiver PublicKey
	// Encoded
	EncodedTransaction string
	Opts               *SimulateTXOpts
}

type SimulateTXReply struct {
	// Error if transaction failed, empty if transaction succeeded.
	Err string

	// Array of log messages the transaction instructions output during execution,
	// null if simulation failed before the transaction was able to execute
	// (for example due to an invalid blockhash or signature verification failure)
	Logs []string

	// Array of accounts with the same length as the accounts.addresses array in the request.
	Accounts []*Account

	// The number of compute budget units consumed during the processing of this transaction.
	UnitsConsumed *uint64
}

type GetSignatureStatusesRequest struct {
	Sigs []Signature
}

type GetSignatureStatusesResult struct {
	// The slot the transaction was processed.
	Slot uint64

	// Number of blocks since signature confirmation,
	// null if rooted or finalized by a supermajority of the cluster.
	Confirmations *uint64

	// Error if transaction failed, empty if transaction succeeded.
	Err string

	// The transaction's cluster confirmation status; either processed, confirmed, or finalized.
	ConfirmationStatus ConfirmationStatusType
}

type GetSignatureStatusesReply struct {
	Results []GetSignatureStatusesResult
}
