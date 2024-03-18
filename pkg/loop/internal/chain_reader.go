package internal

import (
	"context"
	jsonv1 "encoding/json"
	"fmt"

	jsonv2 "github.com/go-json-experiment/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/fxamacker/cbor/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.ChainReader = (*chainReaderClient)(nil)

// NewChainReaderTestClient is a test client for [types.ChainReader]
// internal users should instantiate a client directly and set all private fields
func NewChainReaderTestClient(conn *grpc.ClientConn) types.ChainReader {
	return &chainReaderClient{grpc: pb.NewChainReaderClient(conn)}
}

type chainReaderClient struct {
	*BrokerExt
	grpc pb.ChainReaderClient
}

// enum of all known encoding formats for versioned data
const (
	JSONEncodingVersion1 = iota
	JSONEncodingVersion2
	CBOREncodingVersion
)

// Version to be used for encoding (version used for decoding is determined by data received)
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
		return wrapRPCErr(err)
	}

	return DecodeVersionedBytes(retVal, reply.RetVal)
}

func (c *chainReaderClient) QueryKey(ctx context.Context, key string, queryFilter types.QueryFilter, limitAndSort types.LimitAndSort) ([]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(queryFilter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	_, err = c.grpc.QueryKey(ctx, &pb.QueryKeyRequest{Key: key, QueryFilter: pbQueryFilter, LimitAndSort: pbLimitAndSort})
	if err != nil {
		return nil, wrapRPCErr(err)
	}

	return nil, nil
}

func (c *chainReaderClient) QueryKeys(ctx context.Context, keys []string, queryFilter types.QueryFilter, limitAndSort types.LimitAndSort) ([][]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(queryFilter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	_, err = c.grpc.QueryKeys(ctx, &pb.QueryKeysRequest{Keys: keys, QueryFilter: pbQueryFilter, LimitAndSort: pbLimitAndSort})
	if err != nil {
		return nil, wrapRPCErr(err)
	}
	return nil, nil
}

func (c *chainReaderClient) QueryKeyByValues(ctx context.Context, key string, values []string, queryFilter types.QueryFilter, limitAndSort types.LimitAndSort) ([]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(queryFilter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	_, err = c.grpc.QueryKeyByValues(ctx, &pb.QueryKeyByValuesRequest{Key: key, KeyValues: &pb.KeyValues{Values: values}, QueryFilter: pbQueryFilter, LimitAndSort: pbLimitAndSort})
	if err != nil {
		return nil, wrapRPCErr(err)
	}

	return nil, nil
}

func (c *chainReaderClient) QueryKeysByValues(ctx context.Context, keys []string, values [][]string, queryFilter types.QueryFilter, limitAndSort types.LimitAndSort) ([][]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(queryFilter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	var pbKeyValues []*pb.KeyValues
	for _, keyValues := range values {
		pbKeyValues = append(pbKeyValues, &pb.KeyValues{Values: keyValues})
	}

	_, err = c.grpc.QueryKeysByValues(ctx, &pb.QueryKeysByValuesRequest{Keys: keys, KeysValues: pbKeyValues, QueryFilter: pbQueryFilter, LimitAndSort: pbLimitAndSort})
	if err != nil {
		return nil, wrapRPCErr(err)
	}

	return nil, nil
}

func (c *chainReaderClient) Bind(ctx context.Context, bindings []types.BoundContract) error {
	pbBindings := make([]*pb.BoundContract, len(bindings))
	for i, b := range bindings {
		pbBindings[i] = &pb.BoundContract{Address: b.Address, Name: b.Name, Pending: b.Pending}
	}
	_, err := c.grpc.Bind(ctx, &pb.BindRequest{Bindings: pbBindings})
	return wrapRPCErr(err)
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

func (c *chainReaderServer) QueryKey(ctx context.Context, request *pb.QueryKeyRequest) (*pb.QueryKeysReply, error) {
	queryFilter, err := convertQueryFiltersFromProto(request.QueryFilter)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	_, err = c.impl.QueryKey(ctx, request.Key, queryFilter, limitAndSort)
	if err != nil {
		return nil, err
	}
	return &pb.QueryKeysReply{RetVal: nil}, nil
}

func (c *chainReaderServer) QueryKeys(ctx context.Context, request *pb.QueryKeysRequest) (*pb.QueryKeysReply, error) {
	queryFilters, err := convertQueryFiltersFromProto(request.QueryFilter)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	_, err = c.impl.QueryKeys(ctx, request.Keys, queryFilters, limitAndSort)
	if err != nil {
		return nil, err
	}
	return &pb.QueryKeysReply{RetVal: nil}, nil
}

func (c *chainReaderServer) QueryKeyByValues(ctx context.Context, request *pb.QueryKeyByValuesRequest) (*pb.QueryKeysReply, error) {
	queryFilters, err := convertQueryFiltersFromProto(request.QueryFilter)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	var values []string
	if request.KeyValues != nil {
		values = request.KeyValues.Values
	}

	_, err = c.impl.QueryKeyByValues(ctx, request.Key, values, queryFilters, limitAndSort)
	if err != nil {
		return nil, err
	}
	return &pb.QueryKeysReply{RetVal: nil}, nil
}

func (c *chainReaderServer) QueryKeysByValues(ctx context.Context, request *pb.QueryKeysByValuesRequest) (*pb.QueryKeysReply, error) {
	queryFilters, err := convertQueryFiltersFromProto(request.QueryFilter)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	var values [][]string
	if request.KeysValues != nil {
		for _, keyValues := range request.KeysValues {
			values = append(values, keyValues.Values)
		}
	}

	_, err = c.impl.QueryKeysByValues(ctx, request.Keys, values, queryFilters, limitAndSort)
	if err != nil {
		return nil, err
	}
	return &pb.QueryKeysReply{RetVal: nil}, nil
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

func convertQueryFilterToProto(queryFilter types.QueryFilter) (*pb.QueryFilter, error) {
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

func convertExpressionToProto(expression types.Expression) (*pb.Expression, error) {
	pbExpression := &pb.Expression{}
	if expression.IsPrimitive() {
		pbExpression.Evaluator = &pb.Expression_Primitive{Primitive: &pb.Primitive{}}
		switch primitive := expression.Primitive.(type) {
		case *types.AddressFilter:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_AddressFilter{
				AddressFilter: &pb.AddressFilter{
					Addresses: primitive.Addresses,
				}}
		case *types.ConfirmationsFilter:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_ConfirmationsFilter{
				ConfirmationsFilter: &pb.ConfirmationsFilter{
					Confirmations: pb.Confirmations(primitive.Confirmations),
				}}
		case *types.BlockFilter:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_BlockFilter{
				BlockFilter: &pb.BlockFilter{
					BlockNumber: primitive.Block,
					Operator:    pb.ComparisonOperator(primitive.Operator),
				}}
		case *types.TxHashFilter:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_TxHashFilter{
				TxHashFilter: &pb.TxHashFilter{
					TxHash: primitive.TxHash,
				}}
		case *types.TimestampFilter:
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
		for _, expr := range expression.BooleanExpression.Expressions {
			pbExpr, err := convertExpressionToProto(expr)
			if err != nil {
				return nil, err
			}
			expressions = append(expressions, pbExpr)
		}
		pbExpression.Evaluator = &pb.Expression_BooleanExpression{
			BooleanExpression: &pb.BooleanExpression{
				BooleanOperator: pb.BooleanOperator(expression.BooleanExpression.BooleanOperator),
				Expression:      expressions,
			}}
	}

	return pbExpression, nil
}

func convertLimitAndSortToProto(limitAndSort types.LimitAndSort) (*pb.LimitAndSort, error) {
	var sortByArr []*pb.SortBy
	for _, sortBy := range limitAndSort.SortBy {
		switch sort := sortBy.(type) {
		case *types.SortByTimestamp:
			sortByArr = append(sortByArr,
				&pb.SortBy{SortBy: &pb.SortBy_SortByTimestamp{
					SortByTimestamp: &pb.SortByTimestamp{
						SortDirection: pb.SortDirection(sort.GetDirection()),
					}}})

		case *types.SortByBlock:
			sortByArr = append(sortByArr,
				&pb.SortBy{SortBy: &pb.SortBy_SortByBlock{
					SortByBlock: &pb.SortByBlock{
						SortDirection: pb.SortDirection(sort.GetDirection()),
					}}})
		case *types.SortBySequence:
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

func convertQueryFiltersFromProto(pbQueryFilters *pb.QueryFilter) (types.QueryFilter, error) {
	var queryFilter types.QueryFilter
	for _, pbQueryFilter := range pbQueryFilters.Expression {
		expression, err := convertExpressionFromProto(pbQueryFilter)
		if err != nil {
			return types.QueryFilter{}, err
		}
		queryFilter.Expressions = append(queryFilter.Expressions, expression)
	}
	return queryFilter, nil
}

func convertExpressionFromProto(pbExpression *pb.Expression) (types.Expression, error) {
	switch pbEvaluatedExpr := pbExpression.Evaluator.(type) {
	case *pb.Expression_BooleanExpression:
		var expressions []types.Expression
		for _, expression := range pbEvaluatedExpr.BooleanExpression.Expression {
			convertedExpression, err := convertExpressionFromProto(expression)
			if err != nil {
				return types.Expression{}, err
			}
			expressions = append(expressions, convertedExpression)
		}
		return types.NewBooleanExpression(types.BooleanOperator(pbEvaluatedExpr.BooleanExpression.BooleanOperator), expressions)
	case *pb.Expression_Primitive:
		switch primitive := pbEvaluatedExpr.Primitive.Comparator.(type) {
		case *pb.Primitive_AddressFilter:
			return types.NewAddressesPrimitive(primitive.AddressFilter.Addresses...), nil
		case *pb.Primitive_ConfirmationsFilter:
			return types.NewConfirmationsPrimitive(types.Confirmations(primitive.ConfirmationsFilter.Confirmations)), nil
		case *pb.Primitive_BlockFilter:
			return types.NewBlockPrimitive(primitive.BlockFilter.BlockNumber, types.ComparisonOperator(primitive.BlockFilter.Operator)), nil
		case *pb.Primitive_TxHashFilter:
			return types.NewTxHashPrimitive(primitive.TxHashFilter.TxHash), nil
		case *pb.Primitive_TimestampFilter:
			return types.NewTimestampPrimitive(primitive.TimestampFilter.Timestamp, types.ComparisonOperator(primitive.TimestampFilter.Operator)), nil
		default:
			return types.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type")
		}
	default:
		return types.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type")
	}
}

func convertLimitAndSortFromProto(limitAndSort *pb.LimitAndSort) (types.LimitAndSort, error) {
	var sortByArr []types.SortBy
	for _, sortBy := range limitAndSort.SortBy {
		switch sort := sortBy.SortBy.(type) {
		case *pb.SortBy_SortByTimestamp:
			sortByArr = append(sortByArr, types.NewSortByTimestamp(types.SortDirection(sort.SortByTimestamp.GetSortDirection())))
		case *pb.SortBy_SortByBlock:
			sortByArr = append(sortByArr, types.NewSortByBlock(types.SortDirection(sort.SortByBlock.GetSortDirection())))
		case *pb.SortBy_SortBySequence:
			sortByArr = append(sortByArr, types.NewSortBySequence(types.SortDirection(sort.SortBySequence.GetSortDirection())))
		default:
			return types.LimitAndSort{}, status.Errorf(codes.InvalidArgument, "Unknown order by type")
		}
	}

	return types.NewLimitAndSort(limitAndSort.Limit, sortByArr...), nil
}
