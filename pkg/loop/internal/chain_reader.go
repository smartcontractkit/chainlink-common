package internal

import (
	"context"
	jsonv1 "encoding/json"
	"fmt"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	jsonv2 "github.com/go-json-experiment/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
)

var _ types.ChainReader = (*chainReaderClient)(nil)

// NewChainReaderTestClient is a test client for [types.ChainReader]
// internal users should instantiate a client directly and set all private fields.
func NewChainReaderTestClient(conn *grpc.ClientConn) types.ChainReader {
	return &chainReaderClient{grpc: pb.NewChainReaderClient(conn)}
}

type chainReaderClient struct {
	*net.BrokerExt
	grpc pb.ChainReaderClient
}

// enum of all known encoding formats for versioned data.
const (
	JSONEncodingVersion1 = iota
	JSONEncodingVersion2
	CBOREncodingVersion
)

// Version to be used for encoding (version used for decoding is determined by data received).
const CurrentEncodingVersion = CBOREncodingVersion

func EncodeVersionedBytes(data any, version uint32) (*pb.VersionedBytes, error) {
	var bytes []byte
	var err error

	switch version {
	case JSONEncodingVersion1:
		bytes, err = jsonv1.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	case JSONEncodingVersion2:
		bytes, err = jsonv2.Marshal(data)
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
		bytes, err = enc.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	default:
		return nil, fmt.Errorf("%w: unsupported encoding version %d for data %v", types.ErrInvalidEncoding, version, data)
	}

	return &pb.VersionedBytes{Version: version, Data: bytes}, nil
}

func DecodeVersionedBytes(res any, vData *pb.VersionedBytes) error {
	var err error
	switch vData.Version {
	case JSONEncodingVersion1:
		err = jsonv1.Unmarshal(vData.Data, res)
	case JSONEncodingVersion2:
		err = jsonv2.Unmarshal(vData.Data, res)
	case CBOREncodingVersion:
		decopt := cbor.DecOptions{UTF8: cbor.UTF8DecodeInvalid}
		var dec cbor.DecMode
		dec, err = decopt.DecMode()
		if err != nil {
			return fmt.Errorf("%w: %w", types.ErrInternal, err)
		}
		err = dec.Unmarshal(vData.Data, res)
	default:
		return fmt.Errorf("unsupported encoding version %d for versionedData %v", vData.Version, vData.Data)
	}

	if err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}
	return nil
}

func (c *chainReaderClient) GetLatestValue(ctx context.Context, contractName, method string, params, retVal any) error {
	versionedParams, err := EncodeVersionedBytes(params, CurrentEncodingVersion)
	if err != nil {
		return err
	}

	reply, err := c.grpc.GetLatestValue(ctx, &pb.GetLatestValueRequest{ContractName: contractName, Method: method, Params: versionedParams})
	if err != nil {
		return net.WrapRPCErr(err)
	}

	return DecodeVersionedBytes(retVal, reply.RetVal)
}

