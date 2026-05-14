package stellar

import (
	"context"

	"github.com/stellar/go-stellar-sdk/xdr"
)

// Client wraps Stellar RPC calls via the type/chains/stellar domain types.
// Both methods map 1:1 to the Stellar RPC API.
type Client interface {
	// GetLedgerEntries fetches ledger entries by XDR key (used for sequence number lookups).
	GetLedgerEntries(ctx context.Context, req GetLedgerEntriesRequest) (GetLedgerEntriesResponse, error)
	// GetLatestLedger returns current ledger info (used for timeout detection).
	GetLatestLedger(ctx context.Context) (GetLatestLedgerResponse, error)
	// ReadContract simulates a read-only Soroban contract function call.
	// Each element of Args is an XDR ScVal value.
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
// Use the proto helpers to construct XDR ScVal values conveniently.
type ReadContractRequest struct {
	// ContractID is the Stellar contract address in C… StrKey encoding.
	ContractID string
	// Function is the Soroban function name to call.
	Function string
	// Args holds one XDR ScVal per contract argument.
	// An empty slice is valid for zero-argument functions.
	Args []xdr.ScVal
	// LedgerSequence is the ledger to simulate against; 0 means use the latest.
	LedgerSequence uint32
}

// ReadContractResponse is the domain representation of a Soroban simulation result.
type ReadContractResponse struct {
	// Result is the XDR ScVal returned by the contract (nil when Error is non-empty).
	Result *xdr.ScVal
	// LedgerSequence is the ledger that was used for the simulation.
	LedgerSequence uint32
	// Error is non-empty when the call failed.
	Error string
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
