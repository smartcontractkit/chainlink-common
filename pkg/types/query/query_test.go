package query

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

func Test_AndOrEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		expressions []Expression
		constructor func(expressions ...Expression) Expression
		expected    Expression
	}{
		{
			name:        "And with no expressions",
			constructor: And,
			expected:    And(),
		},
		{
			name:        "Or with no expressions",
			constructor: Or,
			expected:    Or(),
		},
		{
			name:        "And with one expression",
			expressions: []Expression{TxHash("txHash")},
			constructor: And,
			expected:    TxHash("txHash"),
		},
		{
			name:        "Or with one expression",
			expressions: []Expression{TxHash("txHash")},
			constructor: Or,
			expected:    TxHash("txHash"),
		},
		{
			name:        "And with multiple expressions",
			expressions: []Expression{TxHash("txHash"), Block(123, primitives.Eq)},
			constructor: And,
			expected: And(
				TxHash("txHash"),
				Block(123, primitives.Eq),
			),
		},
		{
			name:        "Or with multiple expressions",
			expressions: []Expression{TxHash("txHash"), Block(123, primitives.Eq)},
			constructor: Or,
			expected: Or(
				TxHash("txHash"),
				Block(123, primitives.Eq),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.constructor(tt.expressions...)
			require.Equal(t, tt.expected, got)
		})
	}
}
