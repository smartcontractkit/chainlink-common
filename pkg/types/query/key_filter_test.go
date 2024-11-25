package query_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

func TestIndexedSequencesKeyFilter(t *testing.T) {
	t.Parallel()

	filter := query.IndexedSequencesKeyFilter("readName", "property", []string{"value1", "value2"}, primitives.Finalized)
	expected := query.KeyFilter{
		Key: "readName",
		Expressions: []query.Expression{
			{BoolExpression: query.BoolExpression{
				Expressions: []query.Expression{
					{
						Primitive: &primitives.Comparator{
							Name:             "property",
							ValueComparators: []primitives.ValueComparator{{Value: "value1", Operator: primitives.Eq}},
						},
					},
					{
						Primitive: &primitives.Comparator{
							Name:             "property",
							ValueComparators: []primitives.ValueComparator{{Value: "value2", Operator: primitives.Eq}},
						},
					},
				},
				BoolOperator: query.OR,
			}},
			{Primitive: &primitives.Confidence{ConfidenceLevel: primitives.Finalized}},
		},
	}

	require.Equal(t, expected, filter)
}

func TestIndexedSequencesByBlockRangeKeyFilter(t *testing.T) {
	t.Parallel()

	filter := query.IndexedSequencesByBlockRangeKeyFilter("readName", "start", "end", "property", []string{"value1", "value2"})
	expected := query.KeyFilter{
		Key: "readName",
		Expressions: []query.Expression{
			{BoolExpression: query.BoolExpression{
				Expressions: []query.Expression{
					{
						Primitive: &primitives.Comparator{
							Name:             "property",
							ValueComparators: []primitives.ValueComparator{{Value: "value1", Operator: primitives.Eq}},
						},
					},
					{
						Primitive: &primitives.Comparator{
							Name:             "property",
							ValueComparators: []primitives.ValueComparator{{Value: "value2", Operator: primitives.Eq}},
						},
					},
				},
				BoolOperator: query.OR,
			}},
			{Primitive: &primitives.Block{Block: "start", Operator: primitives.Gte}},
			{Primitive: &primitives.Block{Block: "end", Operator: primitives.Lte}},
		},
	}

	require.Equal(t, expected, filter)
}

func TestIndexedSequencesValueGreaterThanKeyFilter(t *testing.T) {
	t.Parallel()

	filter := query.IndexedSequencesValueGreaterThanKeyFilter("readName", "property", "value1", primitives.Finalized)
	expected := query.KeyFilter{
		Key: "readName",
		Expressions: []query.Expression{
			{
				Primitive: &primitives.Comparator{
					Name:             "property",
					ValueComparators: []primitives.ValueComparator{{Value: "value1", Operator: primitives.Gte}},
				},
			},
			{Primitive: &primitives.Confidence{ConfidenceLevel: primitives.Finalized}},
		},
	}

	require.Equal(t, expected, filter)
}

func TestIndexedSequencesValueRangeKeyFilter(t *testing.T) {
	t.Parallel()

	filter := query.IndexedSequencesValueRangeKeyFilter("readName", "property", "min", "max", primitives.Finalized)
	expected := query.KeyFilter{
		Key: "readName",
		Expressions: []query.Expression{
			{
				Primitive: &primitives.Comparator{
					Name:             "property",
					ValueComparators: []primitives.ValueComparator{{Value: "min", Operator: primitives.Gte}},
				},
			},
			{
				Primitive: &primitives.Comparator{
					Name:             "property",
					ValueComparators: []primitives.ValueComparator{{Value: "max", Operator: primitives.Lte}},
				},
			},
			{Primitive: &primitives.Confidence{ConfidenceLevel: primitives.Finalized}},
		},
	}

	require.Equal(t, expected, filter)
}

func TestIndexedSequencesByTxHashKeyFilter(t *testing.T) {
	t.Parallel()

	filter := query.IndexedSequencesByTxHashKeyFilter("readName", "hash")
	expected := query.KeyFilter{
		Key: "readName",
		Expressions: []query.Expression{
			{Primitive: &primitives.TxHash{TxHash: "hash"}},
		},
	}

	require.Equal(t, expected, filter)
}

func TestSequencesByBlockRangeKeyFilter(t *testing.T) {
	t.Parallel()

	filter := query.SequencesByBlockRangeKeyFilter("readName", "start", "end")
	expected := query.KeyFilter{
		Key: "readName",
		Expressions: []query.Expression{
			{Primitive: &primitives.Block{Block: "start", Operator: primitives.Gte}},
			{Primitive: &primitives.Block{Block: "end", Operator: primitives.Lte}},
		},
	}

	require.Equal(t, expected, filter)
}

func TestSequencesCreatedAfterKeyFilter(t *testing.T) {
	t.Parallel()

	now := time.Now()

	filter := query.SequencesCreatedAfterKeyFilter("readName", now, primitives.Finalized)
	expected := query.KeyFilter{
		Key: "readName",
		Expressions: []query.Expression{
			{Primitive: &primitives.Timestamp{Timestamp: uint64(now.Unix()), Operator: primitives.Gt}},
			{Primitive: &primitives.Confidence{ConfidenceLevel: primitives.Finalized}},
		},
	}

	require.Equal(t, expected, filter)
}

func TestIndexedSequencesCreatedAfterKeyFilter(t *testing.T) {
	t.Parallel()

	now := time.Now()

	filter := query.IndexedSequencesCreatedAfterKeyFilter("readName", "property", []string{"value1", "value2"}, now, primitives.Finalized)
	expected := query.KeyFilter{
		Key: "readName",
		Expressions: []query.Expression{
			{BoolExpression: query.BoolExpression{
				Expressions: []query.Expression{
					{
						Primitive: &primitives.Comparator{
							Name:             "property",
							ValueComparators: []primitives.ValueComparator{{Value: "value1", Operator: primitives.Eq}},
						},
					},
					{
						Primitive: &primitives.Comparator{
							Name:             "property",
							ValueComparators: []primitives.ValueComparator{{Value: "value2", Operator: primitives.Eq}},
						},
					},
				},
				BoolOperator: query.OR,
			}},
			{Primitive: &primitives.Timestamp{Timestamp: uint64(now.Unix()), Operator: primitives.Gt}},
			{Primitive: &primitives.Confidence{ConfidenceLevel: primitives.Finalized}},
		},
	}

	require.Equal(t, expected, filter)
}
