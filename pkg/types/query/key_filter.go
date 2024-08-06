package query

import (
	"strconv"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type ComparatorType string

const (
	AddressComparator   ComparatorType = "ADDRESS"
	SignatureComparator ComparatorType = "SIGNATURE"
	ValueComparator     ComparatorType = "VALUE"
)

// IndexedSequencesKeyFilter creates a KeyFilter that filters logs for the provided sequence values at the specified
// sequence property name. Sequence value filters are 'OR'ed together. A sequence signature is the value that identifies the
// sequence type. The signature value name is the sequence property to apply the filter to and the sequence
// values are the individual values to search for in the provided property.
func IndexedSequencesKeyFilter(
	contractAddress, sequenceSig string, // sequence signature
	sequenceValueName string, // sequence property to filter on
	sequenceValues []string, // sequence property values
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: sequenceValueName,
		Expressions: []Expression{
			addressComparator(contractAddress),
			sigComparator(sequenceSig),
			filtersForValues(sequenceValues),
			Confidence(confidence),
		},
	}
}

// IndexedSequencesByBlockRangeKeyFilter creates a KeyFilter that filters sequences for the provided property values at the specified
// property name. Value filters are 'OR'ed together and results are limited by provided cursor range. A sequence signature is the
// value that identifies the sequence type. The signature value name is the sequence property to apply the filter to and the sequence
// values are the individual values to search for in the provided property.
func IndexedSequencesByBlockRangeKeyFilter(
	contractAddress, sequenceSig string,
	start, end uint64,
	sequenceValueName string,
	sequenceValues []string,
) KeyFilter {
	return KeyFilter{
		Key: sequenceValueName,
		Expressions: []Expression{
			addressComparator(contractAddress),
			sigComparator(sequenceSig),
			filtersForValues(sequenceValues),
			Block(start, primitives.Gte),
			Block(end, primitives.Lte),
		},
	}
}

// IndexedSequencesValueGreaterThanKeyFilter creates a KeyFilter that filters sequences for the provided property value and name
// at or above the specified confidence level. A sequence signature is the value that identifies the sequence type. The signature
// value name is the sequence property to apply the filter to and the sequence value is the individual value to search for in
// the provided property.
func IndexedSequencesValueGreaterThanKeyFilter(
	contractAddress, sequenceSig string,
	sequenceValueName, sequenceValue string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: sequenceValueName,
		Expressions: []Expression{
			addressComparator(contractAddress),
			sigComparator(sequenceSig),
			valueComparator(sequenceValue, primitives.Gte),
			Confidence(confidence),
		},
	}
}

// IndexedSequencesValueRangeKeyFilter creates a KeyFilter that filters logs for the provided sequence value name and values
// between the provided min and max, endpoints inclusive. A sequence signature is the value that identifies the sequence type.
// The signature value name is the sequence property to apply the filter to and the sequence value is the individual value to
// search for in the provided property.

func IndexedSequencesValueRangeKeyFilter(
	contractAddress, sequenceSig string,
	sequenceValueName string,
	min, max string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: sequenceValueName,
		Expressions: []Expression{
			addressComparator(contractAddress),
			sigComparator(sequenceSig),
			valueComparator(min, primitives.Gte),
			valueComparator(max, primitives.Lte),
			Confidence(confidence),
		},
	}
}

// IndexedSequencesByTxHashKeyFilter creates a KeyFilter that filters logs for the provided transaction hash. A sequence
// signature is the value that identifies the sequence type.
func IndexedSequencesByTxHashKeyFilter(
	contractAddress, sequenceSig, txHash string,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(contractAddress),
			sigComparator(sequenceSig),
			TxHash(txHash),
		},
	}
}

// LogsDataOffsetRangeKeyFilter creates a KeyFilter that applies filters on sequence raw data for a specified sequence
// signature. Values are compared between the provided min and max values, endpoints inclusive. The dataOffsetIdx is
// the raw data index to apply value filtering on.
//
// TODO: this type of filter may introduce chain dependencies due to difference in raw data structure across multiple chains
func LogsDataOffsetRangeKeyFilter(
	contractAddress, sequenceSig string,
	dataOffsetIdx uint8, from, to string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(contractAddress),
			sigComparator(sequenceSig),
			wordComparator(
				strconv.FormatUint(uint64(dataOffsetIdx), 10),
				from,
				primitives.Gte,
			),
			wordComparator(
				strconv.FormatUint(uint64(dataOffsetIdx), 10),
				to,
				primitives.Lte,
			),
			Confidence(confidence),
		},
	}
}

// LogsDataOffsetGreaterThanKeyFilter creates a KeyFilter that filters sequences for the provided offset index and greater
// than or equal to the provided data value.
//
// TODO: this type of filter may introduce chain dependencies due to difference in raw data structure across multiple chains
func LogsDataOffsetGreaterThanKeyFilter(
	contractAddress, sequenceSig string,
	dataOffsetIdx uint8, value string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(contractAddress),
			sigComparator(sequenceSig),
			wordComparator(
				strconv.FormatUint(uint64(dataOffsetIdx), 10),
				value,
				primitives.Gte,
			),
			Confidence(confidence),
		},
	}
}

