package solana

import (
	"fmt"

	chainsolana "github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	typesolana "github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

// ConvertComparisonOperatorFromProto converts a proto ComparisonOperator to primitives.ComparisonOperator
func ConvertComparisonOperatorFromProto(op ComparisonOperator) (primitives.ComparisonOperator, error) {
	switch op {
	case ComparisonOperator_COMPARISON_OPERATOR_EQ:
		return primitives.Eq, nil
	case ComparisonOperator_COMPARISON_OPERATOR_NEQ:
		return primitives.Neq, nil
	case ComparisonOperator_COMPARISON_OPERATOR_GT:
		return primitives.Gt, nil
	case ComparisonOperator_COMPARISON_OPERATOR_LT:
		return primitives.Lt, nil
	case ComparisonOperator_COMPARISON_OPERATOR_GTE:
		return primitives.Gte, nil
	case ComparisonOperator_COMPARISON_OPERATOR_LTE:
		return primitives.Lte, nil
	default:
		return 0, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

// ConvertComparisonOperatorToProto converts a primitives.ComparisonOperator to proto ComparisonOperator
func ConvertComparisonOperatorToProto(op primitives.ComparisonOperator) (ComparisonOperator, error) {
	switch op {
	case primitives.Eq:
		return ComparisonOperator_COMPARISON_OPERATOR_EQ, nil
	case primitives.Neq:
		return ComparisonOperator_COMPARISON_OPERATOR_NEQ, nil
	case primitives.Gt:
		return ComparisonOperator_COMPARISON_OPERATOR_GT, nil
	case primitives.Lt:
		return ComparisonOperator_COMPARISON_OPERATOR_LT, nil
	case primitives.Gte:
		return ComparisonOperator_COMPARISON_OPERATOR_GTE, nil
	case primitives.Lte:
		return ComparisonOperator_COMPARISON_OPERATOR_LTE, nil
	default:
		return 0, fmt.Errorf("unknown comparison operator: %s", op)
	}
}

// ConvertValueComparatorsFromProto converts proto ValueComparator slice to primitives.ValueComparator slice
func ConvertValueComparatorsFromProto(comparers []*ValueComparator) ([]primitives.ValueComparator, error) {
	if comparers == nil {
		return nil, nil
	}
	result := make([]primitives.ValueComparator, len(comparers))
	for i, c := range comparers {
		if c != nil {
			operator, err := ConvertComparisonOperatorFromProto(c.Operator)
			if err != nil {
				return nil, fmt.Errorf("failed to convert comparison operator: %w", err)
			}
			result[i] = primitives.ValueComparator{
				Value:    c.Value, // []byte is compatible with any
				Operator: operator,
			}
		}
	}
	return result, nil
}

// ConvertValueComparatorsToProto converts primitives.ValueComparator slice to proto ValueComparator slice
func ConvertValueComparatorsToProto(comparers []primitives.ValueComparator) ([]*ValueComparator, error) {
	if comparers == nil {
		return nil, nil
	}
	result := make([]*ValueComparator, len(comparers))
	for i, c := range comparers {
		// Handle the Value field which could be any type, convert to []byte if possible
		var valueBytes []byte
		if b, ok := c.Value.([]byte); ok {
			valueBytes = b
		} else {
			return nil, fmt.Errorf("value is not a []byte: %T", c.Value)
		}
		operator, err := ConvertComparisonOperatorToProto(c.Operator)
		if err != nil {
			return nil, fmt.Errorf("failed to convert comparison operator: %w", err)
		}
		result[i] = &ValueComparator{
			Value:    valueBytes,
			Operator: operator,
		}
	}
	return result, nil
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
