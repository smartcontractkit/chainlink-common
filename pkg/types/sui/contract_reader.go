package sui

import (
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/types/aptos"
)

type EventFilterByMoveEventModule struct {
	Package string `json:"package"`
	Module  string `json:"module"`
	Event   string `json:"event"`
}

type EventSelector = EventFilterByMoveEventModule

type EventId struct {
	TxDigest string `json:"txDigest"`
	EventSeq string `json:"eventSeq"`
}

type ChainReaderConfig struct {
	IsLoopPlugin bool
	// NormalizeReturnValuesToHex, when set, converts base64-encoded byte/address
	// return values into 0x-prefixed hex strings. Defaults to false to preserve
	// behavior for consumers that decode return values into typed structs.
	// This is used by ccip-o11y and can be ignore otherwise.
	NormalizeReturnValuesToHex bool
	EventsIndexer              EventsIndexerConfig
	TransactionsIndexer        TransactionsIndexerConfig
	Modules                    map[string]*ChainReaderModule
}

type ChainReaderModule struct {
	// The module name (optional). When not provided, the key in the map under which this module
	// is stored is used.
	Name      string
	Functions map[string]*ChainReaderFunction
	Events    map[string]*ChainReaderEvent
}

type ChainReaderFunction struct {
	// The function name (optional). When not provided, the key in the map under which this function
	// is stored is used.
	Name          string
	SignerAddress string
	Params        []SuiFunctionParam
	// Defines a way to transform a tuple result into a JSON object
	ResultTupleToStruct []string
	// Defines a mapping for renaming response fields
	ResultFieldRenames map[string]aptos.RenamedField
	// Static response
	StaticResponse []any
	// Response from inputs
	ResponseFromInputs []string
}

type ChainReaderEvent struct {
	// The event name (optional). When not provided, the key in the map under which this event
	// is stored is used.
	Name      string
	EventType string
	// EventSelector specifies how the event is tagged within a package, and it includes
	// the 3 fields of the tag `packageId::moduleId::eventId`
	EventSelector

	// Renames of event field names (optional). When not provided, the field names are used as-is.
	EventFieldRenames map[string]aptos.RenamedField

	// Renames provided filters to match the event field names (optional). When not provided, the filters are used as-is.
	EventFilterRenames map[string]string

	// The expected event type (optional). When not provided, the event type is used as-is.
	ExpectedEventType any

	// A fallback for events selectors with no offset recorded in the DB and a starting point
	// earlier than the pruning cutoff of the RPC
	EventSelectorDefaultOffset *EventId
}

type EventsIndexerConfig struct {
	PollingInterval time.Duration
	SyncTimeout     time.Duration
}

type TransactionsIndexerConfig struct {
	PollingInterval time.Duration
	SyncTimeout     time.Duration
}

type ChainPollerConfig struct {
	PollingInterval         time.Duration
	SyncTimeout             time.Duration
	BackfillCheckpointCount *uint64 // optional: latest - N
	StartCheckpointSequence *uint64 // optional: explicit start (overrides backfill if set)
	ChannelBufferSize       int     // default e.g. 16
}
