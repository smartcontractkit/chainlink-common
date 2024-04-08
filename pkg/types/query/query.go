package query

import "fmt"

// Visitor should have a per chain per db type implementation that converts primitives to db queries.
type Visitor interface {
	ComparerPrimitive(primitive ComparerPrimitive)
	BlockPrimitive(primitive BlockPrimitive)
	ConfirmationPrimitive(primitive ConfirmationsPrimitive)
	TimestampPrimitive(primitive TimestampPrimitive)
	TxHashPrimitives(primitive TxHashPrimitive)
}

// Primitive is the basic building block for Filter.
type Primitive interface {
	Accept(visitor Visitor)
}

// Filter is used to filter down chain specific data related to a key.
type Filter struct {
	// Key points to the underlying chain contract address and some data that belongs to that contract.
	// Depending on the underlying Chain Reader blockchain implementation key can map to different onchain concepts, but should be able to map differing onchain data to same offchain data if they belong to the same key.
	Key string
	// The base Expressions slice indicates AND logical operation over expressions, which can be primitives or nested boolean expressions.
	// For eg. []Expression{primitive, primitive, BoolExpression{AND, primitive, BoolExpression{OR, primitive, primitive}} is primitive AND primitive AND (primitive AND (primitive OR primitive)).
	Expressions []Expression
}

// Expression contains either a Primitive or a BoolExpression.
type Expression struct {
	Primitive      Primitive
	BoolExpression BoolExpression
}

func (expr Expression) IsPrimitive() bool {
	return expr.Primitive != nil
}

type BoolOperator int

const (
	AND BoolOperator = iota
	OR
)

func (op BoolOperator) String() string {
	switch op {
	case AND:
		return "AND"
	case OR:
		return "OR"
	default:
		return "Unknown"
	}
}

// BoolExpression allows nesting of boolean expressions with different BoolOperator's.
type BoolExpression struct {
	// should have minimum length of two
	Expressions []Expression
	BoolOperator
}

func And(expressions ...Expression) Expression {
	return Expression{
		BoolExpression: BoolExpression{Expressions: expressions, BoolOperator: AND},
	}
}

func Or(expressions ...Expression) Expression {
	return Expression{
		BoolExpression: BoolExpression{Expressions: expressions, BoolOperator: OR},
	}
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

type ValueComparer struct {
	Value    string
	Operator ComparisonOperator
}

// ComparerPrimitive is used to filter over values that belong to key data.
type ComparerPrimitive struct {
	Name           string
	ValueComparers []ValueComparer
}

// Comparer is used for filtering through specific key values.
// e.g. of filtering for key that belongs to a token transfer by values: Comparer("transferValue", [{"150",LTE}, {"300",GTE}])
func Comparer(name string, valueComparers ...ValueComparer) Expression {
	return Expression{
		Primitive: &ComparerPrimitive{Name: name, ValueComparers: valueComparers}}
}

func (f *ComparerPrimitive) Accept(visitor Visitor) {
	visitor.ComparerPrimitive(*f)
}

// BlockPrimitive is a primitive of Filter that filters search in comparison to block number.
type BlockPrimitive struct {
	Block    uint64
	Operator ComparisonOperator
}

func Block(block uint64, operator ComparisonOperator) Expression {
	return Expression{
		Primitive: &BlockPrimitive{Block: block, Operator: operator},
	}
}

func (f *BlockPrimitive) Accept(visitor Visitor) {
	visitor.BlockPrimitive(*f)
}

type ConfirmationLevel int32

const (
	Finalized   = ConfirmationLevel(0)
	Unconfirmed = ConfirmationLevel(1)
)

// ConfirmationsPrimitive is a primitive of Filter that filters search to results that have a certain level of confirmation.
// Confirmation map to different concepts on different blockchains.
type ConfirmationsPrimitive struct {
	ConfirmationLevel
}

func Confirmation(confLevel ConfirmationLevel) Expression {
	return Expression{
		Primitive: &ConfirmationsPrimitive{ConfirmationLevel: confLevel},
	}
}

func (f *ConfirmationsPrimitive) Accept(visitor Visitor) {
	visitor.ConfirmationPrimitive(*f)
}

// TimestampPrimitive is a primitive of Filter that filters search in comparison to timestamp.
type TimestampPrimitive struct {
	Timestamp uint64
	Operator  ComparisonOperator
}

func Timestamp(timestamp uint64, operator ComparisonOperator) Expression {
	return Expression{
		Primitive: &TimestampPrimitive{timestamp, operator},
	}
}

func (f *TimestampPrimitive) Accept(visitor Visitor) {
	visitor.TimestampPrimitive(*f)
}

// TxHashPrimitive is a primitive of Filter that filters search to results that contain txHash.
type TxHashPrimitive struct {
	TxHash string
}

func TxHash(txHash string) Expression {
	return Expression{
		Primitive: &TxHashPrimitive{txHash},
	}
}

func (f *TxHashPrimitive) Accept(visitor Visitor) {
	visitor.TxHashPrimitives(*f)
}

// Where is a helper function for building Filter, eg. usage:
//
//	 queryFilter, err := Where(
//
//				TxHash("0xHash"),
//				And(Block(startBlock, Gte),
//					Block(endBlock, Lte)),
//					Or(
//						And(
//							Timestamp(someTs1, Gte),
//							Timestamp(otherTs1, Lte)),
//						And(
//							Timestamp(someTs2, Gte),
//							Timestamp(otherTs2, Lte))
//					)
//			  	 )
//		 ==> `txHash = txHash AND (
//									 block > startBlock AND block < endBlock
//									 AND (
//										 (timestamp > someTs1 And timestamp < otherTs1)
//										 OR
//										 (timestamp > someTs2 And timestamp < otherTs2)
//									 )
//							 )`
//		if err != nil{return nil, err}
//		QueryOne(key, queryFilter)...
func Where(key string, expressions ...Expression) (Filter, error) {
	for _, expr := range expressions {
		if !expr.IsPrimitive() {
			if len(expr.BoolExpression.Expressions) < 2 {
				return Filter{}, fmt.Errorf("all boolean expressions should have at least 2 expressions")
			}
		}
	}
	return Filter{Key: key, Expressions: expressions}, nil
}

type SortDirection int

const (
	Asc SortDirection = iota
	Desc
)

type SortBy interface {
	GetDirection() SortDirection
}

type CursorDirection int32

const (
	Previous CursorDirection = iota
	Following
)

type Limit struct {
	Cursor          *string
	CursorDirection *CursorDirection
	Count           uint64
}

func CursorLimit(cursor string, cursorDirection CursorDirection, count uint64) Limit {
	return Limit{
		Cursor:          &cursor,
		CursorDirection: &cursorDirection,
		Count:           count,
	}
}

func CountLimit(count uint64) Limit {
	return Limit{Count: count}
}

type LimitAndSort struct {
	SortBy []SortBy
	Limit  Limit
}

func NewLimitAndSort(limit Limit, sortBy ...SortBy) LimitAndSort {
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
