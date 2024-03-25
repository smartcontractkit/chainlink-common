package query

import "fmt"

type ValueComparator struct {
	Value    string
	Operator ComparisonOperator
}

type KeyValuesComparator struct {
	Key              string
	ValueComparators []ValueComparator
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

// Visitor should have a per chain per db type implementation that converts primitives to db queries.
type Visitor interface {
	AddressPrimitive(primitive AddressPrimitive)
	BlockPrimitive(primitive BlockPrimitive)
	ConfirmationPrimitive(primitive ConfirmationsPrimitive)
	TimestampPrimitive(primitive TimestampPrimitive)
	TxHashPrimitives(primitive TxHashPrimitive)
}

// Primitive is the basic building block for Filter.
type Primitive interface {
	Accept(visitor Visitor)
}

// Filter can translate to any combination of nested OR and AND boolean expressions.
// The base Expressions slice indicates AND logical operation over expressions, which can be primitives or nested boolean expressions.
// eg. []Expression{primitive, primitive, BoolExpression{AND, primitive, BoolExpression{OR, primitive, primitive}} is
// primitive AND primitive AND (primitive AND (primitive OR primitive)).
type Filter struct {
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

// AddressPrimitive is a primitive of Filter that filters search to results that contain address in Addresses.
type AddressPrimitive struct {
	Addresses []string
}

// TODO adress is fetched through bindings, so this is most likely not necessary
func Address(addresses ...string) Expression {
	return Expression{
		Primitive: &AddressPrimitive{Addresses: addresses},
	}
}

func (f *AddressPrimitive) Accept(visitor Visitor) {
	visitor.AddressPrimitive(*f)
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
//		QueryKey(key, queryFilter)...
func Where(expressions ...Expression) (Filter, error) {
	for _, expr := range expressions {
		if !expr.IsPrimitive() {
			if len(expr.BoolExpression.Expressions) < 2 {
				return Filter{}, fmt.Errorf("all boolean expressions should have at least 2 expressions")
			}
		}
	}
	return Filter{expressions}, nil
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
