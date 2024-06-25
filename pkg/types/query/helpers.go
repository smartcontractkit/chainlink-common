package query

import (
	"strconv"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type ComparatorType string

const (
	AddressComparator  ComparatorType = "ADDRESS"
	EventSigComparator ComparatorType = "EVENT"
	TopicComparator    ComparatorType = "TOPIC"
)

// SelectIndexedLogs creates a KeyFilter that filters logs for the provided topic values at the specified
// topic index. Topic value filters are 'OR'ed together.
func SelectIndexedLogs(
	address, eventSig string,
	topicIdx uint64,
	topicValues []string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: strconv.FormatUint(topicIdx, 10),
		Expressions: []Expression{
			addressComparator(address),
			eventComparator(eventSig),
			filtersForTopics(topicValues),
			Confidence(confidence),
		},
	}
}

// SelectIndexedLogsByBlockRange creates a KeyFilter that filters logs for the provided topic values at the specified
// topic index. Topic value filters are 'OR'ed together and results are limited by provided block range.
func SelectIndexedLogsByBlockRange(
	addr, eventSig string,
	start, end uint64,
	topicIdx uint64,
	topicValues []string,
) KeyFilter {
	return KeyFilter{
		Key: strconv.FormatUint(topicIdx, 10),
		Expressions: []Expression{
			addressComparator(addr),
			eventComparator(eventSig),
			filtersForTopics(topicValues),
			Block(start, primitives.Gte),
			Block(end, primitives.Lte),
		},
	}
}

// SelectIndexedLogsTopicGreaterThan creates a KeyFilter that filters logs for the provided topic value and index
// at or above the specified confidence level.
func SelectIndexedLogsTopicGreaterThan(
	addr, eventSig string,
	topicIdx uint64, topicValue string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: strconv.FormatUint(topicIdx, 10),
		Expressions: []Expression{
			addressComparator(addr),
			eventComparator(eventSig),
			topicComparator(topicValue, primitives.Gte),
			Confidence(confidence),
		},
	}
}

// SelectIndexedLogsTopicRange creates a KeyFilter that filters logs for the provided topic index and topic
// values between the provided min and max, endpoints inclusive.
func SelectIndexedLogsTopicRange(
	addr, eventSig string,
	topicIdx uint64, min, max string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Key: strconv.FormatUint(topicIdx, 10),
		Expressions: []Expression{
			addressComparator(addr),
			eventComparator(eventSig),
			topicComparator(min, primitives.Gte),
			topicComparator(max, primitives.Lte),
			Confidence(confidence),
		},
	}
}

// SelectIndexedLogsByTxHash creates a KeyFilter that filters logs for the provided transaction hash.
func SelectIndexedLogsByTxHash(
	addr, eventSig, txHash string,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(addr),
			eventComparator(eventSig),
			TxHash(txHash),
		},
	}
}

// SelectLogsDataWordRange creates a KeyFilter that filters logs for the provided word index and word
// values between the provided min and max, endpoints inclusive.
func SelectLogsDataWordRange(
	addr, eventSig string,
	wordIdx uint8, word1, word2 string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(addr),
			eventComparator(eventSig),
			wordComparator(
				strconv.FormatUint(uint64(wordIdx), 10),
				word1,
				primitives.Gte,
			),
			wordComparator(
				strconv.FormatUint(uint64(wordIdx), 10),
				word2,
				primitives.Lte,
			),
			Confidence(confidence),
		},
	}
}

// SelectLogsDataWordGreaterThan creates a KeyFilter that filters logs for the provided word index and
// greater than or equal to the provided word value.
func SelectLogsDataWordGreaterThan(
	addr, eventSig string,
	wordIdx uint8, wordValue string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(addr),
			eventComparator(eventSig),
			wordComparator(
				strconv.FormatUint(uint64(wordIdx), 10),
				wordValue,
				primitives.Gte,
			),
			Confidence(confidence),
		},
	}
}

// SelectLogsWithSigs creates a KeyFilter that filters logs for the provided event signatures within
// the provided block range. Event signature values are 'OR'ed and block range endpoints are inclusive.
func SelectLogsWithSigs(
	addr string, sigs []string,
	startBlock, endBlock uint64,
) KeyFilter {
	filters := []Expression{
		addressComparator(addr),
	}

	if len(sigs) > 0 {
		exp := make([]Expression, len(sigs))
		for idx, val := range sigs {
			exp[idx] = eventComparator(val)
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
				Block(startBlock, primitives.Gte),
				Block(endBlock, primitives.Lte),
			},
			BoolOperator: AND,
		},
	})

	return KeyFilter{
		Expressions: filters,
	}
}

// SelectLogs creates a KeyFilter that filters logs for the provided block range, endpoints inclusive.
func SelectLogs(
	addr, eventSig string,
	start, end uint64,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(addr),
			eventComparator(eventSig),
			Block(start, primitives.Gte),
			Block(end, primitives.Lte),
		},
	}
}

// SelectLogsCreatedAfter creates a KeyFilter that filters logs for after but not equal to the provided time value.
func SelectLogsCreatedAfter(
	address, eventSig string,
	timestamp time.Time,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(address),
			eventComparator(eventSig),
			Timestamp(uint64(timestamp.Unix()), primitives.Gt),
			Confidence(confidence),
		},
	}
}

// SelectIndexedLogsCreatedAfter creates a KeyFilter that filters logs for the provided topic index and topic values
// created after the provided time value. Topic values are 'OR'ed.
func SelectIndexedLogsCreatedAfter(
	address, eventSig string,
	topicIdx uint64,
	topicValues []string,
	timestamp time.Time,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	filters := []Expression{
		addressComparator(address),
		eventComparator(eventSig),
	}

	if len(topicValues) > 0 {
		exp := make([]Expression, len(topicValues))
		for idx, value := range topicValues {
			exp[idx] = topicComparator(value, primitives.Eq)
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
		Key:         strconv.FormatUint(topicIdx, 10),
		Expressions: filters,
	}
}

// SelectLogsDataWordBetween creates a KeyFilter that filters logs between the specified word indexes and
// provided word value, endpoints inclusive.
func SelectLogsDataWordBetween(
	address, eventSig string,
	wordIdx1, wordIdx2 uint64, word string,
	confidence primitives.ConfidenceLevel,
) KeyFilter {
	return KeyFilter{
		Expressions: []Expression{
			addressComparator(address),
			eventComparator(eventSig),
			wordComparator(
				strconv.FormatUint(wordIdx1, 10),
				word,
				primitives.Lte,
			),
			wordComparator(
				strconv.FormatUint(wordIdx2, 10),
				word,
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

func eventComparator(eventSig string) Expression {
	return Comparator(string(EventSigComparator), primitives.ValueComparator{
		Value:    eventSig,
		Operator: primitives.Eq,
	})
}

func topicComparator(topic string, op primitives.ComparisonOperator) Expression {
	return Comparator(string(TopicComparator), primitives.ValueComparator{
		Value:    topic,
		Operator: op,
	})
}

func wordComparator(wordName, wordValue string, op primitives.ComparisonOperator) Expression {
	return Comparator(wordName, primitives.ValueComparator{
		Value:    wordValue,
		Operator: op,
	})
}

func filtersForTopics(topicValues []string) Expression {
	topicFilters := BoolExpression{
		Expressions:  make([]Expression, len(topicValues)),
		BoolOperator: OR,
	}

	for idx, value := range topicValues {
		topicFilters.Expressions[idx] = topicComparator(value, primitives.Eq)
	}

	return Expression{BoolExpression: topicFilters}
}
