package chaincommonpb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	jsonv2 "github.com/go-json-experiment/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type EncodingVersion uint32

func (v EncodingVersion) Uint32() uint32 {
	return uint32(v)
}

// enum of all known encoding formats for versioned data.
const (
	JSONEncodingVersion1 EncodingVersion = iota
	JSONEncodingVersion2
	CBOREncodingVersion
	ValuesEncodingVersion
)

const DefaultEncodingVersion = CBOREncodingVersion

type ValueEncoder func(value any) (*VersionedBytes, error)

func EncodeVersionedBytes(data any, version EncodingVersion) (*VersionedBytes, error) {
	var byt []byte
	var err error

	switch version {
	case JSONEncodingVersion1:
		byt, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	case JSONEncodingVersion2:
		byt, err = jsonv2.Marshal(data, jsonv2.StringifyNumbers(true))
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	case CBOREncodingVersion:
		enco := cbor.CoreDetEncOptions()
		enco.Time = cbor.TimeRFC3339Nano
		var enc cbor.EncMode
		enc, err = enco.EncMode()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInternal, err)
		}
		byt, err = enc.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	case ValuesEncodingVersion:
		val, err := values.Wrap(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
		byt, err = proto.Marshal(values.Proto(val))
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	default:
		return nil, fmt.Errorf("%w: unsupported encoding version %d for data %v", types.ErrInvalidEncoding, version, data)
	}

	return &VersionedBytes{Version: version.Uint32(), Data: byt}, nil
}

func DecodeVersionedBytes(res any, vData *VersionedBytes) error {
	if vData == nil {
		return errors.New("cannot decode nil versioned bytes")
	}

	var err error
	switch EncodingVersion(vData.Version) {
	case JSONEncodingVersion1:
		decoder := json.NewDecoder(bytes.NewBuffer(vData.Data))
		decoder.UseNumber()

		err = decoder.Decode(res)
	case JSONEncodingVersion2:
		err = jsonv2.Unmarshal(vData.Data, res, jsonv2.StringifyNumbers(true))
	case CBOREncodingVersion:
		decopt := cbor.DecOptions{UTF8: cbor.UTF8DecodeInvalid}
		var dec cbor.DecMode
		dec, err = decopt.DecMode()
		if err != nil {
			return fmt.Errorf("%w: %w", types.ErrInternal, err)
		}
		err = dec.Unmarshal(vData.Data, res)
	case ValuesEncodingVersion:
		protoValue := &valuespb.Value{}
		err = proto.Unmarshal(vData.Data, protoValue)
		if err != nil {
			return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}

		var value values.Value
		value, err = values.FromProto(protoValue)
		if err != nil {
			return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}

		valuePtr, ok := res.(*values.Value)
		if ok {
			*valuePtr = value
		} else {
			err = value.UnwrapTo(res)
			if err != nil {
				return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
			}
		}
	default:
		return fmt.Errorf("unsupported encoding version %d for versionedData %v", vData.Version, vData.Data)
	}

	if err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}

	return nil
}

func ConvertExpressionFromProto(
	pbExpression *Expression,
	encodedTypeGetter func(comparatorName string, forEncoding bool) (any, error),
) (query.Expression, error) {
	switch pbEvaluatedExpr := pbExpression.Evaluator.(type) {
	case *Expression_BooleanExpression:
		var expressions []query.Expression
		for _, expression := range pbEvaluatedExpr.BooleanExpression.Expression {
			converted, err := ConvertExpressionFromProto(expression, encodedTypeGetter)
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
		expr, err := ConvertPrimitiveFromProto(pbEvaluatedExpr.Primitive, encodedTypeGetter)
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

func ConvertExpressionToProto(expression query.Expression, encodeValue ValueEncoder) (*Expression, error) {
	pbExpression := &Expression{}
	if expression.IsPrimitive() {
		primitive, err := ConvertPrimitiveToProto(expression.Primitive, encodeValue)
		if err != nil {
			return nil, err
		}
		pbExpression.Evaluator = &Expression_Primitive{Primitive: primitive}
		return pbExpression, nil
	}

	pbExpression.Evaluator = &Expression_BooleanExpression{BooleanExpression: &BooleanExpression{}}
	var expressions []*Expression
	for _, expr := range expression.BoolExpression.Expressions {
		e, err := ConvertExpressionToProto(expr, encodeValue)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}
	pbExpression.Evaluator = &Expression_BooleanExpression{
		BooleanExpression: &BooleanExpression{
			BooleanOperator: BooleanOperator(expression.BoolExpression.BoolOperator),
			Expression:      expressions,
		}}

	return pbExpression, nil
}

func ConvertPrimitiveFromProto(protoPrimitive *Primitive, encodedTypeGetter func(comparatorName string, forEncoding bool) (any, error)) (query.Expression, error) {
	switch primitive := protoPrimitive.Primitive.(type) {
	case *Primitive_Comparator:
		var valueComparators []primitives.ValueComparator

		for _, pbValueComparator := range primitive.Comparator.ValueComparators {
			val, err := encodedTypeGetter(primitive.Comparator.Name, true)
			if err != nil {
				return query.Expression{}, err
			}

			if err = DecodeVersionedBytes(val, pbValueComparator.Value); err != nil {
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

func ConvertPrimitiveToProto(primitive primitives.Primitive, encodeValue ValueEncoder) (*Primitive, error) {
	primitiveToReturn := &Primitive{}
	switch p := primitive.(type) {
	case *primitives.Comparator:
		var pbValueComparators []*ValueComparator

		for _, vc := range p.ValueComparators {
			versioned, err := encodeValue(vc.Value)
			if err != nil {
				return nil, err
			}
			pbValueComparators = append(pbValueComparators,
				&ValueComparator{
					Value:    versioned,
					Operator: ComparisonOperator(vc.Operator),
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
				Operator:    ComparisonOperator(p.Operator),
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
				Operator:  ComparisonOperator(p.Operator),
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
			SortType:  tp,
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
	default:
		return "", fmt.Errorf("invalid pb confidence level: %d", pbConfidence)
	}
}

func ConvertConfidenceToProto(confidenceLevel primitives.ConfidenceLevel) (Confidence, error) {
	switch confidenceLevel {
	case primitives.Finalized:
		return Confidence_Finalized, nil
	case primitives.Unconfirmed:
		return Confidence_Unconfirmed, nil
	default:
		return -1, fmt.Errorf("invalid confidence level %s", confidenceLevel)
	}
}
