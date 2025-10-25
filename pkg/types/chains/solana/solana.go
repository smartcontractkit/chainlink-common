package solana

import (
	"context"
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
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
	ConfidenceLevel primitives.ConfidenceLevel

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
	Data []byte // TODO

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

type TransactionWithMeta struct {
	// The slot this transaction was processed in.
	Slot uint64

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch)
	// of when the transaction was processed.
	// Nil if not available.
	BlockTime *UnixTimeSeconds

	Transaction *DataBytesOrJSON `json:"transaction"`

	// Transaction status metadata object
	Meta    *TransactionMeta   `json:"meta,omitempty"`
	Version TransactionVersion `json:"version"`
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
	ConfidenceLevel primitives.ConfidenceLevel

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

type GetBlockTimeRequest struct {
	//TODO
}

type GetBlockTimeReply struct {
	//TODO
}

type GetBalanceRequest struct {
	Addr            PublicKey
	ConfidenceLevel primitives.ConfidenceLevel
}

type GetBalanceReply struct {
	// Balance in lamports
	Value uint64
}

type GetBlockRequest struct {
	Slot uint64
	Opts *GetBlockOpts
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
	BlockTime *int64

	// The number of blocks beneath this block.
	BlockHeight *UnixTimeSeconds
}

type GetBlockHeightRequest struct {
	// TODO
}

type GetBlockHeightReply struct {
	// TODO
}

type GetTransactionRequest struct {
	// TODO
}

type GetTransactionReply struct {
	//TODO
}

type GetFeeForMessageRequest struct {
	//TODO
}

type GetFeeForMessageReply struct {
	//TODO
}

type GetLatestBlockhashRequest struct {
	// TODO
}

type GetLatestBlockhashReply struct {
	// TODO
}

type Client interface {
	GetAccountInfoWithOpts(ctx context.Context, req GetAccountInfoRequest) (*GetAccountInfoReply, error)
	GetMultipleAccountsWithOpts(ctx context.Context, req GetMultipleAccountsRequest) (*GetMultipleAccountsReply, error)
	GetBalance(ctx context.Context, req GetBalanceRequest) (*GetBalanceReply, error)
	GetBlock(ctx context.Context, req GetBlockRequest) (*GetBlockReply, error)
	GetBlockHeight(ctx context.Context, req GetBlockRequest) (*GetBlockHeightReply, error)
	GetBlockTime(ctx context.Context, req GetBlockTimeRequest) (*GetBlockTimeReply, error)
	GetTransaction(ctx context.Context, req GetTransactionRequest) (*GetTransactionReply, error)
	GetFeeForMessage(ctx context.Context, req GetFeeForMessageRequest) (*GetFeeForMessageReply, error)
	GetLatestBlockHash(ctx context.Context, req GetLatestBlockhashRequest) (*GetLatestBlockhashReply, error)
}

type SubmitTransactionRequest struct {
	Receiver PublicKey

	// base64 encoded transaction
	Payload string
}

// TransactionStatus is the result of the transaction sent to the chain
type TransactionStatus int

const (
// TODO define TransactionStatus enum
)

type SubmitTransactionReply struct {
	Signature      Signature
	IdempotencyKey string
	Status         TransactionStatus
}
