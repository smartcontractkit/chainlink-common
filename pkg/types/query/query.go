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

func NewBoolExpression(operator BoolOperator, expressions ...Expression) Expression {
	return Expression{
		BoolExpression: BoolExpression{Expressions: expressions, BoolOperator: operator},
	}
}

// BlockPrimitive is a primitive of Filter that filters search in comparison to block number.
type BlockPrimitive struct {
	Block    uint64
	Operator ComparisonOperator
}

func NewBlockPrimitive(block uint64, operator ComparisonOperator) Expression {
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

func NewAddressesPrimitive(addresses ...string) Expression {
	return Expression{
		Primitive: &AddressPrimitive{Addresses: addresses},
	}
}

func (f *AddressPrimitive) Accept(visitor Visitor) {
	visitor.AddressPrimitive(*f)
}

type Confirmations int32

const (
	Finalized   = Confirmations(0)
	Unconfirmed = Confirmations(1)
)

// ConfirmationsPrimitive is a primitive of Filter that filters search to results that have a certain level of confirmation.
// Confirmations map to different concepts on different blockchains.
type ConfirmationsPrimitive struct {
	Confirmations
}

func NewConfirmationsPrimitive(confs Confirmations) Expression {
	return Expression{
		Primitive: &ConfirmationsPrimitive{Confirmations: confs},
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

func NewTimestampPrimitive(timestamp uint64, operator ComparisonOperator) Expression {
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

func NewTxHashPrimitive(txHash string) Expression {
	return Expression{
		Primitive: &TxHashPrimitive{txHash},
	}
}

func (f *TxHashPrimitive) Accept(visitor Visitor) {
	visitor.TxHashPrimitives(*f)
}

// Where is a helper function for building Filter, eg. usage:
// queryFilter, err := Where(
//
//		NewTxHashPrimitive("0xHash"),
//		NewBoolExpression("OR",
//			NewBlockPrimitive(startBlock, Gte),
//			NewBlockPrimitive(endBlock, Lte)),
//		NewBoolExpression("AND",
//			NewBoolExpression("OR",
//				NewTimestampPrimitive(someTs1, Gte),
//				NewTimestampPrimitive(otherTs1, Lte)),
//			NewBoolExpression("OR",(endBlock, Lte)),
//				NewTimestampPrimitive(someTs2, Gte),
//				NewTimestampPrimitive(otherTs2, Lte)))
//	   )
//	if err != nil{return nil, err}
//
// QueryKey(key, queryFilter)...
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
