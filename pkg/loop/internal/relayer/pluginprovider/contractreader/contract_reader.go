package contractreader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	types.ContractReader
	serviceClient *goplugin.ServiceClient
	grpc          pb.ContractReaderClient
	encodeWith    EncodingVersion
}

func NewClient(b *net.BrokerExt, cc grpc.ClientConnInterface, opts ...ClientOpt) *Client {
	client := &Client{
		serviceClient: goplugin.NewServiceClient(b, cc),
		grpc:          pb.NewContractReaderClient(cc),
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

func (c *Client) GetLatestValue(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, retVal any) error {
	versionedParams, err := EncodeVersionedBytes(params, c.encodeWith)
	if err != nil {
		return err
	}

	pbConfidence, err := confidenceToProto(confidenceLevel)
	if err != nil {
		return err
	}

	reply, err := c.grpc.GetLatestValue(
		ctx,
		&pb.GetLatestValueRequest{
			ReadIdentifier: readIdentifier,
			Confidence:     pbConfidence,
			Params:         versionedParams,
		},
	)
	if err != nil {
		return net.WrapRPCErr(err)
	}

	return DecodeVersionedBytes(retVal, reply.RetVal)
}

func (c *Client) BatchGetLatestValues(ctx context.Context, request types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	pbRequest, err := convertBatchGetLatestValuesRequestToProto(request, c.encodeWith)
	if err != nil {
		return nil, err
	}

	reply, err := c.grpc.BatchGetLatestValues(ctx, pbRequest)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return parseBatchGetLatestValuesReply(request, reply)
}

func (c *Client) QueryKey(ctx context.Context, contract types.BoundContract, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
	pbQueryFilter, err := convertQueryFilterToProto(filter)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := convertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	reply, err := c.grpc.QueryKey(
		ctx,
		&pb.QueryKeyRequest{
			Contract: &pb.BoundContract{
				Address: contract.Address,
				Name:    contract.Name,
			},
			Filter:       pbQueryFilter,
			LimitAndSort: pbLimitAndSort,
		},
	)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return convertSequencesFromProto(reply.Sequences, sequenceDataType)
}

func (c *Client) Bind(ctx context.Context, bindings []types.BoundContract) error {
	pbBindings := make([]*pb.BoundContract, len(bindings))
	for i, b := range bindings {
		pbBindings[i] = &pb.BoundContract{
			Address: b.Address,
			Name:    b.Name,
		}
	}

	_, err := c.grpc.Bind(ctx, &pb.BindRequest{Bindings: pbBindings})
	return net.WrapRPCErr(err)
}

func (c *Client) Unbind(ctx context.Context, bindings []types.BoundContract) error {
	pbBindings := make([]*pb.BoundContract, len(bindings))
	for i, b := range bindings {
		pbBindings[i] = &pb.BoundContract{
			Address: b.Address,
			Name:    b.Name,
		}
	}

	_, err := c.grpc.Unbind(ctx, &pb.UnbindRequest{Bindings: pbBindings})

	return net.WrapRPCErr(err)
}

func (c *Client) ClientConn() grpc.ClientConnInterface {
	return c.serviceClient.ClientConn()
}

func (c *Client) Close() error {
	return c.serviceClient.Close()
}

func (c *Client) HealthReport() map[string]error {
	return c.serviceClient.HealthReport()
}

func (c *Client) Name() string {
	return c.serviceClient.Name()
}

func (c *Client) Ready() error {
	return c.serviceClient.Ready()
}

func (c *Client) Start(ctx context.Context) error {
	return c.serviceClient.Start(ctx)
}

var _ pb.ContractReaderServer = (*Server)(nil)

type ServerOpt func(*Server)

type Server struct {
	pb.UnimplementedContractReaderServer
	impl       types.ContractReader
	encodeWith EncodingVersion
}

func NewServer(impl types.ContractReader, opts ...ServerOpt) pb.ContractReaderServer {
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
	params, err := getContractEncodedType(request.ReadIdentifier, c.impl, true)
	if err != nil {
		return nil, err
	}

	if err = DecodeVersionedBytes(params, request.Params); err != nil {
		return nil, err
	}

	retVal, err := getContractEncodedType(request.ReadIdentifier, c.impl, false)
	if err != nil {
		return nil, err
	}

	confidenceLevel, err := confidenceFromProto(request.Confidence)
	if err != nil {
		return nil, err
	}

	err = c.impl.GetLatestValue(ctx, request.ReadIdentifier, confidenceLevel, params, retVal)
	if err != nil {
		return nil, err
	}

	encodedRetVal, err := EncodeVersionedBytes(retVal, EncodingVersion(request.Params.Version))
	if err != nil {
		return nil, err
	}

	return &pb.GetLatestValueReply{RetVal: encodedRetVal}, nil
}

func (c *Server) BatchGetLatestValues(ctx context.Context, pbRequest *pb.BatchGetLatestValuesRequest) (*pb.BatchGetLatestValuesReply, error) {
	request, err := convertBatchGetLatestValuesRequestFromProto(pbRequest, c.impl)
	if err != nil {
		return nil, err
	}

	reply, err := c.impl.BatchGetLatestValues(ctx, request)
	if err != nil {
		return nil, err
	}

	return newPbBatchGetLatestValuesReply(reply, c.encodeWith)
}

func (c *Server) QueryKey(ctx context.Context, request *pb.QueryKeyRequest) (*pb.QueryKeyReply, error) {
	queryFilter, err := convertQueryFiltersFromProto(request.Filter)
	if err != nil {
		return nil, err
	}

	contract := convertBoundContractFromProto(request.Contract)

	sequenceDataType, err := getContractEncodedType(contract.ReadIdentifier(queryFilter.Key), c.impl, false)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := convertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	sequences, err := c.impl.QueryKey(ctx, contract, queryFilter, limitAndSort, sequenceDataType)
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
		tBindings[i] = types.BoundContract{
			Address: b.Address,
			Name:    b.Name,
		}
	}

	return &emptypb.Empty{}, c.impl.Bind(ctx, tBindings)
}

func (c *Server) Unbind(ctx context.Context, bindings *pb.UnbindRequest) (*emptypb.Empty, error) {
	tBindings := make([]types.BoundContract, len(bindings.Bindings))
	for i, b := range bindings.Bindings {
		tBindings[i] = types.BoundContract{
			Address: b.Address,
			Name:    b.Name,
		}
	}

	return &emptypb.Empty{}, c.impl.Unbind(ctx, tBindings)
}

func getContractEncodedType(readIdentifier string, possibleTypeProvider any, forEncoding bool) (any, error) {
	if ctp, ok := possibleTypeProvider.(types.ContractTypeProvider); ok {
		return ctp.CreateContractType(readIdentifier, forEncoding)
	}

	return &map[string]any{}, nil
}

func newPbBatchGetLatestValuesReply(result types.BatchGetLatestValuesResult, encodeWith EncodingVersion) (*pb.BatchGetLatestValuesReply, error) {
	resultLookup := make(map[types.BoundContract]*pb.ContractBatchResult)
	results := make([]*pb.ContractBatchResult, 0)

	for binding, nextBatchResult := range result {
		batchResult, exists := resultLookup[binding]
		if !exists {
			batchResult = &pb.ContractBatchResult{
				Contract: convertBoundContractToProto(binding),
				Results:  make([]*pb.BatchReadResult, 0),
			}

			resultLookup[binding] = batchResult
			results = append(results, batchResult)
		}

		for _, batchCall := range nextBatchResult {
			replyErr := ""
			returnVal, err := batchCall.GetResult()
			if err != nil {
				replyErr = err.Error()
			}

			encodedRetVal, err := EncodeVersionedBytes(returnVal, encodeWith)
			if err != nil {
				return nil, err
			}

			batchResult.Results = append(batchResult.Results, &pb.BatchReadResult{
				ReadName:  batchCall.ReadName,
				ReturnVal: encodedRetVal,
				Error:     replyErr,
			})
		}
	}

	return &pb.BatchGetLatestValuesReply{Results: results}, nil
}

func convertBatchGetLatestValuesRequestToProto(request types.BatchGetLatestValuesRequest, encodeWith EncodingVersion) (*pb.BatchGetLatestValuesRequest, error) {
	requests := make([]*pb.ContractBatch, len(request))

	var requestIdx int

	for binding, nextBatch := range request {
		requests[requestIdx] = &pb.ContractBatch{
			Contract: convertBoundContractToProto(binding),
			Reads:    make([]*pb.BatchRead, len(nextBatch)),
		}

		for readIdx, batchCall := range nextBatch {
			versionedParams, err := EncodeVersionedBytes(batchCall.Params, encodeWith)
			if err != nil {
				return nil, err
			}

			requests[requestIdx].Reads[readIdx] = &pb.BatchRead{
				ReadName: batchCall.ReadName,
				Params:   versionedParams,
			}
		}

		requestIdx++
	}

	return &pb.BatchGetLatestValuesRequest{Requests: requests}, nil
}

func convertBoundContractToProto(contract types.BoundContract) *pb.BoundContract {
	return &pb.BoundContract{
		Address: contract.Address,
		Name:    contract.Name,
	}
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
			pbConfidence, err := confidenceToProto(primitive.ConfidenceLevel)
			if err != nil {
				return nil, err
			}
			pbExpression.GetPrimitive().Primitive = &pb.Primitive_Confidence{
				Confidence: pbConfidence,
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

func confidenceToProto(confidenceLevel primitives.ConfidenceLevel) (pb.Confidence, error) {
	switch confidenceLevel {
	case primitives.Finalized:
		return pb.Confidence_Finalized, nil
	case primitives.Unconfirmed:
		return pb.Confidence_Unconfirmed, nil
	default:
		return -1, fmt.Errorf("invalid confidence level %s", confidenceLevel)
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
				Height:    sequence.Height,
				Hash:      sequence.Hash,
				Timestamp: sequence.Timestamp,
			},
			Data: versionedSequenceDataType,
		}
		pbSequences = append(pbSequences, pbSequence)
	}
	return pbSequences, nil
}

func parseBatchGetLatestValuesReply(request types.BatchGetLatestValuesRequest, reply *pb.BatchGetLatestValuesReply) (types.BatchGetLatestValuesResult, error) {
	if reply == nil {
		return nil, fmt.Errorf("received nil reply from grpc BatchGetLatestValues")
	}

	result := make(types.BatchGetLatestValuesResult)
	for _, contractBatch := range reply.Results {
		binding := convertBoundContractFromProto(contractBatch.Contract)
		result[binding] = make([]types.BatchReadResult, len(contractBatch.Results))
		resultsContractBatch := contractBatch.Results

		requestContractBatch, ok := request[binding]
		if !ok {
			return nil, fmt.Errorf("received unexpected contract name %s from grpc BatchGetLatestValues reply", binding)
		}

		if len(requestContractBatch) != len(resultsContractBatch) {
			return nil, fmt.Errorf("request and results length for contract %s are mismatched %d vs %d", binding, len(requestContractBatch), len(resultsContractBatch))
		}

		for i := 0; i < len(resultsContractBatch); i++ {
			// type lives in the request, so we can use it for result
			res, req := resultsContractBatch[i], requestContractBatch[i]
			if err := DecodeVersionedBytes(req.ReturnVal, res.ReturnVal); err != nil {
				return nil, err
			}

			var err error
			if res.Error != "" {
				err = errors.New(res.Error)
			}

			brr := types.BatchReadResult{ReadName: res.ReadName}
			brr.SetResult(req.ReturnVal, err)
			result[binding][i] = brr
		}
	}

	return result, nil
}

func convertBatchGetLatestValuesRequestFromProto(pbRequest *pb.BatchGetLatestValuesRequest, impl types.ContractReader) (types.BatchGetLatestValuesRequest, error) {
	if pbRequest == nil {
		return nil, fmt.Errorf("received nil request from grpc BatchGetLatestValues")
	}

	request := make(types.BatchGetLatestValuesRequest)
	for _, pbBatch := range pbRequest.Requests {
		binding := convertBoundContractFromProto(pbBatch.Contract)

		if _, ok := request[binding]; !ok {
			request[binding] = make([]types.BatchRead, len(pbBatch.Reads))
		}

		for idx, pbReadCall := range pbBatch.Reads {
			call := types.BatchRead{ReadName: pbReadCall.ReadName}
			params, err := getContractEncodedType(binding.ReadIdentifier(pbReadCall.ReadName), impl, true)
			if err != nil {
				return nil, err
			}

			if err = DecodeVersionedBytes(params, pbReadCall.Params); err != nil {
				return nil, err
			}

			retVal, err := getContractEncodedType(binding.ReadIdentifier(call.ReadName), impl, false)
			if err != nil {
				return nil, err
			}

			call.Params = params
			call.ReturnVal = retVal
			request[binding][idx] = call
		}
	}

	return request, nil
}

func convertBoundContractFromProto(contract *pb.BoundContract) types.BoundContract {
	return types.BoundContract{
		Address: contract.Address,
		Name:    contract.Name,
	}
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
			confidence, err := confidenceFromProto(primitive.Confidence)
			return query.Confidence(confidence), err
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

func confidenceFromProto(pbConfidence pb.Confidence) (primitives.ConfidenceLevel, error) {
	switch pbConfidence {
	case pb.Confidence_Finalized:
		return primitives.Finalized, nil
	case pb.Confidence_Unconfirmed:
		return primitives.Unconfirmed, nil
	default:
		return "", fmt.Errorf("invalid pb confidence level: %d", pbConfidence)
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
				Height:    pbSequences[idx].Head.Height,
				Hash:      pbSequences[idx].Head.Hash,
				Timestamp: pbSequences[idx].Head.Timestamp,
			},
			Data: cpy,
		}
	}

	return sequences, nil
}

func RegisterContractReaderService(s *grpc.Server, contractReader types.ContractReader) {
	pb.RegisterServiceServer(s, &goplugin.ServiceServer{Srv: contractReader})
}
