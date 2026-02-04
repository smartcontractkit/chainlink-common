package solana

import (
	"fmt"

	chainsolana "github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	typesolana "github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

// ConvertComparisonOperatorFromProto converts a proto ComparisonOperator to primitives.ComparisonOperator
func ConvertComparisonOperatorFromProto(op ComparisonOperator) primitives.ComparisonOperator {
	switch op {
	case ComparisonOperator_COMPARISON_OPERATOR_EQ:
		return primitives.Eq
	case ComparisonOperator_COMPARISON_OPERATOR_NEQ:
		return primitives.Neq
	case ComparisonOperator_COMPARISON_OPERATOR_GT:
		return primitives.Gt
	case ComparisonOperator_COMPARISON_OPERATOR_LT:
		return primitives.Lt
	case ComparisonOperator_COMPARISON_OPERATOR_GTE:
		return primitives.Gte
	case ComparisonOperator_COMPARISON_OPERATOR_LTE:
		return primitives.Lte
	default:
		return primitives.Eq
	}
}

// ConvertComparisonOperatorToProto converts a primitives.ComparisonOperator to proto ComparisonOperator
func ConvertComparisonOperatorToProto(op primitives.ComparisonOperator) ComparisonOperator {
	switch op {
	case primitives.Eq:
		return ComparisonOperator_COMPARISON_OPERATOR_EQ
	case primitives.Neq:
		return ComparisonOperator_COMPARISON_OPERATOR_NEQ
	case primitives.Gt:
		return ComparisonOperator_COMPARISON_OPERATOR_GT
	case primitives.Lt:
		return ComparisonOperator_COMPARISON_OPERATOR_LT
	case primitives.Gte:
		return ComparisonOperator_COMPARISON_OPERATOR_GTE
	case primitives.Lte:
		return ComparisonOperator_COMPARISON_OPERATOR_LTE
	default:
		return ComparisonOperator_COMPARISON_OPERATOR_EQ
	}
}

// ConvertValueComparatorsFromProto converts proto ValueComparator slice to primitives.ValueComparator slice
func ConvertValueComparatorsFromProto(comparers []*ValueComparator) []primitives.ValueComparator {
	if comparers == nil {
		return nil
	}
	result := make([]primitives.ValueComparator, len(comparers))
	for i, c := range comparers {
		if c != nil {
			result[i] = primitives.ValueComparator{
				Value:    c.Value, // []byte is compatible with any
				Operator: ConvertComparisonOperatorFromProto(c.Operator),
			}
		}
	}
	return result
}

// ConvertValueComparatorsToProto converts primitives.ValueComparator slice to proto ValueComparator slice
func ConvertValueComparatorsToProto(comparers []primitives.ValueComparator) []*ValueComparator {
	if comparers == nil {
		return nil
	}
	result := make([]*ValueComparator, len(comparers))
	for i, c := range comparers {
		// Handle the Value field which could be any type, convert to []byte if possible
		var valueBytes []byte
		if b, ok := c.Value.([]byte); ok {
			valueBytes = b
		}
		result[i] = &ValueComparator{
			Value:    valueBytes,
			Operator: ConvertComparisonOperatorToProto(c.Operator),
		}
	}
	return result
}

// ConvertLogFromProto converts a proto Log to typesolana.Log
func ConvertLogFromProto(p *Log) (*typesolana.Log, error) {
	if p == nil {
		return nil, nil
	}

	blockHash, err := chainsolana.ConvertHashFromProto(p.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to convert block hash: %w", err)
	}

	address, err := chainsolana.ConvertPublicKeyFromProto(p.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to convert address: %w", err)
	}

	eventSig, err := chainsolana.ConvertEventSigFromProto(p.EventSig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert event sig: %w", err)
	}

	txHash, err := chainsolana.ConvertSignatureFromProto(p.TxHash)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tx hash: %w", err)
	}

	return &typesolana.Log{
		ChainID:        p.ChainId,
		LogIndex:       p.LogIndex,
		BlockHash:      blockHash,
		BlockNumber:    p.BlockNumber,
		BlockTimestamp: p.BlockTimestamp,
		Address:        address,
		EventSig:       eventSig,
		TxHash:         txHash,
		Data:           p.Data,
		SequenceNum:    p.SequenceNum,
		Error:          p.Error,
	}, nil
}

// ConvertLogToProto converts a typesolana.Log to proto Log
func ConvertLogToProto(l *typesolana.Log) *Log {
	if l == nil {
		return nil
	}
	return &Log{
		ChainId:        l.ChainID,
		LogIndex:       l.LogIndex,
		BlockHash:      l.BlockHash[:],
		BlockNumber:    l.BlockNumber,
		BlockTimestamp: l.BlockTimestamp,
		Address:        l.Address[:],
		EventSig:       l.EventSig[:],
		TxHash:         l.TxHash[:],
		Data:           l.Data,
		SequenceNum:    l.SequenceNum,
		Error:          l.Error,
	}
}
