package types

import (
	"context"
	"math/big"
	"time"
)

// Errors exposed to product plugins
const (
	ErrInvalidType              = InvalidArgumentError("invalid type")
	ErrInvalidConfig            = InvalidArgumentError("invalid configuration")
	ErrChainReaderConfigMissing = UnimplementedError("ChainReader entry missing from RelayConfig")
	ErrInternal                 = InternalError("internal error")
	ErrNotFound                 = NotFoundError("not found")
)

type ChainReader interface {
	// GetLatestValue gets the latest value....
	// The params argument can be any object which maps a set of generic parameters into chain specific parameters defined in RelayConfig.
	// It must encode as an object via [json.Marshal] and [github.com/fxamacker/cbor/v2.Marshal].
	// Typically, would be either a struct with field names mapping to arguments, or anonymous map such as `map[string]any{"baz": 42, "test": true}}`
	//
	// returnVal must [json.Unmarshal] and and [github.com/fxamacker/cbor/v2.Marshal] as an object.
	//
	// Example use:
	//  type ProductParams struct {
	// 		ID int `json:"id"`
	//  }
	//  type ProductReturn struct {
	// 		Foo string `json:"foo"`
	// 		Bar *big.Int `json:"bar"`
	//  }
	//  func do(ctx context.Context, cr ChainReader) (resp ProductReturn, err error) {
	// 		err = cr.GetLatestValue(ctx, "FooContract", "GetProduct", ProductParams{ID:1}, &resp)
	// 		return
	//  }
	//
	// Note that implementations should ignore extra fields in params that are not expected in the call to allow easier
	// use across chains and contract versions.
	// Similarly, when using a struct for returnVal, fields in the return value that are not on-chain will not be set.
	GetLatestValue(ctx context.Context, contractName string, method string, params, returnVal any) error

	// Bind will override current bindings for the same contract, if one has been set and will return an error if the
	// contract is not known by the ChainReader, or if the Address is invalid
	Bind(ctx context.Context, bindings []BoundContract) error

	QueryKeys(ctx context.Context, queryFilter QueryFilter) ([]Event, error)

	//// give it same event filter and
	//QueryKeysByValue()
	//QueryKeysExcluding()
	//RegisterFilter()
	//
	////  we don't want to tackle querying for arbitrary data right now so these should do
	//// CCIP specific evm log data searching, these can be made into one function
	//GetCommitReportMatchingSeqNum()
	//GetSendRequestsBetweenSeqNums()
	//GetCommitReportGreaterThanSeqNum()
}

type BoundContract struct {
	Address string
	Name    string
	Pending bool
}

type Event struct {
	ChainID        string
	SequenceCursor string
	timestamp      time.Time
	Data           []byte
}

// Multiple EventFilters per QueryEvent call?

type EventFilter struct {
	Name        string     // ??
	AddressList []string   // contract Address
	KeysList    [][]string // 2D list of indexed search Keys, outer dim = AND, inner dim = OR.  Params[0] is the name of the event (or "event type"), rest are any narrowing parameters that may help further specify the event
	Retention   time.Duration
}

type EventQuery struct {
	TxHash             string
	FromBlock, ToBlock *big.Int // either a hash or a big int
	FromTime, ToTime   time.Time
	Filter             EventFilter
	Finality           string
}

type ComparisonOperator int

const (
	Eq ComparisonOperator = iota
	Neq
	Gt
	Lt
	Gte
	Lte
)

type SortDirection int

const (
	Asc SortDirection = iota
	Desc
)

var (
	DefaultSortAndLimit = SortAndLimit{
		sortBy: []SortBy{
			{field: "block_number", dir: Asc},
			// sequence should convert to chain specific sequence unique identifier
			{field: "sequence", dir: Asc},
		},
		limit: 0,
	}
)

type SortAndLimit struct {
	sortBy []SortBy
	limit  int
}

type SortBy struct {
	field string
	dir   SortDirection
}

func NewSortAndLimit(limit int, sortBy ...SortBy) SortAndLimit {
	return SortAndLimit{sortBy: sortBy, limit: limit}
}

