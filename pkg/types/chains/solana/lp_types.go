package solana

import "time"

const EventSignatureLength = 8

type EventSignature [EventSignatureLength]byte

// matches solana lp-types IndexedValue
type IndexedValue []byte

type SubKeyPaths [][]string

// matches cache-filter
// this filter defines what logs should be cached
// cached logs can be retrieved with [types.SolanaService.QueryTrackedLogs]
type LPFilterQuery struct {
	Name          string
	Address       PublicKey
	EventName     string
	EventSig      EventSignature
	StartingBlock int64
	// Deprecated: Use ContractIdlJSON instead
	EventIdlJSON    []byte
	ContractIdlJSON []byte
	SubkeyPaths     SubKeyPaths
	Retention       time.Duration
	MaxLogsKept     int64
	IncludeReverted bool
}

// matches lp-parsed solana logs
type Log struct {
	ChainID        string
	LogIndex       int64
	BlockHash      Hash
	BlockNumber    int64
	BlockTimestamp uint64
	Address        PublicKey
	EventSig       EventSignature
	TxHash         Signature
	Data           []byte
	SequenceNum    int64
	Error          *string
}

type LPBlock struct {
	Slot uint64
}
