package stellar

import "context"

// XDR is a base64-encoded XDR-serialized value (envelope, data, result, key, etc.).
type XDR string

// LedgerHash is the hex-encoded SHA-256 hash of a Stellar ledger.
type LedgerHash string

// Client wraps Stellar RPC calls via the type/chains/stellar domain types.
// Both methods map 1:1 to the Stellar RPC API.
type Client interface {
	// GetLedgerEntries fetches ledger entries by XDR key (used for sequence number lookups).
	GetLedgerEntries(ctx context.Context, req GetLedgerEntriesRequest) (GetLedgerEntriesResponse, error)
	// GetLatestLedger returns current ledger info (used for timeout detection).
	GetLatestLedger(ctx context.Context) (GetLatestLedgerResponse, error)
}

// GetLedgerEntriesRequest fetches ledger entries by XDR-encoded keys.
type GetLedgerEntriesRequest struct {
	// Keys is a slice of base64-encoded XDR ledger keys.
	Keys []XDR
}

// LedgerEntryResult is a single ledger entry returned from GetLedgerEntries.
type LedgerEntryResult struct {
	// KeyXDR is the base64-encoded XDR ledger key matching the request.
	KeyXDR XDR
	// DataXDR is the base64-encoded XDR ledger entry data.
	DataXDR XDR
	// LastModifiedLedger is the ledger sequence of the last modification.
	LastModifiedLedger uint32
	// LiveUntilLedgerSeq is the ledger until which the entry is live; nil if not applicable.
	LiveUntilLedgerSeq *uint32
	// ExtensionXDR is the base64-encoded XDR ledger entry extension; empty if absent.
	ExtensionXDR XDR
}

// GetLedgerEntriesResponse contains the requested ledger entries.
type GetLedgerEntriesResponse struct {
	// Entries holds all found ledger entries (may be fewer than keys requested).
	Entries []LedgerEntryResult
	// LatestLedger is the latest ledger sequence number at query time.
	LatestLedger uint32
}

// GetLatestLedgerResponse holds the current ledger state.
type GetLatestLedgerResponse struct {
	// Hash is the hex-encoded latest ledger hash.
	Hash LedgerHash
	// ProtocolVersion is the Stellar Core protocol version associated with the ledger.
	ProtocolVersion uint32
	// Sequence is the latest ledger sequence number.
	Sequence uint32
	// LedgerCloseTime is the unix timestamp when the latest ledger closed.
	LedgerCloseTime int64
	// LedgerHeaderXDR is the base64-encoded LedgerHeader XDR for the latest ledger.
	LedgerHeaderXDR XDR
	// LedgerMetadataXDR is the base64-encoded LedgerCloseMetaV2 XDR for the latest ledger.
	LedgerMetadataXDR XDR
}
