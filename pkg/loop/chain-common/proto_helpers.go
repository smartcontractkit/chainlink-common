package chaincommonpb

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	loopjson "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/json"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

func ConvertExpressionFromProto(pbExpression *Expression) (query.Expression, error) {
	switch pbEvaluatedExpr := pbExpression.Evaluator.(type) {
	case *Expression_BooleanExpression:
		var expressions []query.Expression
		for _, expression := range pbEvaluatedExpr.BooleanExpression.Expression {
			converted, err := ConvertExpressionFromProto(expression)
			if err != nil {
				return query.Expression{}, err
			}
			expressions = append(expressions, converted)
		}
		if pbEvaluatedExpr.BooleanExpression.BooleanOperator == BooleanOperator_AND {
			return query.And(expressions...), nil
		}
		return query.Or(expressions...), nil
	case *Expression_Primitive:
		expr, err := ConvertPrimitiveFromProto(pbEvaluatedExpr.Primitive)
		if err != nil {
			return query.Expression{}, err
		}
		return expr, nil
	default:
		return query.Expression{}, status.Errorf(
			codes.InvalidArgument, "unknown expression type: %T", pbEvaluatedExpr,
		)
	}
}

func ConvertExpressionToProto(expression query.Expression) (*Expression, error) {
	pbExpression := &Expression{}
	if expression.IsPrimitive() {
		primitive, err := ConvertPrimitiveToProto(expression.Primitive)
		if err != nil {
			return nil, err
		}
		pbExpression.Evaluator = &Expression_Primitive{Primitive: primitive}
		return pbExpression, nil
	}

	pbExpression.Evaluator = &Expression_BooleanExpression{BooleanExpression: &BooleanExpression{}}
	expressions := make([]*Expression, 0)
	for _, expr := range expression.BoolExpression.Expressions {
		e, err := ConvertExpressionToProto(expr)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}
	pbExpression.Evaluator = &Expression_BooleanExpression{
		BooleanExpression: &BooleanExpression{
			//nolint: gosec // G115
			BooleanOperator: BooleanOperator(expression.BoolExpression.BoolOperator),
			Expression:      expressions,
		}}

	return pbExpression, nil
}

func ConvertPrimitiveFromProto(protoPrimitive *Primitive) (query.Expression, error) {
	switch primitive := protoPrimitive.Primitive.(type) {
	case *Primitive_Comparator:
		var valueComparators []primitives.ValueComparator

		for _, pbValueComparator := range primitive.Comparator.ValueComparators {
			val, err := loopjson.UnmarshalWithHint(pbValueComparator.Value, pbValueComparator.ValueTypeHint)
			if err != nil {
				return query.Expression{}, err
			}

			valueComparators = append(valueComparators,
				primitives.ValueComparator{
					Value:    val,
					Operator: primitives.ComparisonOperator(pbValueComparator.Operator),
				})
		}
		return query.Comparator(primitive.Comparator.Name, valueComparators...), nil

	case *Primitive_Confidence:
		confidence, err := ConfidenceFromProto(primitive.Confidence)
		return query.Confidence(confidence), err

	case *Primitive_Block:
		return query.Block(primitive.Block.BlockNumber,
			primitives.ComparisonOperator(primitive.Block.Operator)), nil

	case *Primitive_TxHash:
		return query.TxHash(primitive.TxHash.TxHash), nil

	case *Primitive_Timestamp:
		return query.Timestamp(primitive.Timestamp.Timestamp,
			primitives.ComparisonOperator(primitive.Timestamp.Operator)), nil

	default:
		return query.Expression{}, status.Errorf(
			codes.InvalidArgument, "unknown primitive type: %T", primitive,
		)
	}
}

func ConvertPrimitiveToProto(primitive primitives.Primitive) (*Primitive, error) {
	primitiveToReturn := &Primitive{}
	switch p := primitive.(type) {
	case *primitives.Comparator:
		var pbValueComparators []*ValueComparator

		for _, vc := range p.ValueComparators {
			jsonBytes, typeHint, err := loopjson.MarshalWithHint(vc.Value)
			if err != nil {
				return nil, err
			}

			pbValueComparators = append(pbValueComparators,
				&ValueComparator{
					Value: jsonBytes,
					//nolint: gosec // G115
					Operator:      ComparisonOperator(vc.Operator),
					ValueTypeHint: typeHint,
				})
		}

		primitiveToReturn.Primitive = &Primitive_Comparator{
			Comparator: &Comparator{
				Name:             p.Name,
				ValueComparators: pbValueComparators,
			},
		}

	case *primitives.Block:
		primitiveToReturn.Primitive = &Primitive_Block{
			Block: &Block{
				BlockNumber: p.Block,
				//nolint: gosec // G115
				Operator: ComparisonOperator(p.Operator),
			}}
	case *primitives.Confidence:
		pbConfidence, err := ConvertConfidenceToProto(p.ConfidenceLevel)
		if err != nil {
			return nil, err
		}
		primitiveToReturn.Primitive = &Primitive_Confidence{Confidence: pbConfidence}
	case *primitives.Timestamp:
		primitiveToReturn.Primitive = &Primitive_Timestamp{
			Timestamp: &Timestamp{
				Timestamp: p.Timestamp,
				//nolint: gosec // G115
				Operator: ComparisonOperator(p.Operator),
			}}
	case *primitives.TxHash:
		primitiveToReturn.Primitive = &Primitive_TxHash{
			TxHash: &TxHash{
				TxHash: p.TxHash,
			}}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", p)
	}

	return primitiveToReturn, nil
}