func (c *chainReaderClient) QueryKey(ctx context.Context, key string, queryFilter query.Filter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(queryFilter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	pbSequences, err := c.grpc.QueryKey(ctx, &pb.QueryKeyRequest{Key: key, QueryFilter: pbQueryFilter, LimitAndSort: pbLimitAndSort})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return convertSequencesFromProto(pbSequences.Sequences, sequenceDataType)
}

func (c *chainReaderClient) QueryKeys(ctx context.Context, keys []string, queryFilter query.Filter, limitAndSort query.LimitAndSort, sequenceDataTypes []any) ([][]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(queryFilter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	pbSequencesMatrix, err := c.grpc.QueryKeys(ctx, &pb.QueryKeysRequest{Keys: keys, QueryFilter: pbQueryFilter, LimitAndSort: pbLimitAndSort})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return convertSequencesMatrixFromProto(pbSequencesMatrix.Sequences, sequenceDataTypes)
}

func (c *chainReaderClient) QueryByKeyValuesComparison(ctx context.Context, keyValuesComparator query.KeyValuesComparator, queryFilter query.Filter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(queryFilter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	pbSequences, err := c.grpc.QueryByKeyValuesComparison(ctx, &pb.QueryByKeyValuesComparisonRequest{KeyValuesComparator: keyValuesByEqualityToProto(keyValuesComparator), QueryFilter: pbQueryFilter, LimitAndSort: pbLimitAndSort})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return convertSequencesFromProto(pbSequences.Sequences, sequenceDataType)
}

func (c *chainReaderClient) QueryByKeysValuesComparison(ctx context.Context, keysValuesComparator []query.KeyValuesComparator, queryFilter query.Filter, limitAndSort query.LimitAndSort, sequenceDataTypes []any) ([][]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(queryFilter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	var pbKeysValuesComparator []*pb.KeyValuesComparator
	for _, kvComparator := range keysValuesComparator {
		pbKeysValuesComparator = append(pbKeysValuesComparator, keyValuesByEqualityToProto(kvComparator))
	}

	pbSequencesMatrix, err := c.grpc.QueryByKeysValuesComparison(ctx, &pb.QueryByKeysValuesComparisonRequest{KeysValuesComparator: pbKeysValuesComparator, QueryFilter: pbQueryFilter, LimitAndSort: pbLimitAndSort})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return convertSequencesMatrixFromProto(pbSequencesMatrix.Sequences, sequenceDataTypes)
}

func (c *chainReaderClient) Bind(ctx context.Context, bindings []types.BoundContract) error {
	pbBindings := make([]*pb.BoundContract, len(bindings))
	for i, b := range bindings {
		pbBindings[i] = &pb.BoundContract{Address: b.Address, Name: b.Name, Pending: b.Pending}
	}
	_, err := c.grpc.Bind(ctx, &pb.BindRequest{Bindings: pbBindings})
	return net.WrapRPCErr(err)
}

var _ pb.ChainReaderServer = (*chainReaderServer)(nil)

func NewChainReaderServer(impl types.ChainReader) pb.ChainReaderServer {
	return &chainReaderServer{impl: impl}
}

type chainReaderServer struct {
	pb.UnimplementedChainReaderServer
	impl types.ChainReader
}

func (c *chainReaderServer) GetLatestValue(ctx context.Context, request *pb.GetLatestValueRequest) (*pb.GetLatestValueReply, error) {
	contractName := request.ContractName
	params, err := getContractEncodedType(contractName, request.Method, c.impl, true)
	if err != nil {
		return nil, err
	}

	if err = DecodeVersionedBytes(params, request.Params); err != nil {
		return nil, err
	}

	retVal, err := getContractEncodedType(contractName, request.Method, c.impl, false)
	if err != nil {
		return nil, err
	}
	err = c.impl.GetLatestValue(ctx, contractName, request.Method, params, retVal)
	if err != nil {
		return nil, err
	}

	encodedRetVal, err := EncodeVersionedBytes(retVal, request.Params.Version)
	if err != nil {
		return nil, err
	}

	return &pb.GetLatestValueReply{RetVal: encodedRetVal}, nil
}

func (c *chainReaderServer) QueryKey(ctx context.Context, request *pb.QueryKeyRequest) (*pb.QueryKeyReply, error) {
	queryFilter, err := convertQueryFiltersFromProto(request.QueryFilter)
	if err != nil {
		return nil, err
	}

	sequenceDataType, err := getContractEncodedTypeByKey(request.Key, c.impl, false)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	sequences, err := c.impl.QueryKey(ctx, request.Key, queryFilter, limitAndSort, sequenceDataType)
	if err != nil {
		return nil, err
	}

	pbSequences, err := convertSequencesToProto(sequences, sequenceDataType)
	if err != nil {
		return nil, err
	}

	return &pb.QueryKeyReply{Sequences: pbSequences}, nil
}

func (c *chainReaderServer) QueryKeys(ctx context.Context, request *pb.QueryKeysRequest) (*pb.QueryKeysReply, error) {

	queryFilter, err := convertQueryFiltersFromProto(request.QueryFilter)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	var sequenceDataTypes []any
	for _, key := range request.Keys {
		sequenceDataType, err := getContractEncodedTypeByKey(key, c.impl, false)
		if err != nil {
			return nil, err
		}

		sequenceDataTypes = append(sequenceDataTypes, sequenceDataType)
	}

	sequencesMatrix, err := c.impl.QueryKeys(ctx, request.Keys, queryFilter, limitAndSort, sequenceDataTypes)
	if err != nil {
		return nil, err
	}

	var pbSequencesMatrix []*pb.Sequences
	for i, sequences := range sequencesMatrix {
		pbSequences, err := convertSequencesToProto(sequences, sequenceDataTypes[i])
		if err != nil {
			return nil, err
		}
		pbSequencesMatrix = append(pbSequencesMatrix, pbSequences)
	}

	return &pb.QueryKeysReply{Sequences: pbSequencesMatrix}, nil
}

func (c *chainReaderServer) QueryByKeyValuesComparison(ctx context.Context, request *pb.QueryByKeyValuesComparisonRequest) (*pb.QueryByKeyValuesComparisonReply, error) {
	if request.KeyValuesComparator == nil {
		return nil, fmt.Errorf("key and values not defined")
	}

	queryFilters, err := convertQueryFiltersFromProto(request.QueryFilter)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	sequenceDataType, err := getContractEncodedTypeByKey(request.KeyValuesComparator.Key, c.impl, false)
	if err != nil {
		return nil, err
	}

	sequences, err := c.impl.QueryByKeyValuesComparison(ctx, keyValuesComparatorFromProto(request.KeyValuesComparator), queryFilters, limitAndSort, sequenceDataType)
	if err != nil {
		return nil, err
	}

	pbSequences, err := convertSequencesToProto(sequences, sequenceDataType)
	if err != nil {
		return nil, err
	}

	return &pb.QueryByKeyValuesComparisonReply{Sequences: pbSequences}, nil
}

func (c *chainReaderServer) QueryByKeysValuesComparison(ctx context.Context, request *pb.QueryByKeysValuesComparisonRequest) (*pb.QueryByKeysValuesComparisonReply, error) {
	if request.KeysValuesComparator != nil {
		return nil, fmt.Errorf("keys and values not defined")
	}

	var keysValuesByEquality []query.KeyValuesComparator
	for _, keyValuesByEquality := range request.KeysValuesComparator {
		keysValuesByEquality = append(keysValuesByEquality, keyValuesComparatorFromProto(keyValuesByEquality))
	}

	var sequenceDataTypes []any
	for _, keyByEquality := range keysValuesByEquality {
		sequenceDataType, err := getContractEncodedTypeByKey(keyByEquality.Key, c.impl, false)
		if err != nil {
			return nil, err
		}
		sequenceDataTypes = append(sequenceDataTypes, sequenceDataType)
	}

	queryFilter, err := convertQueryFiltersFromProto(request.QueryFilter)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	sequencesMatrix, err := c.impl.QueryByKeysValuesComparison(ctx, keysValuesByEquality, queryFilter, limitAndSort, sequenceDataTypes)
	if err != nil {
		return nil, err
	}

	var pbSequencesMatrix []*pb.Sequences
	for i, sequences := range sequencesMatrix {
		pbSequences, err := convertSequencesToProto(sequences, sequenceDataTypes[i])
		if err != nil {
			return nil, err
		}
		pbSequencesMatrix = append(pbSequencesMatrix, pbSequences)
	}

	return &pb.QueryByKeysValuesComparisonReply{Sequences: pbSequencesMatrix}, nil
}

func (c *chainReaderServer) Bind(ctx context.Context, bindings *pb.BindRequest) (*emptypb.Empty, error) {
	tBindings := make([]types.BoundContract, len(bindings.Bindings))
	for i, b := range bindings.Bindings {
		tBindings[i] = types.BoundContract{Address: b.Address, Name: b.Name, Pending: b.Pending}
	}

	return &emptypb.Empty{}, c.impl.Bind(ctx, tBindings)
}

func getContractEncodedType(contractName, itemType string, possibleTypeProvider any, forEncoding bool) (any, error) {
	if ctp, ok := possibleTypeProvider.(types.ContractTypeProvider); ok {
		return ctp.CreateContractType(contractName, itemType, forEncoding)
	}

	return &map[string]any{}, nil
}

func getContractEncodedTypeByKey(key string, possibleTypeProvider any, forEncoding bool) (any, error) {
	if ctp, ok := possibleTypeProvider.(types.ContractTypeProvider); ok {
		return ctp.CreateContractTypeByKey(key, forEncoding)
	}

	return &map[string]any{}, nil
}

func convertQueryFilterToProto(queryFilter query.Filter) (*pb.QueryFilter, error) {
	pbQueryFilter := &pb.QueryFilter{}
	for _, expression := range queryFilter.Expressions {
		pbExpression, err := convertExpressionToProto(expression)
		if err != nil {
			return nil, err
		}
		pbQueryFilter.Expression = append(pbQueryFilter.Expression, pbExpression)
	}

	return pbQueryFilter, nil
}

func convertExpressionToProto(expression query.Expression) (*pb.Expression, error) {
	pbExpression := &pb.Expression{}
	if expression.IsPrimitive() {
		pbExpression.Evaluator = &pb.Expression_Primitive{Primitive: &pb.Primitive{}}
		switch primitive := expression.Primitive.(type) {
		case *query.AddressPrimitive:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_AddressFilter{
				AddressFilter: &pb.AddressFilter{
					Addresses: primitive.Addresses,
				}}
		case *query.ConfirmationsPrimitive:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_ConfirmationsFilter{
				ConfirmationsFilter: &pb.ConfirmationsFilter{
					Confirmations: pb.Confirmations(primitive.Confirmations),
				}}
		case *query.BlockPrimitive:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_BlockFilter{
				BlockFilter: &pb.BlockFilter{
					BlockNumber: primitive.Block,
					Operator:    pb.ComparisonOperator(primitive.Operator),
				}}
		case *query.TxHashPrimitive:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_TxHashFilter{
				TxHashFilter: &pb.TxHashFilter{
					TxHash: primitive.TxHash,
				}}
		case *query.TimestampPrimitive:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_TimestampFilter{
				TimestampFilter: &pb.TimestampFilter{
					Timestamp: primitive.Timestamp,
					Operator:  pb.ComparisonOperator(primitive.Operator),
				}}
		default:
			return nil, status.Errorf(codes.InvalidArgument, "Unknown expression type")
		}
		return pbExpression, nil
	} else {
		pbExpression.Evaluator = &pb.Expression_BooleanExpression{BooleanExpression: &pb.BooleanExpression{}}
		var expressions []*pb.Expression
		for _, expr := range expression.BoolExpression.Expressions {
			pbExpr, err := convertExpressionToProto(expr)
			if err != nil {
				return nil, err
			}
			expressions = append(expressions, pbExpr)
		}
		pbExpression.Evaluator = &pb.Expression_BooleanExpression{
			BooleanExpression: &pb.BooleanExpression{
				BooleanOperator: pb.BooleanOperator(expression.BoolExpression.BoolOperator),
				Expression:      expressions,
			}}
	}

	return pbExpression, nil
}

func keyValuesByEqualityToProto(keyValuesComparator query.KeyValuesComparator) *pb.KeyValuesComparator {
	var pbValueComparator []*pb.ValueComparator
	for _, valueComparators := range keyValuesComparator.ValueComparators {
		pbValueComparator = append(pbValueComparator, &pb.ValueComparator{
			Value:    valueComparators.Value,
			Operator: pb.ComparisonOperator(valueComparators.Operator),
		})
	}
	return &pb.KeyValuesComparator{Key: keyValuesComparator.Key, ValueComparators: pbValueComparator}
}

func convertLimitAndSortToProto(limitAndSort query.LimitAndSort) (*pb.LimitAndSort, error) {
	var sortByArr []*pb.SortBy
	for _, sortBy := range limitAndSort.SortBy {
		switch sort := sortBy.(type) {
		case *query.SortByTimestamp:
			sortByArr = append(sortByArr,
				&pb.SortBy{SortBy: &pb.SortBy_SortByTimestamp{
					SortByTimestamp: &pb.SortByTimestamp{
						SortDirection: pb.SortDirection(sort.GetDirection()),
					}}})

		case *query.SortByBlock:
			sortByArr = append(sortByArr,
				&pb.SortBy{SortBy: &pb.SortBy_SortByBlock{
					SortByBlock: &pb.SortByBlock{
						SortDirection: pb.SortDirection(sort.GetDirection()),
					}}})
		case *query.SortBySequence:
			sortByArr = append(sortByArr,
				&pb.SortBy{SortBy: &pb.SortBy_SortBySequence{
					SortBySequence: &pb.SortBySequence{
						SortDirection: pb.SortDirection(sort.GetDirection()),
					}}})
		default:
			return &pb.LimitAndSort{}, status.Errorf(codes.InvalidArgument, "Unknown order by type")
		}
	}
	return &pb.LimitAndSort{Limit: limitAndSort.Limit, SortBy: sortByArr}, nil
}

func convertSequencesToProto(sequences []types.Sequence, sequenceDataType any) (*pb.Sequences, error) {
	var pbSequences []*pb.Sequence
	for _, sequence := range sequences {
		versionedSequenceDataType, err := EncodeVersionedBytes(sequenceDataType, CurrentEncodingVersion)
		if err != nil {
			return nil, err
		}

		pbSequence := &pb.Sequence{
			SequenceCursor: sequence.Cursor,
			Head: &pb.Head{
				Number:    sequence.Number,
				Hash:      sequence.Hash,
				Timestamp: sequence.Timestamp,
			},
			Data: versionedSequenceDataType,
		}
		pbSequences = append(pbSequences, pbSequence)
	}
	return &pb.Sequences{Sequences: pbSequences}, nil
}

func convertQueryFiltersFromProto(pbQueryFilters *pb.QueryFilter) (query.Filter, error) {
	var queryFilter query.Filter
	for _, pbQueryFilter := range pbQueryFilters.Expression {
		expression, err := convertExpressionFromProto(pbQueryFilter)
		if err != nil {
			return query.Filter{}, err
		}
		queryFilter.Expressions = append(queryFilter.Expressions, expression)
	}
	return queryFilter, nil
}

func convertExpressionFromProto(pbExpression *pb.Expression) (query.Expression, error) {
	switch pbEvaluatedExpr := pbExpression.Evaluator.(type) {
	case *pb.Expression_BooleanExpression:
		var expressions []query.Expression
		for _, expression := range pbEvaluatedExpr.BooleanExpression.Expression {
			convertedExpression, err := convertExpressionFromProto(expression)
			if err != nil {
				return query.Expression{}, err
			}
			expressions = append(expressions, convertedExpression)
		}

		if pbEvaluatedExpr.BooleanExpression.BooleanOperator == pb.BooleanOperator_AND {
			return query.NewAndBoolExpression(expressions...), nil
		}
		return query.NewOrBoolExpression(expressions...), nil
	case *pb.Expression_Primitive:
		switch primitive := pbEvaluatedExpr.Primitive.Comparator.(type) {
		case *pb.Primitive_AddressFilter:
			return query.NewAddressesPrimitive(primitive.AddressFilter.Addresses...), nil
		case *pb.Primitive_ConfirmationsFilter:
			return query.NewConfirmationsPrimitive(query.Confirmations(primitive.ConfirmationsFilter.Confirmations)), nil
		case *pb.Primitive_BlockFilter:
			return query.NewBlockPrimitive(primitive.BlockFilter.BlockNumber, query.ComparisonOperator(primitive.BlockFilter.Operator)), nil
		case *pb.Primitive_TxHashFilter:
			return query.NewTxHashPrimitive(primitive.TxHashFilter.TxHash), nil
		case *pb.Primitive_TimestampFilter:
			return query.NewTimestampPrimitive(primitive.TimestampFilter.Timestamp, query.ComparisonOperator(primitive.TimestampFilter.Operator)), nil
		default:
			return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type")
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type")
	}
}

func keyValuesComparatorFromProto(pbKeysByEquality *pb.KeyValuesComparator) query.KeyValuesComparator {
	var valuesEqualities []query.ValueComparator
	for _, valueEquality := range pbKeysByEquality.ValueComparators {
		valuesEqualities = append(valuesEqualities, query.ValueComparator{
			Value:    valueEquality.Value,
			Operator: query.ComparisonOperator(valueEquality.Operator),
		})
	}

	return query.KeyValuesComparator{
		Key:              pbKeysByEquality.Key,
		ValueComparators: valuesEqualities,
	}
}

func convertLimitAndSortFromProto(limitAndSort *pb.LimitAndSort) (query.LimitAndSort, error) {
	var sortByArr []query.SortBy
	for _, sortBy := range limitAndSort.SortBy {
		switch sort := sortBy.SortBy.(type) {
		case *pb.SortBy_SortByTimestamp:
			sortByArr = append(sortByArr, query.NewSortByTimestamp(query.SortDirection(sort.SortByTimestamp.GetSortDirection())))
		case *pb.SortBy_SortByBlock:
			sortByArr = append(sortByArr, query.NewSortByBlock(query.SortDirection(sort.SortByBlock.GetSortDirection())))
		case *pb.SortBy_SortBySequence:
			sortByArr = append(sortByArr, query.NewSortBySequence(query.SortDirection(sort.SortBySequence.GetSortDirection())))
		default:
			return query.LimitAndSort{}, status.Errorf(codes.InvalidArgument, "Unknown order by type")
		}
	}

	return query.NewLimitAndSort(limitAndSort.Limit, sortByArr...), nil
}

func convertSequencesMatrixFromProto(pbSequencesMatrix []*pb.Sequences, sequenceDataTypes []any) ([][]types.Sequence, error) {
	var sequencesMatrix [][]types.Sequence
	for i, sequences := range pbSequencesMatrix {
		convertedSequences, err := convertSequencesFromProto(sequences, sequenceDataTypes[i])
		if err != nil {
			return nil, err
		}

		sequencesMatrix = append(sequencesMatrix, convertedSequences)
	}
	return sequencesMatrix, nil
}

func convertSequencesFromProto(pbSequences *pb.Sequences, sequenceDataType any) ([]types.Sequence, error) {
	var sequences []types.Sequence
	for _, pbSequence := range pbSequences.Sequences {
		data := reflect.New(reflect.TypeOf(sequenceDataType).Elem())
		if err := DecodeVersionedBytes(data, pbSequence.Data); err != nil {
			return nil, err
		}

		sequence := types.Sequence{
			Cursor: pbSequence.SequenceCursor,
			Head: types.Head{
				Number:    pbSequence.Head.Number,
				Hash:      pbSequence.Head.Hash,
				Timestamp: pbSequence.Head.Timestamp,
			},
			Data: data,
		}
		sequences = append(sequences, sequence)
	}
	return sequences, nil
}
