package query

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

// KeyFilter is used to filter down chain specific data related to a key.
type KeyFilter struct {
	// Key points to the underlying chain contract address and some data that belongs to that contract.
	// Depending on the underlying Contract Reader blockchain implementation key can map to different onchain concepts, but should be able to map differing onchain data to same offchain data if they belong to the same key.
	Key string
	// The base Expressions slice indicates AND logical operation over expressions, which can be primitives or nested boolean expressions.
	// For eg. []Expression{primitive, primitive, BoolExpression{AND, primitive, BoolExpression{OR, primitive, primitive}} is primitive AND primitive AND (primitive AND (primitive OR primitive)).
	Expressions []Expression
}

// Expression contains either a Primitive or a BoolExpression.
type Expression struct {
	Primitive      primitives.Primitive
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

// Comparator is used for filtering through specific key values.
// e.g. of filtering for key that belongs to a token transfer by values: Comparator("transferValue", [{"150",LTE}, {"300",GTE}])
func Comparator(name string, valueComparators ...primitives.ValueComparator) Expression {
	return Expression{Primitive: &primitives.Comparator{Name: name, ValueComparators: valueComparators}}
}

func Block(block string, operator primitives.ComparisonOperator) Expression {
	return Expression{
		Primitive: &primitives.Block{Block: block, Operator: operator},
	}
}

func Confidence(confLevel primitives.ConfidenceLevel) Expression {
	return Expression{
		Primitive: &primitives.Confidence{ConfidenceLevel: confLevel},
	}
}

func Timestamp(timestamp uint64, operator primitives.ComparisonOperator) Expression {
	return Expression{
		Primitive: &primitives.Timestamp{Timestamp: timestamp, Operator: operator},
	}
}

func TxHash(txHash string) Expression {
	return Expression{
		Primitive: &primitives.TxHash{TxHash: txHash},
	}
}

func And(expressions ...Expression) Expression {
	if len(expressions) == 1 {
		return expressions[0]
	}
	return Expression{
		BoolExpression: BoolExpression{Expressions: expressions, BoolOperator: AND},
	}
}

func Or(expressions ...Expression) Expression {
	if len(expressions) == 1 {
		return expressions[0]
	}
	return Expression{
		BoolExpression: BoolExpression{Expressions: expressions, BoolOperator: OR},
	}
}

// Where is a helper function for building KeyFilter, eg. usage:
//
//	queryFilter, err := Where(
//			"key",
//			TxHash("0xHash"),
//			Block(startBlock, Gte),
//			Block(endBlock, Lte),
//			Or(
//				And(
//					Timestamp(someTs1, Gte),
//					Timestamp(otherTs1, Lte)),
//				And(
//					Timestamp(someTs2, Gte),
//					Timestamp(otherTs2, Lte)),
//			),
//		)
//	Equals:`txHash = '0xHash' AND
//			block > startBlock AND
//			block < endBlock   AND
//			((timestamp > someTs1 And timestamp < otherTs1)
//				OR
//			(timestamp > someTs2 And timestamp < otherTs2))`
func Where(key string, expressions ...Expression) (KeyFilter, error) {
	for _, expr := range expressions {
		if !expr.IsPrimitive() {
			if len(expr.BoolExpression.Expressions) < 2 {
				return KeyFilter{}, errors.New("all boolean expressions should have at least 2 expressions")
			}
		}
	}
	return KeyFilter{Key: key, Expressions: expressions}, nil
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
	CursorPrevious CursorDirection = iota + 1
	CursorFollowing
)

type Limit struct {
	Cursor          string
	CursorDirection CursorDirection
	Count           uint64
}

func CursorLimit(cursor string, cursorDirection CursorDirection, count uint64) Limit {
	return Limit{
		Cursor:          cursor,
		CursorDirection: cursorDirection,
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

func (p LimitAndSort) HasCursorLimit() bool {
	return p.Limit.Cursor != "" && p.Limit.CursorDirection != 0
}

func (p LimitAndSort) HasSequenceSort() bool {
	for _, order := range p.SortBy {
		switch order.(type) {
		case SortBySequence:
			return true
		default:
			continue
		}
	}
	return false
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

const MAX_EXPRESSION_DEPTH = 32

func SerializeExpressions(exprs []Expression) ([]Expression, error) {
	serializedExprs := make([]Expression, 0, len(exprs))
	for _, expr := range exprs {
		serializedExpr, err := serializeExpressionValues(expr, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize expression: %w", err)
		}
		serializedExprs = append(serializedExprs, serializedExpr)
	}
	return serializedExprs, nil
}

func DeserializeExpressions(exprs []Expression) ([]Expression, error) {
	deserializedExprs := make([]Expression, 0, len(exprs))
	for _, expr := range exprs {
		deserializedExpr, err := deserializeExpressionValues(expr, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize expression: %w", err)
		}
		deserializedExprs = append(deserializedExprs, deserializedExpr)
	}
	return deserializedExprs, nil
}

// the relay wrapper calls getContractEncodedType which defaults to returning a map[string]any
// https://github.com/smartcontractkit/chainlink-common/blob/fe3ec4466fb5adfffd8fc77eef1cef67c4a918cc/pkg/loop/internal/relayer/pluginprovider/contractreader/contract_reader.go#L1033
// in ccipChainReader.ExecutedMessages, it's a primitive

func serializeExpressionValues(expr Expression, depth int) (Expression, error) {
	if depth > MAX_EXPRESSION_DEPTH {
		return expr, fmt.Errorf("expression depth exceeds maximum allowed depth of %d", MAX_EXPRESSION_DEPTH)
	}

	resultExpr := expr
	var err error

	if expr.Primitive != nil {
		if comp, ok := expr.Primitive.(*primitives.Comparator); ok {
			if comp == nil {
				return expr, fmt.Errorf("invalid Expression: Primitive is *primitives.Comparator but pointer is nil")
			}
			newComp := &primitives.Comparator{
				Name:             comp.Name,
				ValueComparators: make([]primitives.ValueComparator, len(comp.ValueComparators)),
			}
			for i, vc := range comp.ValueComparators {
				var jsonData []byte
				jsonData, err = json.Marshal(vc.Value)
				if err != nil {
					return expr, fmt.Errorf("failed to marshal value in comparator '%s' (type %T): %w", comp.Name, vc.Value, err)
				}
				newComp.ValueComparators[i] = primitives.ValueComparator{
					Value:    &jsonData,
					Operator: vc.Operator,
				}
			}
			resultExpr.Primitive = newComp
		}
	}

	if !isZeroBoolExpression(expr.BoolExpression) {
		originalBoolExpr := expr.BoolExpression
		processedSubExprs := make([]Expression, 0, len(originalBoolExpr.Expressions))

		for _, subExpr := range originalBoolExpr.Expressions {
			var processedSubExpr Expression
			processedSubExpr, err = serializeExpressionValues(subExpr, depth+1)
			if err != nil {
				return expr, fmt.Errorf("failed processing sub-expression within %s: %w", originalBoolExpr.BoolOperator, err)
			}
			processedSubExprs = append(processedSubExprs, processedSubExpr)
		}
		resultExpr.BoolExpression = BoolExpression{
			Expressions:  processedSubExprs,
			BoolOperator: originalBoolExpr.BoolOperator,
		}
	}

	return resultExpr, nil
}

func deserializeExpressionValues(expr Expression, depth int) (Expression, error) {
	if depth > MAX_EXPRESSION_DEPTH {
		return expr, fmt.Errorf("expression depth exceeds maximum allowed depth of %d", MAX_EXPRESSION_DEPTH)
	}

	resultExpr := expr
	var err error

	if expr.Primitive != nil {
		if comp, ok := expr.Primitive.(*primitives.Comparator); ok {
			if comp == nil {
				return expr, fmt.Errorf("invalid Expression: Primitive is *primitives.Comparator but pointer is nil")
			}
			newComp := &primitives.Comparator{
				Name:             comp.Name,
				ValueComparators: make([]primitives.ValueComparator, len(comp.ValueComparators)),
			}
			for i, vc := range comp.ValueComparators {
				if vc.Value == nil {
					newComp.ValueComparators[i] = primitives.ValueComparator{
						Value:    nil,
						Operator: vc.Operator,
					}
					continue
				}

				jsonData, ok := vc.Value.(*[]byte)
				if !ok {
					return expr, fmt.Errorf("failed to deserialize value in comparator '%s': expected []byte, got %T", comp.Name, vc.Value)
				}

				var target uint64
				err = json.Unmarshal(*jsonData, &target)
				if err != nil {
					return expr, fmt.Errorf("failed to unmarshal value '%s' in comparator '%s': %w", string(*jsonData), comp.Name, err)
				}
				newComp.ValueComparators[i] = primitives.ValueComparator{
					Value:    target,
					Operator: vc.Operator,
				}
			}
			resultExpr.Primitive = newComp
		}
	}

	if !isZeroBoolExpression(expr.BoolExpression) {
		originalBoolExpr := expr.BoolExpression
		processedSubExprs := make([]Expression, 0, len(originalBoolExpr.Expressions))

		for _, subExpr := range originalBoolExpr.Expressions {
			var processedSubExpr Expression
			processedSubExpr, err = deserializeExpressionValues(subExpr, depth+1)
			if err != nil {
				return expr, fmt.Errorf("failed processing sub-expression within %s: %w", originalBoolExpr.BoolOperator, err)
			}
			processedSubExprs = append(processedSubExprs, processedSubExpr)
		}
		resultExpr.BoolExpression = BoolExpression{
			Expressions:  processedSubExprs,
			BoolOperator: originalBoolExpr.BoolOperator,
		}
	}

	return resultExpr, nil
}

func isZeroBoolExpression(be BoolExpression) bool {
	return len(be.Expressions) == 0
}