func ConvertLimitAndSortFromProto(limitAndSort *LimitAndSort) (query.LimitAndSort, error) {
	if limitAndSort == nil {
		return query.LimitAndSort{}, nil
	}

	sortByArr := make([]query.SortBy, len(limitAndSort.SortBy))
	for idx, sortBy := range limitAndSort.SortBy {
		switch sortBy.SortType {
		case SortType_SortTimestamp:
			sortByArr[idx] = query.NewSortByTimestamp(query.SortDirection(sortBy.GetDirection()))
		case SortType_SortBlock:
			sortByArr[idx] = query.NewSortByBlock(query.SortDirection(sortBy.GetDirection()))
		case SortType_SortSequence:
			sortByArr[idx] = query.NewSortBySequence(query.SortDirection(sortBy.GetDirection()))
		default:
			return query.LimitAndSort{}, status.Errorf(codes.InvalidArgument, "Unknown sort by type: %T", sortBy)
		}
	}

	cursorDefined := false
	cursorDirectionDefined := false
	limit := &Limit{}
	if limitAndSort.Limit != nil {
		limit = limitAndSort.Limit
		cursorDefined = limit.Cursor != nil
		cursorDirectionDefined = limit.Direction != nil
	}

	if cursorDefined && cursorDirectionDefined {
		return query.NewLimitAndSort(query.CursorLimit(*limit.Cursor, (query.CursorDirection)(*limit.Direction), limit.Count)), nil
	} else if (!cursorDefined && cursorDirectionDefined) || (cursorDefined && !cursorDirectionDefined) {
		return query.LimitAndSort{}, status.Errorf(codes.InvalidArgument, "Limit cursor and cursor direction must both be defined or undefined")
	}

	return query.NewLimitAndSort(query.CountLimit(limit.Count), sortByArr...), nil
}

func ConvertLimitAndSortToProto(limitAndSort query.LimitAndSort) (*LimitAndSort, error) {
	sortByArr := make([]*SortBy, len(limitAndSort.SortBy))

	for idx, sortBy := range limitAndSort.SortBy {
		var tp SortType

		switch sort := sortBy.(type) {
		case query.SortByBlock, *query.SortByBlock:
			tp = SortType_SortBlock
		case query.SortByTimestamp, *query.SortByTimestamp:
			tp = SortType_SortTimestamp
		case query.SortBySequence, *query.SortBySequence:
			tp = SortType_SortSequence
		default:
			return &LimitAndSort{}, status.Errorf(codes.InvalidArgument, "Unknown sort by type: %T", sort)
		}

		sortByArr[idx] = &SortBy{
			SortType: tp,
			//nolint: gosec // G115
			Direction: SortDirection(sortBy.GetDirection()),
		}
	}

	pbLimitAndSort := &LimitAndSort{
		SortBy: sortByArr,
		Limit:  &Limit{Count: limitAndSort.Limit.Count},
	}

	cursorDefined := limitAndSort.Limit.Cursor != ""
	cursorDirectionDefined := limitAndSort.Limit.CursorDirection != 0

	if limitAndSort.HasCursorLimit() {
		pbLimitAndSort.Limit.Cursor = &limitAndSort.Limit.Cursor
		pbLimitAndSort.Limit.Direction = (*CursorDirection)(&limitAndSort.Limit.CursorDirection)
	} else if (!cursorDefined && cursorDirectionDefined) || (cursorDefined && !cursorDirectionDefined) {
		return nil, status.Errorf(codes.InvalidArgument, "Limit cursor and cursor direction must both be defined or undefined")
	}

	return pbLimitAndSort, nil
}

func ConfidenceFromProto(pbConfidence Confidence) (primitives.ConfidenceLevel, error) {
	switch pbConfidence {
	case Confidence_Finalized:
		return primitives.Finalized, nil
	case Confidence_Unconfirmed:
		return primitives.Unconfirmed, nil
	case Confidence_Safe:
		return primitives.Safe, nil
	default:
		return "", fmt.Errorf("invalid pb confidence level: %d", pbConfidence)
	}
}

func ConvertConfidenceToProto(confidenceLevel primitives.ConfidenceLevel) (Confidence, error) {
	switch confidenceLevel {
	case primitives.Finalized:
		return Confidence_Finalized, nil
	case "", primitives.Unconfirmed:
		return Confidence_Unconfirmed, nil
	case primitives.Safe:
		return Confidence_Safe, nil
	default:
		return -1, fmt.Errorf("invalid confidence level %s", confidenceLevel)
	}
}
