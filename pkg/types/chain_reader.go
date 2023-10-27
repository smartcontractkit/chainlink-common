package types

import (
	"context"
	"time"
)

/*
type BigInt big.Int
type String string

// Union of BigInt & String
type BlockID interface {
	bi() *BigInt
	s() String
}

func (b *BigInt) bi() *BigInt {
	return b
}

func (b String) s() String {
	return b
}
*/

type ChainReader interface {
	//RegisterEventFilter(ctx context.Context, filterName string, filter EventFilter, startingBlock BlockID) error
	//UnregisterEventFilter(ctx context.Context, filterName string) error
	//QueryEvents(ctx context.Context, query EventQuery) ([]Event, error)
	// The params argument of GetLatestValue() can be any object which maps a set of generic parameters into chain specific parameters defined in RelayConfig. It must be something which can be passed into
	// json.Marshal(). Typically would be either an anonymous map such as `map[string]any{"baz": 42, "test": true}}` or something which implements the `MarshalJSON()` method (satisfying `Marshaller` interface).
	//
	// returnVal should satisfy Marshaller interface.
	GetLatestValue(ctx context.Context, bc BoundContract, method string, params, returnVal any) error
}

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

//type EventQuery struct {
//	FromBlock BlockID
//	ToBlock   BlockID
//	Filter    EventFilter
//	Pending   bool
//}