// SequencesWithSigsKeyFilter creates a KeyFilter that filters sequences for the provided signatures within the provided
// block range. Sequence signature values are 'OR'ed and block range endpoints are inclusive. A sequence signature is
// the value that identifies the sequence type.
func SequencesWithSigsKeyFilter(
	contractAddress string, sequenceSigs []string,
	start, end uint64,
) KeyFilter {
	filters := []Expression{
		addressComparator(contractAddress),
	}

	if len(sequenceSigs) > 0 {
		exp := make([]Expression, len(sequenceSigs))
		for idx, val := range sequenceSigs {
			exp[idx] = sigComparator(val)
		}

		filters = append(filters, Expression{
			BoolExpression: BoolExpression{
				Expressions:  exp,
				BoolOperator: OR,
			},
		})
	}

	filters = append(filters, Expression{
		BoolExpression: BoolExpression{
			Expressions: []Expression{
				Block(start, primitives.Gte),
				Block(end, primitives.Lte),
			},
			BoolOperator: AND,
		},
	})

	return KeyFilter{
		Expressions: filters,
	}
}

// SequencesByBlockRangeKeyFilter creates a KeyFilter that filters sequences for the provided block range, endpoints inclusive.
func SequencesByBlockRangeKeyFilter(
	contractAddress, sequenceSig string,
	start, end uint64,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(contractAddress),
			sigComparator(sequenceSig),
			Block(start, primitives.Gte),
			Block(end, primitives.Lte),
		},
	}
}

// SequencesCreatedAfterKeyFilter creates a KeyFilter that filters sequences for after but not equal to the provided time value.
func SequencesCreatedAfterKeyFilter(
	contractAddress, sequenceSig string,
	timestamp time.Time,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(contractAddress),
			sigComparator(sequenceSig),
			Timestamp(uint64(timestamp.Unix()), primitives.Gt),
			Confidence(confidence),
		},
	}
}

// IndexedSequencesCreatedAfterKeyFilter creates a KeyFilter that filters sequences for the provided sequence property and values
// created after the provided time value. Sequence property values filters are 'OR'ed. A sequence signature is the value that
// identifies the sequence type. The signature value name is the sequence property to apply the filter to and the sequence
// values are the individual values to search for in the provided property.
func IndexedSequencesCreatedAfterKeyFilter(
	contractAddress, sequenceSig string,
	sequenceValueName string,
	sequenceValues []string,
	timestamp time.Time,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	filters := []Expression{
		addressComparator(contractAddress),
		sigComparator(sequenceSig),
	}

	if len(sequenceValues) > 0 {
		exp := make([]Expression, len(sequenceValues))
		for idx, value := range sequenceValues {
			exp[idx] = valueComparator(value, primitives.Eq)
		}

		filters = append(filters, Expression{
			BoolExpression: BoolExpression{
				Expressions:  exp,
				BoolOperator: OR,
			},
		})
	}

	filters = append(filters, []Expression{
		Timestamp(uint64(timestamp.Unix()), primitives.Gt),
		Confidence(confidence),
	}...)

	return KeyFilter{
		Key:         sequenceValueName,
		Expressions: filters,
	}
}

// SequencesDataOffsetBetweenKeyFilter creates a KeyFilter that filters sequences between the specified offset indexes and
// provided word value, endpoints inclusive. The specified value must be greater than or equal to the lower offset index
// and less than or equal to the upper offset index.
//
// TODO: this type of filter may introduce chain dependencies due to difference in raw data structure across multiple chains
func SequencesDataOffsetBetweenKeyFilter(
	contractAddress, sequenceSig string,
	lowerOffsetIdx, upperOffsetIdx uint64, value string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(contractAddress),
			sigComparator(sequenceSig),
			wordComparator(
				strconv.FormatUint(lowerOffsetIdx, 10),
				value,
				primitives.Lte,
			),
			wordComparator(
				strconv.FormatUint(upperOffsetIdx, 10),
				value,
				primitives.Gte,
			),
			Confidence(confidence),
		},
	}
}

func addressComparator(address string) Expression {
	return Comparator(string(AddressComparator), primitives.ValueComparator{
		Value:    address,
		Operator: primitives.Eq,
	})
}

func sigComparator(eventSig string) Expression {
	return Comparator(string(SignatureComparator), primitives.ValueComparator{
		Value:    eventSig,
		Operator: primitives.Eq,
	})
}

func valueComparator(value string, op primitives.ComparisonOperator) Expression {
	return Comparator(string(ValueComparator), primitives.ValueComparator{
		Value:    value,
		Operator: op,
	})
}

func wordComparator(wordName, wordValue string, op primitives.ComparisonOperator) Expression {
	return Comparator(wordName, primitives.ValueComparator{
		Value:    wordValue,
		Operator: op,
	})
}

func filtersForValues(values []string) Expression {
	valueFilters := BoolExpression{
		Expressions:  make([]Expression, len(values)),
		BoolOperator: OR,
	}

	for idx, value := range values {
		valueFilters.Expressions[idx] = valueComparator(value, primitives.Eq)
	}

	return Expression{BoolExpression: valueFilters}
}
