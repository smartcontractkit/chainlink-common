package solana

import (
	"context"
	"math/big"
)

const (
	PublicKeyLength = 32
	SignatureLength = 64
)

// represents solana-go PublicKey
type PublicKey [PublicKeyLength]byte

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

type Context struct {
	Slot uint64
}

// represents solana-go RPCContext
type RPCContext struct {
	Context Context
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

type TransactionWithMeta struct {
	// The slot this transaction was processed in.
	Slot uint64

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch)
	// of when the transaction was processed.
	// Nil if not available.
	BlockTime *UnixTimeSeconds

	// Encoded Transaction
	Transaction *DataBytesOrJSON
	// JSON encoded solana-go TransactionMeta
	MetaJSON []byte

	Version TransactionVersion
}

// represents solana-go GetBlockOpts
type GetBlockOpts struct {
	// Encoding for each returned Transaction, either "json", "jsonParsed", "base58" (slow), "base64".
	// If parameter not provided, the default encoding is "json".
	// - "jsonParsed" encoding attempts to use program-specific instruction parsers to return
	//   more human-readable and explicit data in the transaction.message.instructions list.
	// - If "jsonParsed" is requested but a parser cannot be found, the instruction falls back
	//   to regular JSON encoding (accounts, data, and programIdIndex fields).
	//
	// This parameter is optional.
	Encoding EncodingType

	// Level of transaction detail to return.
	// If parameter not provided, the default detail level is "full".
	//
	// This parameter is optional.
	TransactionDetails TransactionDetailsType

	// Whether to populate the rewards array.
	// If parameter not provided, the default includes rewards.
	//
	// This parameter is optional.
	Rewards *bool

	// "processed" is not supported.
	// If parameter not provided, the default is "finalized".
	//
	// This parameter is optional.
	Commitment CommitmentType

	// Max transaction version to return in responses.
	// If the requested block contains a transaction with a higher version, an error will be returned.
	MaxSupportedTransactionVersion *uint64
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

	// Present if "full" transaction details are requested.
	Transactions []TransactionWithMeta

	// Present if "signatures" are requested for transaction details;
	// an array of signatures, corresponding to the transaction order in the block.
	Signatures []Signature

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

type GetTransactionRequest struct {
	Signature Signature
}

type GetTransactionReply struct {
	// Encoded TransactionResultEnvelope
	Transaction []byte

	// JSON encoded TransactionMeta
	Meta []byte
}

type GetFeeForMessageRequest struct {
	// base64 encoded message
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
	Hash                 Signature
	LastValidBlockHeight uint64
}

type SimulateTXRequest struct {
	Receiver           PublicKey
	EncodedTransaction string
	Opts               *SimulateTXOpts
}

type SimulateTXReply struct {
	// Error if transaction failed, null if transaction succeeded.
	Err *string

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
	Slot uint64 `json:"slot"`

	// Number of blocks since signature confirmation,
	// null if rooted or finalized by a supermajority of the cluster.
	Confirmations *uint64 `json:"confirmations"`

	// Error if transaction failed, null if transaction succeeded.
	Err *string

	// The transaction's cluster confirmation status; either processed, confirmed, or finalized.
	ConfirmationStatus ConfirmationStatusType `json:"confirmationStatus"`
}

type GetSignatureStatusesReply struct {
	Results []GetSignatureStatusesResult
}

type Client interface {
	GetBalance(ctx context.Context, req GetBalanceRequest) (*GetBalanceReply, error)
	GetAccountInfoWithOpts(ctx context.Context, req GetAccountInfoRequest) (*GetAccountInfoReply, error)
	GetMultipleAccountsWithOpts(ctx context.Context, req GetMultipleAccountsRequest) (*GetMultipleAccountsReply, error)
	GetBlock(ctx context.Context, req GetBlockRequest) (*GetBlockReply, error)
	GetSlotHeight(ctx context.Context, req GetSlotHeightRequest) (*GetSlotHeightReply, error)
	GetTransaction(ctx context.Context, req GetTransactionRequest) (*GetTransactionReply, error)
	GetFeeForMessage(ctx context.Context, req GetFeeForMessageRequest) (*GetFeeForMessageReply, error)
	GetLatestBlockhash(ctx context.Context, req GetLatestBlockhashRequest) (*GetLatestBlockhashReply, error)
	GetSignatureStatuses(ctx context.Context, req GetSignatureStatusesRequest) (*GetSignatureStatusesReply, error)
	SimulateTX(ctx context.Context, req SimulateTXRequest) (*SimulateTXReply, error)
}
