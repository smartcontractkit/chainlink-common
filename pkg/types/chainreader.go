package types

import (
	"time"
)

type BoundContract struct {
	Address string
	Name    string
	Pending bool
}

type Event struct {
	ChainId           string
	EventIndexInBlock string
	Address           string
	TxHash            string
	BlockHash         string
	BlockNumber       int64
	BlockTimestamp    time.Time
	CreatedAt         time.Time
	Keys              []string
	Data              []byte
}

type EventFilter struct {
	AddressList []string   // contract address
	KeysList    [][]string // 2D list of indexed search keys, outer dim = AND, inner dim = OR.  Params[0] is the name of the event (or "event type"), rest are any narrowing parameters that may help further specify the event
	Retention   time.Duration
}

// BlockID should be either a hash or a big int
type BlockID string

type EventQuery struct {
	FromBlock BlockID
	ToBlock   BlockID
	Filter    EventFilter
	Pending   bool
}