func NewSortBy(field string, dir SortDirection) SortBy {
	return SortBy{field: field, dir: dir}
}

type QueryFilter interface {
	Accept(visitor Visitor)
}

type AndFilter struct {
	Filters []QueryFilter
}

func NewAndFilter(filters ...QueryFilter) *AndFilter {
	return &AndFilter{Filters: filters}
}

func NewBasicAndFilter(address string, eventSig string, filters ...QueryFilter) *AndFilter {
	allFilters := make([]QueryFilter, 0, len(filters)+2)
	allFilters = append(allFilters, NewAddressFilter(address), NewKeysFilter(eventSig))
	allFilters = append(allFilters, filters...)
	return NewAndFilter(allFilters...)
}

func AppendedNewFilter(root *AndFilter, other ...QueryFilter) *AndFilter {
	filters := make([]QueryFilter, 0, len(root.Filters)+len(other))
	filters = append(filters, root.Filters...)
	filters = append(filters, other...)
	return NewAndFilter(filters...)
}

func (f *AndFilter) Accept(visitor Visitor) {
	visitor.VisitAndFilter(*f)
}

type AddressFilter struct {
	Address []string
}

func NewAddressFilter(address ...string) *AddressFilter {
	return &AddressFilter{Address: address}
}

func (f *AddressFilter) Accept(visitor Visitor) {
	visitor.VisitAddressFilter(*f)
}

type KeysFilter struct {
	keys []string
}

func NewKeysFilter(keys ...string) *KeysFilter {
	return &KeysFilter{keys: keys}
}

func (f *KeysFilter) Accept(visitor Visitor) {
	visitor.VisitKeysFilter(*f)
}

type KeysByValueFilter struct {
	Keys   []string
	Values [][]string
}

func NewKeysByValueFilter(keys []string, values [][]string) *KeysByValueFilter {
	return &KeysByValueFilter{Keys: keys, Values: values}
}

func (f *KeysByValueFilter) Accept(visitor Visitor) {
	visitor.VisitKeysByValueFilter(*f)
}

type ConfirmationFilter struct {
	confs string
}

func NewConfirmationFilter(confs string) *ConfirmationFilter {
	return &ConfirmationFilter{confs: confs}
}

func (f *ConfirmationFilter) Accept(visitor Visitor) {
	visitor.VisitConfirmationFilter(*f)
}

func NewBlockFilter(block int64, operator ComparisonOperator) *BlockFilter {
	return &BlockFilter{operator, block}
}

func NewBlockRangeFilter(start, end int64) *AndFilter {
	return NewAndFilter(
		NewBlockFilter(start, Gte),
		NewBlockFilter(end, Lte),
	)
}

type BlockFilter struct {
	operator ComparisonOperator
	block    int64
}

func (f *BlockFilter) Accept(visitor Visitor) {
	visitor.VisitBlockFilter(*f)
}

func NewTimeStampFilter(timestamp time.Time, operator ComparisonOperator) *TimestampFilter {
	return &TimestampFilter{operator, timestamp}
}

type TimestampFilter struct {
	operator  ComparisonOperator
	timestamp time.Time
}

func (f *TimestampFilter) Accept(visitor Visitor) {
	visitor.VisitTimestampFilter(*f)
}

func NewTxHashFilter(txHash string) *TxHashFilter {
	return &TxHashFilter{txHash}
}

type TxHashFilter struct {
	txHash string
}

func (f *TxHashFilter) Accept(visitor Visitor) {
	visitor.VisitTxHashFilter(*f)
}

type Visitor interface {
	VisitAndFilter(node AndFilter)
	VisitAddressFilter(node AddressFilter)
	VisitKeysFilter(node KeysFilter)
	VisitKeysByValueFilter(node KeysByValueFilter)
	VisitBlockFilter(node BlockFilter)
	VisitConfirmationFilter(node ConfirmationFilter)
	VisitTimestampFilter(node TimestampFilter)
	VisitTxHashFilter(node TxHashFilter)
}
