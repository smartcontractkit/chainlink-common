package types

import (
	"context"
	"fmt"
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
	// TODO some filters have to be dynamic, so this has to override chain reader bind that comes from config?
	// Maube this should just be Bind?
	// RegisterFilter()
	// UnRegisterFilter()

	QueryKey(ctx context.Context, keys string, queryFilter QueryFilter, limitAndSort LimitAndSort, sequenceDataType any) ([]Sequence, error)
	QueryKeys(ctx context.Context, keys []string, queryFilter QueryFilter, limitAndSort LimitAndSort, sequenceDataTypes []any) ([][]Sequence, error)
	QueryKeyByValues(ctx context.Context, key string, values []string, queryFilter QueryFilter, limitAndSort LimitAndSort, sequenceDataType any) ([]Sequence, error)
	QueryKeysByValues(ctx context.Context, keys []string, values [][]string, queryFilter QueryFilter, limitAndSort LimitAndSort, sequenceDataTypes []any) ([][]Sequence, error)

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
	// TODO SequenceCursor, this should be a unique sequence identifier that chain reader impl. understands.
	// This way we can retrieve past/future sequences (EVM log events) very granularly but still hiding the chain detail.
	SequenceCursor string
	Timestamp      uint64
	// TODO this is a general chain agnositc Head, it should be moved from mercury package.
	//mercury.Head
	Data any
}

// TODO define If Register should be done outside of Binding?
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

type Primitive interface {
	Accept(visitor Visitor)
}

// QueryFilter can translate to any combination of nested OR and AND boolean expressions.
// The base Expressions slice indicates AND logical operation over expressions, which can be primitives or nested boolean expressions.
// eg. []Expression{primitive, primitive, BooleanExpression{AND, primitive, BooleanExpression{OR, primitive, primitive}} is
// primitive AND primitive AND (primitive AND (primitive OR primitive)).
type QueryFilter struct {
	Expressions []Expression
}

// Expression contains either a Primitive or a BooleanExpression.
type Expression struct {
	Primitive         Primitive
	BooleanExpression BooleanExpression
}

func (expr Expression) IsPrimitive() bool {
	return expr.Primitive != nil
}

type BooleanOperator int

const (
	AND BooleanOperator = iota
	OR
)

func (op BooleanOperator) String() string {
	switch op {
	case AND:
		return "AND"
	case OR:
		return "OR"
	default:
		return "Unknown"
	}
}

// BooleanExpression allows nesting of boolean expressions with different BooleanOperator's.
type BooleanExpression struct {
	// should have minimum length of two
	Expressions []Expression
	BooleanOperator
}

func NewBooleanExpression(operator BooleanOperator, expressions []Expression) (Expression, error) {
	if len(expressions) < 2 {
		return Expression{}, fmt.Errorf("boolean expression %s has to contain at least two expressions", operator)
	}

	return Expression{
		BooleanExpression: BooleanExpression{Expressions: expressions, BooleanOperator: operator},
	}, nil
}

// BlockFilter is a primitive of QueryFilter that filters search in comparison to block number.
type BlockFilter struct {
	Block    uint64
	Operator ComparisonOperator
}

func NewBlockPrimitive(block uint64, operator ComparisonOperator) Expression {
	return Expression{
		Primitive: &BlockFilter{Block: block, Operator: operator},
	}
}

func (f *BlockFilter) Accept(visitor Visitor) {
	visitor.VisitBlockFilter(*f)
}

// AddressFilter is a primitive of QueryFilter that filters search to results that contain address in Addresses.
type AddressFilter struct {
	Addresses []string
}

func NewAddressesPrimitive(addresses ...string) Expression {
	return Expression{
		Primitive: &AddressFilter{Addresses: addresses},
	}
}

func (f *AddressFilter) Accept(visitor Visitor) {
	visitor.VisitAddressFilter(*f)
}

type Confirmations int32

const (
	Finalized   = Confirmations(0)
	Unconfirmed = Confirmations(1)
)

// ConfirmationsFilter is a primitive of QueryFilter that filters search to results that have a certain level of confirmation.
// Confirmations map to different concepts on different blockchains.
type ConfirmationsFilter struct {
	Confirmations
}

func NewConfirmationsPrimitive(confs Confirmations) Expression {
	return Expression{
		Primitive: &ConfirmationsFilter{Confirmations: confs},
	}
}

func (f *ConfirmationsFilter) Accept(visitor Visitor) {
	visitor.VisitConfirmationFilter(*f)
}

// TimestampFilter is a primitive of QueryFilter that filters search in comparison to timestamp.
type TimestampFilter struct {
	Timestamp uint64
	Operator  ComparisonOperator
}

func NewTimestampPrimitive(timestamp uint64, operator ComparisonOperator) Expression {
	return Expression{
		Primitive: &TimestampFilter{timestamp, operator},
	}
}

func (f *TimestampFilter) Accept(visitor Visitor) {
	visitor.VisitTimestampFilter(*f)
}

// TxHashFilter is a primitive of QueryFilter that filters search to results that contain txHash.
type TxHashFilter struct {
	TxHash string
}

func NewTxHashPrimitive(txHash string) Expression {
	return Expression{
		Primitive: &TxHashFilter{txHash},
	}
}

func (f *TxHashFilter) Accept(visitor Visitor) {
	visitor.VisitTxHashFilter(*f)
}

// Where eg. usage:
// queryFilter := Where(
//
//		NewTxHashPrimitive("0xHash"),
//		NewBooleanExpression("OR",
//			NewBlockPrimitive(startBlock, Gte),
//			NewBlockPrimitive(endBlock, Lte)),
//		NewBooleanExpression("AND",
//			NewBooleanExpression("OR",
//				NewTimestampPrimitive(someTs1, Gte),
//				NewTimestampPrimitive(otherTs1, Lte)),
//			NewBooleanExpression("OR",(endBlock, Lte)),
//				NewTimestampPrimitive(someTs2, Gte),
//				NewTimestampPrimitive(otherTs2, Lte)))
//	   )
//
// QueryKey(key, queryFilter)...
func Where(expressions ...Expression) QueryFilter {
	return QueryFilter{expressions}
}

type Visitor interface {
	VisitAddressFilter(filter AddressFilter)
	VisitBlockFilter(filter BlockFilter)
	VisitConfirmationFilter(filter ConfirmationsFilter)
	VisitTimestampFilter(filter TimestampFilter)
	VisitTxHashFilter(filter TxHashFilter)
}
