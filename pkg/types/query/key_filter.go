package query

import (
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

// IndexedSequencesKeyFilter creates a KeyFilter that filters logs for the provided sequence property values at the
// specified property name. Sequence value filters are 'OR'ed together. A sequence read name is the value that
// identifies the sequence type. The signature value name is the sequence property to apply the filter to and the
// sequence values are the individual values to search for in the provided property.
func IndexedSequencesKeyFilter(
	readName string,
	comparatorName string,
	values []string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: readName,
		Expressions: []Expression{
			filtersForValues(comparatorName, values),
			Confidence(confidence),
		},
	}
}

// IndexedSequencesByBlockRangeKeyFilter creates a KeyFilter that filters sequences for the provided property values at
// the specified property name. Value filters are 'OR'ed together and results are limited by provided cursor range. A
// read name is the value that identifies the sequence type. The signature property name is the sequence property to
// apply the filter to and the sequence values are the individual values to search for in the provided property.
func IndexedSequencesByBlockRangeKeyFilter(
	readName string,
	start, end string,
	comparatorName string,
	values []string,
) KeyFilter {
	return KeyFilter{
		Key: readName,
		Expressions: []Expression{
			filtersForValues(comparatorName, values),
			Block(start, primitives.Gte),
			Block(end, primitives.Lte),
		},
	}
}

// IndexedSequencesValueGreaterThanKeyFilter creates a KeyFilter that filters sequences for the provided property value
// and name at or above the specified confidence level. A sequence read name is the value that identifies the sequence
// type. The property name is the sequence property to apply the filter to and the value is the individual value to
// search for in the provided property.
func IndexedSequencesValueGreaterThanKeyFilter(
	readName string,
	comparatorName, value string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: readName,
		Expressions: []Expression{
			valueComparator(comparatorName, value, primitives.Gte),
			Confidence(confidence),
		},
	}
}

// IndexedSequencesValueRangeKeyFilter creates a KeyFilter that filters logs on the provided sequence property between
// the provided min and max, endpoints inclusive. A sequence read name is the value that identifies the sequence type.
func IndexedSequencesValueRangeKeyFilter(
	readName string,
	comparatorName string,
	min, max string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: readName,
		Expressions: []Expression{
			valueComparator(comparatorName, min, primitives.Gte),
			valueComparator(comparatorName, max, primitives.Lte),
			Confidence(confidence),
		},
	}
}

// IndexedSequencesByTxHashKeyFilter creates a KeyFilter that filters logs for the provided transaction hash. A sequence
// read name is the value that identifies the sequence type.
func IndexedSequencesByTxHashKeyFilter(
	readName, txHash string,
) KeyFilter {
	return KeyFilter{
		Key: readName,
		Expressions: []Expression{
			TxHash(txHash),
		},
	}
}

// SequencesByBlockRangeKeyFilter creates a KeyFilter that filters sequences for the provided block range, endpoints inclusive.
func SequencesByBlockRangeKeyFilter(
	readName string,
	start, end string,
) KeyFilter {
	return KeyFilter{
		Key: readName,
		Expressions: []Expression{
			Block(start, primitives.Gte),
			Block(end, primitives.Lte),
		},
	}
}

// SequencesCreatedAfterKeyFilter creates a KeyFilter that filters sequences for after but not equal to the provided time value.
func SequencesCreatedAfterKeyFilter(
	readName string,
	timestamp time.Time,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: readName,
		Expressions: []Expression{
			Timestamp(uint64(timestamp.Unix()), primitives.Gt),
			Confidence(confidence),
		},
	}
}

// IndexedSequencesCreatedAfterKeyFilter creates a KeyFilter that filters sequences for the provided property and values
// created after the provided time value. Sequence property values filters are 'OR'ed. A sequence read name is the value
// that identifies the sequence type.
func IndexedSequencesCreatedAfterKeyFilter(
	readName string,
	comparatorName string,
	values []string,
	timestamp time.Time,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: readName,
		Expressions: []Expression{
			filtersForValues(comparatorName, values),
			Timestamp(uint64(timestamp.Unix()), primitives.Gt),
			Confidence(confidence),
		},
	}
}

func valueComparator(comparatorName, value string, op primitives.ComparisonOperator) Expression {
	return Comparator(comparatorName, primitives.ValueComparator{
		Value:    value,
		Operator: op,
	})
}

func filtersForValues(comparatorName string, values []string) Expression {
	valueFilters := BoolExpression{
		Expressions:  make([]Expression, len(values)),
		BoolOperator: OR,
	}

	for idx, value := range values {
		valueFilters.Expressions[idx] = valueComparator(comparatorName, value, primitives.Eq)
	}

	return Expression{BoolExpression: valueFilters}
}
