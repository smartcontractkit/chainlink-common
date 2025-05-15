package evm

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

func ConvertQueryTrackedLogsRequestFromProto(pbQueryTrackedLogsRequest *QueryTrackedLogsRequest) ([]query.Expression, error) {
	var expr []query.Expression
	for _, pbQueryFilter := range pbQueryTrackedLogsRequest.GetExpression() {
		expression, err := convertExpressionFromProto(pbQueryFilter)
		if err != nil {
			return nil, err
		}
		expr = append(expr, expression)
	}

	return expr, nil
}

func convertExpressionFromProto(pbExpression *Expression) (query.Expression, error) {
	switch pbEvaluatedExpr := pbExpression.Evaluator.(type) {
	case *Expression_BooleanExpression:
		var expressions []query.Expression
		for _, expression := range pbEvaluatedExpr.BooleanExpression.Expression {
			convertedExpression, err := convertExpressionFromProto(expression)
			if err != nil {
				return query.Expression{}, err
			}
			expressions = append(expressions, convertedExpression)
		}
		if pbEvaluatedExpr.BooleanExpression.BooleanOperator == pb.BooleanOperator_AND {
			return query.And(expressions...), nil
		}
		return query.Or(expressions...), nil
	case *Expression_Primitive:
		switch primitive := pbEvaluatedExpr.Primitive.GetPrimitive().(type) {
		case *Primitive_EventByTopic:
		case *Primitive_EventByWord:
		case *Primitive_ContractAddress:
		case *Primitive_EventSig:
		case *Primitive_GeneralPrimitive:
			switch primitive.GeneralPrimitive.GetPrimitive().(type) {
			case *pb.Primitive_Confidence:
				confidence, err := ConfidenceFromProto(primitive.GeneralPrimitive.GetConfidence())
				return query.Confidence(confidence), err
			case *pb.Primitive_Block:
				return query.Block(primitive.GeneralPrimitive.GetBlock().GetBlockNumber(), primitives.ComparisonOperator(primitive.GeneralPrimitive.GetBlock().GetOperator())), nil
			case *pb.Primitive_TxHash:
				return query.TxHash(primitive.GeneralPrimitive.GetTxHash().GetTxHash()), nil
			case *pb.Primitive_Timestamp:
				return query.Timestamp(primitive.GeneralPrimitive.GetTimestamp().GetTimestamp(), primitives.ComparisonOperator(primitive.GeneralPrimitive.GetTimestamp().GetOperator())), nil
			}
		default:
			return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type: %T", pbEvaluatedExpr)
	}
}

func ConfidenceFromProto(pbConfidence pb.Confidence) (primitives.ConfidenceLevel, error) {
	switch pbConfidence {
	case pb.Confidence_Finalized:
		return primitives.Finalized, nil
	case pb.Confidence_Unconfirmed:
		return primitives.Unconfirmed, nil
	default:
		return "", fmt.Errorf("invalid pb confidence level: %d", pbConfidence)
	}
}
