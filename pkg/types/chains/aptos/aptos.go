package aptos

import "context"

const (
	AccountAddressLength = 32
)

type AccountAddress [AccountAddressLength]byte

// Client wraps the Aptos RPC client methods used for reading on-chain state.
type Client interface {
	// AccountAPTBalance returns the native APT coin balance (in octas) for the given account address.
	AccountAPTBalance(ctx context.Context, req AccountAPTBalanceRequest) (*AccountAPTBalanceReply, error)
	// View executes a Move view function (read-only) and returns the raw result.
	View(ctx context.Context, req ViewRequest) (*ViewReply, error)
	// TransactionByHash looks up a transaction (pending or committed) by its hash.
	TransactionByHash(ctx context.Context, req TransactionByHashRequest) (*TransactionByHashReply, error)
	// AccountTransactions returns committed transactions associated with an account.
	AccountTransactions(ctx context.Context, req AccountTransactionsRequest) (*AccountTransactionsReply, error)
}

// ========== AccountAPTBalance ==========

type AccountAPTBalanceRequest struct {
	Address AccountAddress
}

type AccountAPTBalanceReply struct {
	Value uint64
}

// ========== View ==========

type ViewRequest struct {
	Payload *ViewPayload
}

type ViewReply struct {
	Data []byte // this is marshaled JSON because the aptos rpc client returns JSON
}

// ViewPayload represents the payload for a view function call.
type ViewPayload struct {
	Module   ModuleID
	Function string
	ArgTypes []TypeTag
	Args     [][]byte // Arguments encoded in BCS
}

// ModuleID identifies a Move module.
type ModuleID struct {
	Address AccountAddress
	Name    string
}

// TypeTag is a wrapper around a TypeTagImpl for serialization/deserialization.
// This represents Move type arguments like u64, address, vector<u8>, etc.
type TypeTag struct {
	Value TypeTagImpl
}

// TypeTagImpl is the interface for all type tag implementations.
// Different type tags represent different Move types.
type TypeTagImpl interface {
	// TypeTagType returns the type discriminator for this type tag.
	TypeTagType() TypeTagType
}

// TypeTagType is an enum for different type tag variants.
type TypeTagType uint8

const (
	TypeTagBool TypeTagType = iota
	TypeTagU8
	TypeTagU16
	TypeTagU32
	TypeTagU64
	TypeTagU128
	TypeTagU256
	TypeTagAddress
	TypeTagSigner
	TypeTagVector
	TypeTagStruct
	TypeTagGeneric
)

// ========== Type Tag Implementations ==========

// BoolTag represents a boolean type.
type BoolTag struct{}

func (BoolTag) TypeTagType() TypeTagType { return TypeTagBool }

// U8Tag represents an unsigned 8-bit integer type.
type U8Tag struct{}

func (U8Tag) TypeTagType() TypeTagType { return TypeTagU8 }

// U16Tag represents an unsigned 16-bit integer type.
type U16Tag struct{}

func (U16Tag) TypeTagType() TypeTagType { return TypeTagU16 }

// U32Tag represents an unsigned 32-bit integer type.
type U32Tag struct{}

func (U32Tag) TypeTagType() TypeTagType { return TypeTagU32 }

// U64Tag represents an unsigned 64-bit integer type.
type U64Tag struct{}

func (U64Tag) TypeTagType() TypeTagType { return TypeTagU64 }

// U128Tag represents an unsigned 128-bit integer type.
type U128Tag struct{}

func (U128Tag) TypeTagType() TypeTagType { return TypeTagU128 }

// U256Tag represents an unsigned 256-bit integer type.
type U256Tag struct{}

func (U256Tag) TypeTagType() TypeTagType { return TypeTagU256 }

// AddressTag represents an account address type.
type AddressTag struct{}

func (AddressTag) TypeTagType() TypeTagType { return TypeTagAddress }

// SignerTag represents a signer type.
type SignerTag struct{}

func (SignerTag) TypeTagType() TypeTagType { return TypeTagSigner }

// VectorTag represents a vector type with an element type.
type VectorTag struct {
	ElementType TypeTag
}

func (VectorTag) TypeTagType() TypeTagType { return TypeTagVector }

// StructTag represents a struct type with full type information.
type StructTag struct {
	Address    AccountAddress
	Module     string
	Name       string
	TypeParams []TypeTag
}

func (StructTag) TypeTagType() TypeTagType { return TypeTagStruct }

// GenericTag represents a generic type parameter (e.g., T in a generic function).
type GenericTag struct {
	Index uint16
}

func (GenericTag) TypeTagType() TypeTagType { return TypeTagGeneric }

// TransactionByHashRequest represents a request to get a transaction by hash
type TransactionByHashRequest struct {
	Hash string // Transaction hash (hex string with 0x prefix)
}

// TransactionByHashReply represents the response from TransactionByHash
type TransactionByHashReply struct {
	Transaction *Transaction
}

// TransactionVariant represents the type of transaction
type TransactionVariant string

const (
	TransactionVariantPending         TransactionVariant = "pending_transaction"
	TransactionVariantUser            TransactionVariant = "user_transaction"
	TransactionVariantGenesis         TransactionVariant = "genesis_transaction"
	TransactionVariantBlockMetadata   TransactionVariant = "block_metadata_transaction"
	TransactionVariantBlockEpilogue   TransactionVariant = "block_epilogue_transaction"
	TransactionVariantStateCheckpoint TransactionVariant = "state_checkpoint_transaction"
	TransactionVariantValidator       TransactionVariant = "validator_transaction"
	TransactionVariantUnknown         TransactionVariant = "unknown"
)

// Transaction represents any transaction type (pending or committed)
type Transaction struct {
	Type    TransactionVariant
	Hash    string
	Version *uint64 // nil for pending transactions
	Success *bool   // nil for pending/genesis transactions
	Data    []byte  // Raw transaction data
}

// ========== AccountTransactions ==========

type AccountTransactionsRequest struct {
	Address AccountAddress
	Start   *uint64 // Starting version number; nil for most recent
	Limit   *uint64 // Number of transactions to return; nil for default (~100)
}

type AccountTransactionsReply struct {
	Transactions []*Transaction
}

// ========== SubmitTransaction ==========

type SubmitTransactionRequest struct {
	ReceiverModuleID ModuleID // This can potentially be removed if the EncodedPayload is of type EntryFunction which has all the details
	EncodedPayload   []byte
	GasConfig        *GasConfig
}

type TransactionStatus int

const (
	// Transaction processing failed due to a network issue, RPC issue, or other fatal error
	TxFatal TransactionStatus = iota
	// Transaction was sent successfully to the chain but the smart contract execution reverted
	TxReverted
	// Transaction was sent successfully to the chain, smart contract executed successfully and mined into a block.
	TxSuccess
)

type SubmitTransactionReply struct {
	TxStatus         TransactionStatus
	TxHash           string
	TxIdempotencyKey string
}

// GasConfig represents gas configuration for a transaction
type GasConfig struct {
	MaxGasAmount uint64 // Maximum gas units willing to pay
	GasUnitPrice uint64 // Price per gas unit in octas
}
