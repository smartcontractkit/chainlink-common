package chainreader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	jsonv2 "github.com/go-json-experiment/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

var _ types.ContractReader = (*Client)(nil)

type EncodingVersion uint32

func (v EncodingVersion) Uint32() uint32 {
	return uint32(v)
}

// enum of all known encoding formats for versioned data.
const (
	JSONEncodingVersion1 EncodingVersion = iota
	JSONEncodingVersion2
	CBOREncodingVersion
)

const DefaultEncodingVersion = CBOREncodingVersion

type ClientOpt func(*Client)

type Client struct {
	*goplugin.ServiceClient
	grpc       pb.ChainReaderClient
	encodeWith EncodingVersion
}

func NewClient(b *net.BrokerExt, cc grpc.ClientConnInterface, opts ...ClientOpt) *Client {
	client := &Client{
		ServiceClient: goplugin.NewServiceClient(b, cc),
		grpc:          pb.NewChainReaderClient(cc),
		encodeWith:    DefaultEncodingVersion,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func WithClientEncoding(version EncodingVersion) ClientOpt {
	return func(client *Client) {
		client.encodeWith = version
	}
}

func EncodeVersionedBytes(data any, version EncodingVersion) (*pb.VersionedBytes, error) {
	var bytes []byte
	var err error

	switch version {
	case JSONEncodingVersion1:
		bytes, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	case JSONEncodingVersion2:
		bytes, err = jsonv2.Marshal(data, jsonv2.StringifyNumbers(true))
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

	return &pb.VersionedBytes{Version: version.Uint32(), Data: bytes}, nil
}

func DecodeVersionedBytes(res any, vData *pb.VersionedBytes) error {
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
	default:
		return fmt.Errorf("unsupported encoding version %d for versionedData %v", vData.Version, vData.Data)
	}

	if err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}

	return nil
}

func (c *Client) GetLatestValue(ctx context.Context, contractName, method string, params, retVal any) error {
	versionedParams, err := EncodeVersionedBytes(params, c.encodeWith)
	if err != nil {
		return err
	}

	reply, err := c.grpc.GetLatestValue(ctx, &pb.GetLatestValueRequest{ContractName: contractName, Method: method, Params: versionedParams})
	if err != nil {
		return net.WrapRPCErr(err)
	}

	return DecodeVersionedBytes(retVal, reply.RetVal)
}

func (c *Client) BatchGetLatestValue(ctx context.Context, request types.BatchGetLatestValueRequest) (types.BatchGetLatestValueResult, error) {
	pbRequest, err := convertBatchGetLatestValueRequestToProto(request, c.encodeWith)
	if err != nil {
		return nil, err
	}

	reply, err := c.grpc.BatchGetLatestValue(ctx, pbRequest)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return parseBatchGetLatestValueReply(request, reply)
}

func (c *Client) QueryKey(ctx context.Context, contractName string, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(filter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	reply, err := c.grpc.QueryKey(ctx, &pb.QueryKeyRequest{ContractName: contractName, Filter: pbQueryFilter, LimitAndSort: pbLimitAndSort})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return convertSequencesFromProto(reply.Sequences, sequenceDataType)
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

type ServerOpt func(*Server)

type Server struct {
	pb.UnimplementedChainReaderServer
	impl       types.ContractReader
	encodeWith EncodingVersion
}

func NewServer(impl types.ContractReader, opts ...ServerOpt) pb.ChainReaderServer {
	server := &Server{
		impl:       impl,
		encodeWith: DefaultEncodingVersion,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func WithServerEncoding(version EncodingVersion) ServerOpt {
	return func(server *Server) {
		server.encodeWith = version
	}
}

func (c *Server) GetLatestValue(ctx context.Context, request *pb.GetLatestValueRequest) (*pb.GetLatestValueReply, error) {
	params, err := getContractEncodedType(request.ContractName, request.Method, c.impl, true)
	if err != nil {
		return nil, err
	}

	if err = DecodeVersionedBytes(params, request.Params); err != nil {
		return nil, err
	}

	retVal, err := getContractEncodedType(request.ContractName, request.Method, c.impl, false)
	if err != nil {
		return nil, err
	}

	err = c.impl.GetLatestValue(ctx, request.ContractName, request.Method, params, retVal)
	if err != nil {
		return nil, err
	}

	encodedRetVal, err := EncodeVersionedBytes(retVal, EncodingVersion(request.Params.Version))
	if err != nil {
		return nil, err
	}

	return &pb.GetLatestValueReply{RetVal: encodedRetVal}, nil
}

func (c *Server) BatchGetLatestValue(ctx context.Context, pbRequest *pb.BatchGetLatestValueRequest) (*pb.BatchGetLatestValueReply, error) {
	request, err := convertBatchGetLatestValueRequestFromProto(pbRequest, c.impl)
	if err != nil {
		return nil, err
	}

	reply, err := c.impl.BatchGetLatestValue(ctx, request)
	if err != nil {
		return nil, err
	}

	return newPbBatchGetLatestValueReply(reply, c.encodeWith)
}

func (c *Server) QueryKey(ctx context.Context, request *pb.QueryKeyRequest) (*pb.QueryKeyReply, error) {
	queryFilter, err := convertQueryFiltersFromProto(request.Filter)
	if err != nil {
		return nil, err
	}

	sequenceDataType, err := getContractEncodedType(request.ContractName, queryFilter.Key, c.impl, false)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	sequences, err := c.impl.QueryKey(ctx, request.ContractName, queryFilter, limitAndSort, sequenceDataType)
	if err != nil {
		return nil, err
	}

	pbSequences, err := convertSequencesToProto(sequences, c.encodeWith)
	if err != nil {
		return nil, err
	}

	return &pb.QueryKeyReply{Sequences: pbSequences}, nil
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

func newPbBatchGetLatestValueReply(result types.BatchGetLatestValueResult, encodeWith EncodingVersion) (*pb.BatchGetLatestValueReply, error) {
	var pbBatchGetLatestValueReply = &pb.BatchGetLatestValueReply{Results: make(map[string]*pb.ContractBatchResult)}
	for contractName, contractBatch := range result {
		pbBatchGetLatestValueReply.Results[contractName] = &pb.ContractBatchResult{Results: []*pb.BatchReadResult{}}
		for _, batchCall := range contractBatch {
			encodedRetVal, err := EncodeVersionedBytes(batchCall.ReturnValue, encodeWith)
			if err != nil {
				return nil, err
			}
			pbBatchReadResult := &pb.BatchReadResult{ReadName: batchCall.ReadName, ReturnVal: encodedRetVal}
			pbBatchGetLatestValueReply.Results[contractName].Results = append(pbBatchGetLatestValueReply.Results[contractName].Results, pbBatchReadResult)
		}
	}
	return pbBatchGetLatestValueReply, nil
}

func convertBatchGetLatestValueRequestToProto(request types.BatchGetLatestValueRequest, encodeWith EncodingVersion) (*pb.BatchGetLatestValueRequest, error) {
	pbRequest := &pb.BatchGetLatestValueRequest{Requests: make(map[string]*pb.ContractBatch)}
	for contractName, contractBatch := range request {
		if _, ok := pbRequest.Requests[contractName]; !ok {
			pbContractBatch := &pb.ContractBatch{Reads: []*pb.BatchRead{}}
			pbRequest.Requests[contractName] = pbContractBatch
		}

		for _, batchCall := range contractBatch {
			versionedParams, err := EncodeVersionedBytes(batchCall.Params, encodeWith)
			if err != nil {
				return nil, err
			}

			pbRequest.Requests[contractName].Reads = append(
				pbRequest.Requests[contractName].Reads,
				&pb.BatchRead{
					ReadName: batchCall.ReadName,
					Params:   versionedParams,
				},
			)
		}
	}
	return pbRequest, nil
}

func convertQueryFilterToProto(filter query.KeyFilter) (*pb.QueryKeyFilter, error) {
	pbQueryFilter := &pb.QueryKeyFilter{Key: filter.Key}
	for _, expression := range filter.Expressions {
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
		case *primitives.Comparator:
			var pbValueComparators []*pb.ValueComparator
			for _, valueComparator := range primitive.ValueComparators {
				pbValueComparators = append(pbValueComparators, &pb.ValueComparator{Value: valueComparator.Value, Operator: pb.ComparisonOperator(valueComparator.Operator)})
			}
			pbExpression.GetPrimitive().Primitive = &pb.Primitive_Comparator{
				Comparator: &pb.Comparator{
					Name:             primitive.Name,
					ValueComparators: pbValueComparators,
				}}

		case *primitives.Block:
			pbExpression.GetPrimitive().Primitive = &pb.Primitive_Block{
				Block: &pb.Block{
					BlockNumber: primitive.Block,
					Operator:    pb.ComparisonOperator(primitive.Operator),
				}}
		case *primitives.Confidence:
			pbExpression.GetPrimitive().Primitive = &pb.Primitive_Confidence{
				Confidence: confidenceToProto(primitive.ConfidenceLevel),
			}
		case *primitives.Timestamp:
			pbExpression.GetPrimitive().Primitive = &pb.Primitive_Timestamp{
				Timestamp: &pb.Timestamp{
					Timestamp: primitive.Timestamp,
					Operator:  pb.ComparisonOperator(primitive.Operator),
				}}
		case *primitives.TxHash:
			pbExpression.GetPrimitive().Primitive = &pb.Primitive_TxHash{
				TxHash: &pb.TxHash{
					TxHash: primitive.TxHash,
				}}
		default:
			return nil, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
		}
		return pbExpression, nil
	}

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

	return pbExpression, nil
}

func confidenceToProto(level primitives.ConfidenceLevel) pb.Confidence {
	switch level {
	case primitives.Finalized:
		return pb.Confidence_Finalized
	case primitives.Unconfirmed:
		return pb.Confidence_Unconfirmed
	default:
		panic("invalid confidence level")
	}
}

func convertLimitAndSortToProto(limitAndSort query.LimitAndSort) (*pb.LimitAndSort, error) {
	sortByArr := make([]*pb.SortBy, len(limitAndSort.SortBy))

	for idx, sortBy := range limitAndSort.SortBy {
		var tp pb.SortType

		switch sort := sortBy.(type) {
		case *query.SortByBlock:
			tp = pb.SortType_SortBlock
		case *query.SortByTimestamp:
			tp = pb.SortType_SortTimestamp
		case *query.SortBySequence:
			tp = pb.SortType_SortSequence
		default:
			return &pb.LimitAndSort{}, status.Errorf(codes.InvalidArgument, "Unknown sort by type: %T", sort)
		}

		sortByArr[idx] = &pb.SortBy{
			SortType:  tp,
			Direction: pb.SortDirection(sortBy.GetDirection()),
		}
	}

	pbLimitAndSort := &pb.LimitAndSort{
		SortBy: sortByArr,
		Limit:  &pb.Limit{Count: limitAndSort.Limit.Count},
	}

	cursorDefined := limitAndSort.Limit.Cursor != ""
	cursorDirectionDefined := limitAndSort.Limit.CursorDirection != 0

	if limitAndSort.HasCursorLimit() {
		pbLimitAndSort.Limit.Cursor = &limitAndSort.Limit.Cursor
		pbLimitAndSort.Limit.Direction = (*pb.CursorDirection)(&limitAndSort.Limit.CursorDirection)
	} else if (!cursorDefined && cursorDirectionDefined) || (cursorDefined && !cursorDirectionDefined) {
		return nil, status.Errorf(codes.InvalidArgument, "Limit cursor and cursor direction must both be defined or undefined")
	}

	return pbLimitAndSort, nil
}

func convertSequencesToProto(sequences []types.Sequence, version EncodingVersion) ([]*pb.Sequence, error) {
	var pbSequences []*pb.Sequence
	for _, sequence := range sequences {
		versionedSequenceDataType, err := EncodeVersionedBytes(sequence.Data, version)
		if err != nil {
			return nil, err
		}
		pbSequence := &pb.Sequence{
			SequenceCursor: sequence.Cursor,
			Head: &pb.Head{
				Identifier: sequence.Identifier,
				Hash:       sequence.Hash,
				Timestamp:  sequence.Timestamp,
			},
			Data: versionedSequenceDataType,
		}
		pbSequences = append(pbSequences, pbSequence)
	}
	return pbSequences, nil
}

func parseBatchGetLatestValueReply(request types.BatchGetLatestValueRequest, reply *pb.BatchGetLatestValueReply) (types.BatchGetLatestValueResult, error) {
	if reply == nil {
		return nil, fmt.Errorf("received nil reply from grpc BatchGetLatestValue")
	}

	result := make(types.BatchGetLatestValueResult)
	for contractName, contractBatch := range reply.Results {
		result[contractName] = make([]types.BatchReadResult, len(contractBatch.Results))
		resultsContractBatch := contractBatch.Results
		requestContractBatch, ok := request[contractName]
		if !ok {
			return nil, fmt.Errorf("received unexpected contract name %s from grpc BatchGetLatestValue reply", contractName)
		}

		if len(requestContractBatch) != len(resultsContractBatch) {
			return nil, fmt.Errorf("request and results length for contract %s are mismatched %d vs %d", contractName, len(requestContractBatch), len(resultsContractBatch))
		}

		for i := 0; i < len(resultsContractBatch); i++ {
			// type lives in the request, so we can use it for result
			res, req := resultsContractBatch[i], requestContractBatch[i]
			if err := DecodeVersionedBytes(req.ReturnVal, res.ReturnVal); err != nil {
				return nil, err
			}
			var err error
			if res.Error != "" {
				err = fmt.Errorf(res.Error)
			}
			result[contractName][i] = types.BatchReadResult{ReadName: res.ReadName, ReturnValue: req.ReturnVal, Err: err}
		}
	}
	return result, nil
}

func convertBatchGetLatestValueRequestFromProto(pbRequest *pb.BatchGetLatestValueRequest, impl types.ContractReader) (types.BatchGetLatestValueRequest, error) {
	if pbRequest == nil {
		return nil, fmt.Errorf("received nil request from grpc BatchGetLatestValue")
	}

	request := make(types.BatchGetLatestValueRequest)
	for pbContractName, pbContractBatch := range pbRequest.Requests {
		if _, ok := request[pbContractName]; !ok {
			request[pbContractName] = []types.BatchRead{}
		}

		for _, pbCall := range pbContractBatch.Reads {
			call := types.BatchRead{ReadName: pbCall.ReadName}
			params, err := getContractEncodedType(pbContractName, pbCall.ReadName, impl, true)
			if err != nil {
				return nil, err
			}

			if err = DecodeVersionedBytes(params, pbCall.Params); err != nil {
				return nil, err
			}

			retVal, err := getContractEncodedType(pbContractName, call.ReadName, impl, false)
			if err != nil {
				return nil, err
			}

			call.Params = params
			call.ReturnVal = retVal
			request[pbContractName] = append(request[pbContractName], call)
		}
	}
	return request, nil
}

func convertQueryFiltersFromProto(pbQueryFilters *pb.QueryKeyFilter) (query.KeyFilter, error) {
	queryFilter := query.KeyFilter{Key: pbQueryFilters.Key}
	for _, pbQueryFilter := range pbQueryFilters.Expression {
		expression, err := convertExpressionFromProto(pbQueryFilter)
		if err != nil {
			return query.KeyFilter{}, err
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
		switch primitive := pbEvaluatedExpr.Primitive.GetPrimitive().(type) {
		case *pb.Primitive_Comparator:
			var valueComparators []primitives.ValueComparator
			for _, pbValueComparator := range primitive.Comparator.ValueComparators {
				valueComparators = append(valueComparators, primitives.ValueComparator{Value: pbValueComparator.Value, Operator: primitives.ComparisonOperator(pbValueComparator.Operator)})
			}
			return query.Comparator(primitive.Comparator.Name, valueComparators...), nil
		case *pb.Primitive_Confidence:
			return query.Confidence(confidenceFromProto(primitive.Confidence)), nil
		case *pb.Primitive_Block:
			return query.Block(primitive.Block.BlockNumber, primitives.ComparisonOperator(primitive.Block.Operator)), nil
		case *pb.Primitive_TxHash:
			return query.TxHash(primitive.TxHash.TxHash), nil
		case *pb.Primitive_Timestamp:
			return query.Timestamp(primitive.Timestamp.Timestamp, primitives.ComparisonOperator(primitive.Timestamp.Operator)), nil
		default:
			return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type: %T", pbEvaluatedExpr)
	}
}

func confidenceFromProto(conf pb.Confidence) primitives.ConfidenceLevel {
	switch conf {
	case pb.Confidence_Finalized:
		return primitives.Finalized
	case pb.Confidence_Unconfirmed:
		return primitives.Unconfirmed
	default:
		panic("invalid confidence level")
	}
}

func convertLimitAndSortFromProto(limitAndSort *pb.LimitAndSort) (query.LimitAndSort, error) {
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

func convertSequencesFromProto(pbSequences []*pb.Sequence, sequenceDataType any) ([]types.Sequence, error) {
	seqTypeOf := reflect.TypeOf(sequenceDataType)

	// get the non-pointer data type for the sequence data
	nonPointerType := seqTypeOf
	if seqTypeOf.Kind() == reflect.Pointer {
		nonPointerType = seqTypeOf.Elem()
	}

	if nonPointerType.Kind() == reflect.Pointer {
		return nil, fmt.Errorf("%w: sequenceDataType does not support pointers to pointers", types.ErrInvalidType)
	}

	sequences := make([]types.Sequence, len(pbSequences))

	for idx, pbSequence := range pbSequences {
		cpy := reflect.New(nonPointerType).Interface()
		if err := DecodeVersionedBytes(cpy, pbSequence.Data); err != nil {
			return nil, err
		}

		// match provided data type either as pointer or non-pointer
		if seqTypeOf.Kind() != reflect.Pointer {
			cpy = reflect.Indirect(reflect.ValueOf(cpy)).Interface()
		}

		sequences[idx] = types.Sequence{
			Cursor: pbSequences[idx].SequenceCursor,
			Head: types.Head{
				Identifier: pbSequences[idx].Head.Identifier,
				Hash:       pbSequences[idx].Head.Hash,
				Timestamp:  pbSequences[idx].Head.Timestamp,
			},
			Data: cpy,
		}
	}

	return sequences, nil
}

func RegisterContractReaderService(s *grpc.Server, contractReader types.ContractReader) {
	pb.RegisterServiceServer(s, &goplugin.ServiceServer{Srv: contractReader})
}
