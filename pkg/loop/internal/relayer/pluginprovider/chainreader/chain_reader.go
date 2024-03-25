package chainreader

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

var _ types.ChainReader = (*Client)(nil)

// NewChainReaderTestClient is a test client for [types.ChainReader]
// internal users should instantiate a client directly and set all private fields.
func NewChainReaderTestClient(conn *grpc.ClientConn) types.ChainReader {
	return &Client{grpc: pb.NewChainReaderClient(conn)}
}

type Client struct {
	*net.BrokerExt
	grpc pb.ChainReaderClient
}

func NewClient(b *net.BrokerExt, cc grpc.ClientConnInterface) *Client {
	return &Client{BrokerExt: b, grpc: pb.NewChainReaderClient(cc)}
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

func (c *Client) GetLatestValue(ctx context.Context, contractName, method string, params, retVal any) error {
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

func (c *Client) QueryKey(ctx context.Context, keyFilter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(keyFilter.Filter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	pbSequences, err := c.grpc.QueryKey(ctx, &pb.QueryKeyRequest{KeyFilter: &pb.KeyFilter{Key: keyFilter.Key, QueryFilter: pbQueryFilter}, LimitAndSort: pbLimitAndSort})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return convertSequencesFromProto(pbSequences.Sequences, sequenceDataType)
}

func (c *Client) QueryKeys(ctx context.Context, keysFilters []query.KeyFilter, limitsAndSorts []query.LimitAndSort, sequenceDataTypes []any) ([][]types.Sequence, error) {
	var pbKeysFilters []*pb.KeyFilter
	for _, keyFilter := range keysFilters {
		pbQueryFilter, err := convertQueryFilterToProto(keyFilter.Filter)
		if err != nil {
			return nil, err
		}
		pbKeysFilters = append(pbKeysFilters, &pb.KeyFilter{Key: keyFilter.Key, QueryFilter: pbQueryFilter})
	}

	var pbLimitsAndSorts []*pb.LimitAndSort
	for _, limitAndSort := range limitsAndSorts {
		pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
		if err != nil {
			return nil, err
		}
		pbLimitsAndSorts = append(pbLimitsAndSorts, pbLimitAndSort)
	}

	pbSequencesMatrix, err := c.grpc.QueryKeys(ctx, &pb.QueryKeysRequest{KeysFilters: pbKeysFilters, LimitsAndSorts: pbLimitsAndSorts})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return convertSequencesMatrixFromProto(pbSequencesMatrix.Sequences, sequenceDataTypes)
}

func (c *Client) Bind(ctx context.Context, bindings []types.BoundContract) error {
	pbBindings := make([]*pb.BoundContract, len(bindings))
	for i, b := range bindings {
		pbBindings[i] = &pb.BoundContract{Address: b.Address, Name: b.Name, Pending: b.Pending}
	}
	_, err := c.grpc.Bind(ctx, &pb.BindRequest{Bindings: pbBindings})
	return net.WrapRPCErr(err)
}

var _ pb.ChainReaderServer = (*Server)(nil)

func NewServer(impl types.ChainReader) pb.ChainReaderServer {
	return &Server{impl: impl}
}

type Server struct {
	pb.UnimplementedChainReaderServer
	impl types.ChainReader
}

func (c *Server) GetLatestValue(ctx context.Context, request *pb.GetLatestValueRequest) (*pb.GetLatestValueReply, error) {
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

func (c *Server) QueryKey(ctx context.Context, request *pb.QueryKeyRequest) (*pb.QueryKeyReply, error) {
	queryFilter, err := convertQueryFiltersFromProto(request.KeyFilter.QueryFilter)
	if err != nil {
		return nil, err
	}

	sequenceDataType, err := getContractEncodedTypeByKey(request.KeyFilter.Key, c.impl, false)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	sequences, err := c.impl.QueryKey(ctx, query.KeyFilter{Key: request.KeyFilter.Key, Filter: queryFilter}, limitAndSort, sequenceDataType)
	if err != nil {
		return nil, err
	}

	pbSequences, err := convertSequencesToProto(sequences, sequenceDataType)
	if err != nil {
		return nil, err
	}

	return &pb.QueryKeyReply{Sequences: pbSequences}, nil
}

func (c *Server) QueryKeys(ctx context.Context, request *pb.QueryKeysRequest) (*pb.QueryKeysReply, error) {
	var keysFilters []query.KeyFilter
	for _, pbKeyFilter := range request.KeysFilters {
		queryFilter, err := convertQueryFiltersFromProto(pbKeyFilter.QueryFilter)
		if err != nil {
			return nil, err
		}
		keysFilters = append(keysFilters, query.KeyFilter{Key: pbKeyFilter.Key, Filter: queryFilter})
	}

	var limitsAndSorts []query.LimitAndSort
	for _, pbLimitAndSort := range request.LimitsAndSorts {
		limitAndSort, err := convertLimitAndSortFromProto(pbLimitAndSort)
		if err != nil {
			return nil, err
		}
		limitsAndSorts = append(limitsAndSorts, limitAndSort)
	}

	var sequenceDataTypes []any
	for _, keyFilter := range request.KeysFilters {
		sequenceDataType, err := getContractEncodedTypeByKey(keyFilter.Key, c.impl, false)
		if err != nil {
			return nil, err
		}

		sequenceDataTypes = append(sequenceDataTypes, sequenceDataType)
	}

	sequencesMatrix, err := c.impl.QueryKeys(ctx, keysFilters, limitsAndSorts, sequenceDataTypes)
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

func (c *Server) Bind(ctx context.Context, bindings *pb.BindRequest) (*emptypb.Empty, error) {
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
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_Address{
				Address: &pb.Address{
					Addresses: primitive.Addresses,
				}}
		case *query.ConfirmationsPrimitive:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_Confirmations{
				Confirmations: &pb.Confirmations{
					Confirmations: pb.ConfirmationLevel(primitive.ConfirmationLevel),
				}}
		case *query.BlockPrimitive:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_Block{
				Block: &pb.Block{
					BlockNumber: primitive.Block,
					Operator:    pb.ComparisonOperator(primitive.Operator),
				}}
		case *query.TxHashPrimitive:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_TxHash{
				TxHash: &pb.TxHash{
					TxHash: primitive.TxHash,
				}}
		case *query.TimestampPrimitive:
			pbExpression.GetPrimitive().Comparator = &pb.Primitive_Timestamp{
				Timestamp: &pb.Timestamp{
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

	pbLimitAndSort := &pb.LimitAndSort{
		SortBy: sortByArr,
		Limit:  &pb.Limit{Count: limitAndSort.Limit.Count},
	}

	cursorDefined := limitAndSort.Limit.Cursor != nil
	cursorDirectionDefined := limitAndSort.Limit.CursorDirection != nil
	if cursorDefined && cursorDirectionDefined {
		pbLimitAndSort.Limit.Cursor = limitAndSort.Limit.Cursor
		pbLimitAndSort.Limit.Direction = (*pb.CursorDirection)(limitAndSort.Limit.CursorDirection)
	} else if !cursorDefined && !cursorDirectionDefined {
		return nil, status.Errorf(codes.InvalidArgument, "Limit cursor and cursor direction must both be defined or undefined")
	}

	return pbLimitAndSort, nil
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
			return query.And(expressions...), nil
		}
		return query.Or(expressions...), nil
	case *pb.Expression_Primitive:
		switch primitive := pbEvaluatedExpr.Primitive.Comparator.(type) {
		case *pb.Primitive_Address:
			return query.Address(primitive.Address.Addresses...), nil
		case *pb.Primitive_Confirmations:
			return query.Confirmation(query.ConfirmationLevel(primitive.Confirmations.Confirmations)), nil
		case *pb.Primitive_Block:
			return query.Block(primitive.Block.BlockNumber, query.ComparisonOperator(primitive.Block.Operator)), nil
		case *pb.Primitive_TxHash:
			return query.TxHash(primitive.TxHash.TxHash), nil
		case *pb.Primitive_Timestamp:
			return query.Timestamp(primitive.Timestamp.Timestamp, query.ComparisonOperator(primitive.Timestamp.Operator)), nil
		default:
			return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type")
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type")
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

	limit := limitAndSort.Limit
	cursorDefined := limit.Cursor != nil
	cursorDirectionDefined := limit.Direction != nil
	if cursorDefined && cursorDirectionDefined {
		return query.NewLimitAndSort(query.CursorLimit(*limit.Cursor, (query.CursorDirection)(*limit.Direction), limit.Count)), nil
	} else if !cursorDefined && !cursorDirectionDefined {
		return query.LimitAndSort{}, status.Errorf(codes.InvalidArgument, "Limit cursor and cursor direction must both be defined or undefined")
	}

	return query.NewLimitAndSort(query.CountLimit(limit.Count), sortByArr...), nil
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
