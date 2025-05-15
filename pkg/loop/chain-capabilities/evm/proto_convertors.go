package evm

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/evm"
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
			return evm.NewEventByTopicFilter(
				primitive.EventByTopic.GetTopic(),
				convertHashedValueComparatorsFromProto(primitive.EventByTopic.HashedValueComparers)), nil
		case *Primitive_EventByWord:
			return evm.NewEventByWordFilter(int(primitive.EventByWord.GetWordIndex()), convertHashedValueComparatorsFromProto(primitive.EventByWord.HashedValueComparers)), nil
		case *Primitive_ContractAddress:
			return evm.NewAddressFilter(evmtypes.Address(primitive.ContractAddress.GetAddress().Address)), nil
		case *Primitive_EventSig:
			return evm.NewEventSigFilter(evmtypes.Hash(primitive.EventSig.GetEventSig().Hash)), nil
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
			default:
				return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
			}
		default:
			return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type: %T", pbEvaluatedExpr)
	}
}

func convertHashedValueComparatorsFromProto(hvc []*HashValueComparator) []evm.HashedValueComparator {
	var parsed []evm.HashedValueComparator
	for i := range hvc {
		hv := evm.HashedValueComparator{}
		for _, v := range hvc[i].Values {
			hv.Values = append(hv.Values, evmtypes.Hash(v.Hash[:]))
		}
		hv.Operator = primitives.ComparisonOperator(hvc[i].GetOperator())
		parsed = append(parsed, hv)
	}
	return parsed
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

func ConvertLimitAndSortFromProto(limitAndSort *pb.LimitAndSort) (query.LimitAndSort, error) {
	sortByArr := make([]query.SortBy, len(limitAndSort.SortBy))

	for idx, sortBy := range limitAndSort.SortBy {
		switch sortBy.SortType {
		case pb.SortType_SortTimestamp:
			sortByArr[idx] = query.NewSortByTimestamp(query.SortDirection(sortBy.GetDirection()))
		case pb.SortType_SortBlock:
			sortByArr[idx] = query.NewSortByBlock(query.SortDirection(sortBy.GetDirection()))
		case pb.SortType_SortSequence:
			sortByArr[idx] = query.NewSortBySequence(query.SortDirection(sortBy.GetDirection()))
		default:
			return query.LimitAndSort{}, status.Errorf(codes.InvalidArgument, "Unknown sort by type: %T", sortBy)
		}
	}

	limit := limitAndSort.Limit
	cursorDefined := limit.Cursor != nil
	cursorDirectionDefined := limit.Direction != nil

	if cursorDefined && cursorDirectionDefined {
		return query.NewLimitAndSort(query.CursorLimit(*limit.Cursor, (query.CursorDirection)(*limit.Direction), limit.Count)), nil
	} else if (!cursorDefined && cursorDirectionDefined) || (cursorDefined && !cursorDirectionDefined) {
		return query.LimitAndSort{}, status.Errorf(codes.InvalidArgument, "Limit cursor and cursor direction must both be defined or undefined")
	}

	return query.NewLimitAndSort(query.CountLimit(limit.Count), sortByArr...), nil
}
