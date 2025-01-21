package primitives

import "fmt"

// Visitor should have a per chain per db type implementation that converts primitives to db queries.
type Visitor interface {
	Comparator(primitive Comparator)
	Block(primitive Block)
	Confidence(primitive Confidence)
	Timestamp(primitive Timestamp)
	TxHash(primitive TxHash)
}

// Primitive is the basic building block for KeyFilter.
type Primitive interface {
	Accept(visitor Visitor)
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

func (cmpOp ComparisonOperator) String() string {
	switch cmpOp {
	case Eq:
		return "=="
	case Neq:
		return "!="
	case Gt:
		return ">"
	case Lt:
		return "<"
	case Gte:
		return ">="
	case Lte:
		return "<="
	default:
		return "Unknown"
	}
}

type ValueComparator struct {
	Value    any
	Operator ComparisonOperator
}

// AnyOperator - represents SQL's `ANY (array expression)` operator. Useful to replace multiple comparators joined with OR.
// i.e. allows to replace `field1 >= v1 OR field1 >= v2 OR field1 >= v3` with `field1 >= ANY(v1, v2, v3)`
type AnyOperator []any

func Any[T any](slice []T) AnyOperator {
	result := make([]any, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}

// Comparator is used to filter over values that belong to key data.
type Comparator struct {
	Name             string
	ValueComparators []ValueComparator
}

func (f *Comparator) Accept(visitor Visitor) {
	visitor.Comparator(*f)
}

// Block is a primitive of KeyFilter that filters search in comparison to block number.
type Block struct {
	Block    string
	Operator ComparisonOperator
}

func (f *Block) Accept(visitor Visitor) {
	visitor.Block(*f)
}

type ConfidenceLevel string

const (
	Finalized   ConfidenceLevel = "finalized"
	Unconfirmed ConfidenceLevel = "unconfirmed"
)

// Confidence is a primitive of KeyFilter that filters search to results that have a certain level of finalization.
// Confidence maps to different concepts on different blockchains.
type Confidence struct {
	ConfidenceLevel
}

func ConfidenceLevelFromString(value string) (ConfidenceLevel, error) {
	switch value {
	case "finalized":
		return Finalized, nil
	case "unconfirmed":
		return Unconfirmed, nil
	default:
		return "", fmt.Errorf("invalid ConfidenceLevel: %s", value)
	}
}

func (f *Confidence) Accept(visitor Visitor) {
	visitor.Confidence(*f)
}

// Timestamp is a primitive of KeyFilter that filters search in comparison to timestamp.
type Timestamp struct {
	Timestamp uint64
	Operator  ComparisonOperator
}

func (f *Timestamp) Accept(visitor Visitor) {
	visitor.Timestamp(*f)
}

// TxHash is a primitive of KeyFilter that filters search to results that contain txHash.
type TxHash struct {
	TxHash string
}

func (f *TxHash) Accept(visitor Visitor) {
	visitor.TxHash(*f)
}
