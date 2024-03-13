package types

import (
	"context"
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

	//TODO Rebind binding address
	//ReBind(ctx context.Context, name, address string)

	QueryKey(ctx context.Context, keys string, queryFilter QueryFilter, limitAndSort LimitAndSort) ([]Sequence, error)
	QueryKeys(ctx context.Context, keys []string, queryFilter QueryFilter, limitAndSort LimitAndSort) ([][]Sequence, error)
	QueryKeyByValues(ctx context.Context, key string, values []string, queryFilter QueryFilter, limitAndSort LimitAndSort) ([]Sequence, error)
	QueryKeysByValues(ctx context.Context, keys []string, values [][]string, queryFilter QueryFilter, limitAndSort LimitAndSort) ([][]Sequence, error)

	// TODO some filters have to be dynamic, so this has to override chain reader bind that comes from config?
	// RegisterFilter()
	// UnRegisterFilter()

	// TODO make EVM words map to a key and then do this through the query methods.
	// GetCommitReportMatchingSeqNum()
	// GetSendRequestsBetweenSeqNums()
	// GetCommitReportGreaterThanSeqNum()
}

type BoundContract struct {
	Address string
	Name    string
	Pending bool
}

type Sequence struct {
	ChainID        string
	SequenceCursor string
	Timestamp      time.Time
	// TODO any or byte? Probably need to do codec transforms here too
	Data []byte
}

// TODO define If Register should be done outside of Binding, probably yes because of remapping
type KeysFilterer struct {
	Name string // see FilterName(id, args) below
	// TODO Retrieve key polling unique identifiers from chain reader config by using this identifier (evm eg. point to specific event sigs by contract name and event name)
	Identifier string
	//TODO Just Keys instead to do a similar thing?
	// Keys []string
	// Addresses [][]string
	// TODO may need to be mapped to event sigs a bit more creatively because of Solana? But we currently don't have Solana polling component so this is fine for now.
	Addresses []string
	Retention time.Duration
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

type SortBy interface {
	GetDirection() SortDirection
}

type LimitAndSort struct {
	SortBy []SortBy
	Limit  uint64
}

func NewLimitAndSort(limit uint64, sortBy ...SortBy) LimitAndSort {
	return LimitAndSort{SortBy: sortBy, Limit: limit}
}

type SortByTimestamp struct {
	dir SortDirection
}

func NewSortByTimestamp(sortDir SortDirection) SortByTimestamp {
	return SortByTimestamp{dir: sortDir}
}

func (o SortByTimestamp) GetDirection() SortDirection {
	return o.dir
}

type SortByBlock struct {
	dir SortDirection
}

func NewSortByBlock(sortDir SortDirection) SortByBlock {
	return SortByBlock{dir: sortDir}
}

func (o SortByBlock) GetDirection() SortDirection {
	return o.dir
}

type SortBySequence struct {
	dir SortDirection
}

func NewSortBySequence(sortDir SortDirection) SortBySequence {
	return SortBySequence{dir: sortDir}
}

func (o SortBySequence) GetDirection() SortDirection {
	return o.dir
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

func NewBasicAndFilter(filters ...QueryFilter) *AndFilter {
	allFilters := make([]QueryFilter, 0, len(filters)+2)
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
	Addresses []string
}

func NewAddressesFilter(addresses ...string) *AddressFilter {
	return &AddressFilter{Addresses: addresses}
}

func (f *AddressFilter) Accept(visitor Visitor) {
	visitor.VisitAddressFilter(*f)
}

type Confirmations int32

const (
	Finalized   = Confirmations(-1)
	Unconfirmed = Confirmations(0)
)

type ConfirmationsFilter struct {
	Confirmations
}

func NewConfirmationsFilter(confs Confirmations) *ConfirmationsFilter {
	return &ConfirmationsFilter{Confirmations: confs}
}

func (f *ConfirmationsFilter) Accept(visitor Visitor) {
	visitor.VisitConfirmationFilter(*f)
}

type BlockFilter struct {
	Block    uint64
	Operator ComparisonOperator
}

func NewBlockFilter(block uint64, operator ComparisonOperator) *BlockFilter {
	return &BlockFilter{block, operator}
}

func NewBlockRangeFilter(start, end uint64) *AndFilter {
	return NewAndFilter(
		NewBlockFilter(start, Gte),
		NewBlockFilter(end, Lte),
	)
}

func (f *BlockFilter) Accept(visitor Visitor) {
	visitor.VisitBlockFilter(*f)
}

func NewTimestampFilter(timestamp uint64, operator ComparisonOperator) *TimestampFilter {
	return &TimestampFilter{timestamp, operator}
}

type TimestampFilter struct {
	Timestamp uint64
	Operator  ComparisonOperator
}

func (f *TimestampFilter) Accept(visitor Visitor) {
	visitor.VisitTimestampFilter(*f)
}

func NewTxHashFilter(txHash string) *TxHashFilter {
	return &TxHashFilter{txHash}
}

type TxHashFilter struct {
	TxHash string
}

func (f *TxHashFilter) Accept(visitor Visitor) {
	visitor.VisitTxHashFilter(*f)
}

type Visitor interface {
	VisitAndFilter(filter AndFilter)
	VisitAddressFilter(filter AddressFilter)
	VisitBlockFilter(filter BlockFilter)
	VisitConfirmationFilter(filter ConfirmationsFilter)
	VisitTimestampFilter(filter TimestampFilter)
	VisitTxHashFilter(filter TxHashFilter)
}
