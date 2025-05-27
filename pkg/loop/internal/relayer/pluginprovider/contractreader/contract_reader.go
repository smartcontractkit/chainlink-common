package contractreader

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	codecpb "github.com/smartcontractkit/chainlink-common/pkg/internal/codec"
	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

var _ types.ContractReader = (*Client)(nil)

type serviceClient interface {
	ClientConn() grpc.ClientConnInterface
	Close() error
	HealthReport() map[string]error
	Name() string
	Ready() error
	Start(ctx context.Context) error
}

type ClientOpt func(*Client)

type Client struct {
	types.UnimplementedContractReader
	serviceClient serviceClient
	grpc          pb.ContractReaderClient
	encodeWith    codecpb.EncodingVersion
}

func NewClient(serviceClient serviceClient, grpc pb.ContractReaderClient, opts ...ClientOpt) *Client {
	client := &Client{
		serviceClient: serviceClient,
		grpc:          grpc,
		encodeWith:    codecpb.DefaultEncodingVersion,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func WithClientEncoding(version codecpb.EncodingVersion) ClientOpt {
	return func(client *Client) {
		client.encodeWith = version
	}
}

func (c *Client) GetLatestValue(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, retVal any) error {
	_, asValueType := retVal.(*values.Value)

	versionedParams, err := codecpb.EncodeVersionedBytes(params, c.encodeWith)
	if err != nil {
		return err
	}

	pbConfidence, err := chaincommonpb.ConvertConfidenceToProto(confidenceLevel)
	if err != nil {
		return err
	}

	reply, err := c.grpc.GetLatestValue(
		ctx,
		&pb.GetLatestValueRequest{
			ReadIdentifier: readIdentifier,
			Confidence:     pbConfidence,
			Params:         versionedParams,
			AsValueType:    asValueType,
		},
	)
	if err != nil {
		return net.WrapRPCErr(err)
	}

	return codecpb.DecodeVersionedBytes(retVal, reply.RetVal)
}

func (c *Client) GetLatestValueWithHeadData(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, retVal any) (*types.Head, error) {
	_, asValueType := retVal.(*values.Value)

	versionedParams, err := codecpb.EncodeVersionedBytes(params, c.encodeWith)
	if err != nil {
		return nil, err
	}

	pbConfidence, err := chaincommonpb.ConvertConfidenceToProto(confidenceLevel)
	if err != nil {
		return nil, err
	}

	reply, err := c.grpc.GetLatestValueWithHeadData(
		ctx,
		&pb.GetLatestValueRequest{
			ReadIdentifier: readIdentifier,
			Confidence:     pbConfidence,
			Params:         versionedParams,
			AsValueType:    asValueType,
		},
	)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	var headData *types.Head
	if reply.HeadData != nil {
		headData = &types.Head{
			Height:    reply.HeadData.Height,
			Hash:      reply.HeadData.Hash,
			Timestamp: reply.HeadData.Timestamp,
		}
	}

	return headData, codecpb.DecodeVersionedBytes(retVal, reply.RetVal)
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
	_, asValueType := sequenceDataType.(*values.Value)

	pbQueryFilter, err := convertQueryFilterToProto(filter, c.encodeWith)
	if err != nil {
		return nil, err
	}

	pbLimitAndSort, err := chaincommonpb.ConvertLimitAndSortToProto(limitAndSort)
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
			AsValueType:  asValueType,
		},
	)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return convertSequencesFromProto(reply.Sequences, sequenceDataType)
}

func (c *Client) QueryKeys(ctx context.Context, keyQueries []types.ContractKeyFilter, limitAndSort query.LimitAndSort) (iter.Seq2[string, types.Sequence], error) {
	var filters []*pb.ContractKeyFilter
	for _, keyQuery := range keyQueries {
		_, asValueType := keyQuery.SequenceDataType.(*values.Value)
		contract := convertBoundContractToProto(keyQuery.Contract)

		pbQueryFilter, err := convertQueryFilterToProto(keyQuery.KeyFilter, c.encodeWith)
		if err != nil {
			return nil, err
		}

		filters = append(filters, &pb.ContractKeyFilter{
			Contract:    contract,
			Filter:      pbQueryFilter,
			AsValueType: asValueType,
		})
	}

	pbLimitAndSort, err := chaincommonpb.ConvertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	reply, err := c.grpc.QueryKeys(
		ctx,
		&pb.QueryKeysRequest{
			Filters:      filters,
			LimitAndSort: pbLimitAndSort,
		},
	)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return convertSequencesWithKeyFromProto(reply.Sequences, keyQueries)
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
	encodeWith codecpb.EncodingVersion
}

func NewServer(impl types.ContractReader, opts ...ServerOpt) pb.ContractReaderServer {
	server := &Server{
		impl:       impl,
		encodeWith: codecpb.DefaultEncodingVersion,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func WithServerEncoding(version codecpb.EncodingVersion) ServerOpt {
	return func(server *Server) {
		server.encodeWith = version
	}
}

func (c *Server) GetLatestValue(ctx context.Context, request *pb.GetLatestValueRequest) (*pb.GetLatestValueReply, error) {
	params, err := getContractEncodedType(request.ReadIdentifier, c.impl, true)
	if err != nil {
		return nil, err
	}

	if err = codecpb.DecodeVersionedBytes(params, request.Params); err != nil {
		return nil, err
	}

	retVal, err := getContractEncodedType(request.ReadIdentifier, c.impl, false)
	if err != nil {
		return nil, err
	}

	confidenceLevel, err := chaincommonpb.ConvertConfidenceFromProto(request.Confidence)
	if err != nil {
		return nil, err
	}

	err = c.impl.GetLatestValue(ctx, request.ReadIdentifier, confidenceLevel, params, retVal)
	if err != nil {
		return nil, err
	}

	encodeWith := codecpb.EncodingVersion(request.Params.Version)
	if request.AsValueType {
		encodeWith = codecpb.ValuesEncodingVersion
	}

	versionedBytes, err := codecpb.EncodeVersionedBytes(retVal, encodeWith)
	if err != nil {
		return nil, err
	}

	return &pb.GetLatestValueReply{RetVal: versionedBytes}, nil
}

func (c *Server) GetLatestValueWithHeadData(ctx context.Context, request *pb.GetLatestValueRequest) (*pb.GetLatestValueWithHeadDataReply, error) {
	params, err := getContractEncodedType(request.ReadIdentifier, c.impl, true)
	if err != nil {
		return nil, err
	}

	if err = codecpb.DecodeVersionedBytes(params, request.Params); err != nil {
		return nil, err
	}

	retVal, err := getContractEncodedType(request.ReadIdentifier, c.impl, false)
	if err != nil {
		return nil, err
	}

	confidenceLevel, err := chaincommonpb.ConvertConfidenceFromProto(request.Confidence)
	if err != nil {
		return nil, err
	}

	headData, err := c.impl.GetLatestValueWithHeadData(ctx, request.ReadIdentifier, confidenceLevel, params, retVal)
	if err != nil {
		return nil, err
	}

	encodeWith := codecpb.EncodingVersion(request.Params.Version)
	if request.AsValueType {
		encodeWith = codecpb.ValuesEncodingVersion
	}

	versionedBytes, err := codecpb.EncodeVersionedBytes(retVal, encodeWith)
	if err != nil {
		return nil, err
	}

	var headDataProto *pb.Head
	if headData != nil {
		headDataProto = &pb.Head{
			Height:    headData.Height,
			Hash:      headData.Hash,
			Timestamp: headData.Timestamp,
		}
	}

	return &pb.GetLatestValueWithHeadDataReply{RetVal: versionedBytes, HeadData: headDataProto}, nil
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
	contract := convertBoundContractFromProto(request.Contract)

	queryFilter, err := convertQueryFiltersFromProto(request.Filter, contract, c.impl)
	if err != nil {
		return nil, err
	}

	sequenceDataType, err := getContractEncodedType(contract.ReadIdentifier(queryFilter.Key), c.impl, false)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := chaincommonpb.ConvertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	sequences, err := c.impl.QueryKey(ctx, contract, queryFilter, limitAndSort, sequenceDataType)
	if err != nil {
		return nil, err
	}

	encodeWith := c.encodeWith
	if request.AsValueType {
		encodeWith = codecpb.ValuesEncodingVersion
	}

	pbSequences, err := convertSequencesToVersionedBytesProto(sequences, encodeWith)
	if err != nil {
		return nil, err
	}

	return &pb.QueryKeyReply{Sequences: pbSequences}, nil
}

func (c *Server) QueryKeys(ctx context.Context, request *pb.QueryKeysRequest) (*pb.QueryKeysReply, error) {
	var filters []types.ContractKeyFilter
	for _, keyQuery := range request.Filters {
		contract := convertBoundContractFromProto(keyQuery.Contract)

		queryFilter, err := convertQueryFiltersFromProto(keyQuery.Filter, contract, c.impl)
		if err != nil {
			return nil, err
		}

		sequenceDataType, err := getContractEncodedType(contract.ReadIdentifier(queryFilter.Key), c.impl, false)
		if err != nil {
			return nil, err
		}

		filters = append(filters, types.ContractKeyFilter{
			Contract:         contract,
			KeyFilter:        queryFilter,
			SequenceDataType: sequenceDataType,
		})
	}

	limitAndSort, err := chaincommonpb.ConvertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	sequences, err := c.impl.QueryKeys(ctx, filters, limitAndSort)
	if err != nil {
		return nil, err
	}

	pbSequences, err := convertSequencesWithKeyToVersionedBytesProto(sequences, request.Filters, c.encodeWith)
	if err != nil {
		return nil, err
	}

	return &pb.QueryKeysReply{Sequences: pbSequences}, nil
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

func newPbBatchGetLatestValuesReply(result types.BatchGetLatestValuesResult, encodeWith codecpb.EncodingVersion) (*pb.BatchGetLatestValuesReply, error) {
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

			encodedRetVal, err := codecpb.EncodeVersionedBytes(returnVal, encodeWith)
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

func convertBatchGetLatestValuesRequestToProto(request types.BatchGetLatestValuesRequest, encodeWith codecpb.EncodingVersion) (*pb.BatchGetLatestValuesRequest, error) {
	requests := make([]*pb.ContractBatch, len(request))

	var requestIdx int

	for binding, nextBatch := range request {
		requests[requestIdx] = &pb.ContractBatch{
			Contract: convertBoundContractToProto(binding),
			Reads:    make([]*pb.BatchRead, len(nextBatch)),
		}

		for readIdx, batchCall := range nextBatch {
			versionedParams, err := codecpb.EncodeVersionedBytes(batchCall.Params, encodeWith)
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

func convertQueryFilterToProto(filter query.KeyFilter, encodeWith codecpb.EncodingVersion) (*pb.QueryKeyFilter, error) {
	pbQueryFilter := &pb.QueryKeyFilter{Key: filter.Key}
	for _, expression := range filter.Expressions {
		pbExpression, err := chaincommonpb.ConvertExpressionToProto(expression, func(value any) (*codecpb.VersionedBytes, error) {
			return codecpb.EncodeVersionedBytes(value, encodeWith)
		})
		if err != nil {
			return nil, err
		}

		pbQueryFilter.Expression = append(pbQueryFilter.Expression, pbExpression)
	}

	return pbQueryFilter, nil
}

func convertSequencesToVersionedBytesProto(sequences []types.Sequence, version codecpb.EncodingVersion) ([]*pb.Sequence, error) {
	var pbSequences []*pb.Sequence
	for _, sequence := range sequences {
		versionedSequenceDataType, err := codecpb.EncodeVersionedBytes(sequence.Data, version)
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

func convertSequencesWithKeyToVersionedBytesProto(sequences iter.Seq2[string, types.Sequence], filters []*pb.ContractKeyFilter, encodeWith codecpb.EncodingVersion) ([]*pb.SequenceWithKey, error) {
	keyToEncodingVersion := make(map[string]codecpb.EncodingVersion)
	for _, filter := range filters {
		if filter.AsValueType {
			keyToEncodingVersion[filter.Filter.Key] = codecpb.ValuesEncodingVersion
		} else {
			keyToEncodingVersion[filter.Filter.Key] = encodeWith
		}
	}

	var pbSequences []*pb.SequenceWithKey
	for key, sequence := range sequences {
		version, ok := keyToEncodingVersion[key]
		if !ok {
			return nil, fmt.Errorf("missing encoding version for key %s", key)
		}

		versionedSequenceDataType, err := codecpb.EncodeVersionedBytes(sequence.Data, version)
		if err != nil {
			return nil, err
		}
		pbSequence := &pb.SequenceWithKey{
			Key:            key,
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
			if err := codecpb.DecodeVersionedBytes(req.ReturnVal, res.ReturnVal); err != nil {
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

			if err = codecpb.DecodeVersionedBytes(params, pbReadCall.Params); err != nil {
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

func convertQueryFiltersFromProto(pbQueryFilters *pb.QueryKeyFilter, contract types.BoundContract, impl types.ContractReader) (query.KeyFilter, error) {
	queryFilter := query.KeyFilter{Key: pbQueryFilters.Key}
	for _, pbExpr := range pbQueryFilters.Expression {
		expression, err := chaincommonpb.ConvertExpressionFromProto(pbExpr, func(comparatorName string, forEncoding bool) (any, error) {
			if ctp, ok := impl.(types.ContractTypeProvider); ok {
				return ctp.CreateContractType(contract.ReadIdentifier(pbQueryFilters.Key+"."+comparatorName), forEncoding)
			}
			return &map[string]any{}, nil
		})
		if err != nil {
			return query.KeyFilter{}, err
		}
		queryFilter.Expressions = append(queryFilter.Expressions, expression)
	}
	return queryFilter, nil
}

func convertSequencesFromProto(pbSequences []*pb.Sequence, sequenceDataType any) ([]types.Sequence, error) {
	sequences := make([]types.Sequence, len(pbSequences))

	seqTypeOf, nonPointerType, err := getSequenceTypeInformation(sequenceDataType)
	if err != nil {
		return nil, err
	}

	for idx, pbSequence := range pbSequences {
		cpy := reflect.New(nonPointerType).Interface()
		if err = codecpb.DecodeVersionedBytes(cpy, pbSequence.Data); err != nil {
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

func getSequenceTypeInformation(sequenceDataType any) (reflect.Type, reflect.Type, error) {
	seqTypeOf := reflect.TypeOf(sequenceDataType)

	// get the non-pointer data type for the sequence data
	nonPointerType := seqTypeOf
	if seqTypeOf.Kind() == reflect.Pointer {
		nonPointerType = seqTypeOf.Elem()
	}

	if nonPointerType.Kind() == reflect.Pointer {
		return nil, nil, fmt.Errorf("%w: sequenceDataType does not support pointers to pointers", types.ErrInvalidType)
	}
	return seqTypeOf, nonPointerType, nil
}

func convertSequencesWithKeyFromProto(pbSequences []*pb.SequenceWithKey, keyQueries []types.ContractKeyFilter) (iter.Seq2[string, types.Sequence], error) {
	type sequenceWithKey struct {
		Key      string
		Sequence types.Sequence
	}

	sequencesWithKey := make([]sequenceWithKey, len(pbSequences))

	keyToSeqTypeOf := make(map[string]reflect.Type)
	keyToNonPointerType := make(map[string]reflect.Type)

	for _, keyQuery := range keyQueries {
		seqTypeOf, nonPointerType, err := getSequenceTypeInformation(keyQuery.SequenceDataType)
		if err != nil {
			return nil, err
		}

		keyToSeqTypeOf[keyQuery.Key] = seqTypeOf
		keyToNonPointerType[keyQuery.Key] = nonPointerType
	}

	for idx, pbSequence := range pbSequences {
		seqTypeOf, nonPointerType := keyToSeqTypeOf[pbSequence.Key], keyToNonPointerType[pbSequence.Key]

		cpy := reflect.New(nonPointerType).Interface()
		if err := codecpb.DecodeVersionedBytes(cpy, pbSequence.Data); err != nil {
			return nil, err
		}

		// match provided data type either as pointer or non-pointer
		if seqTypeOf.Kind() != reflect.Pointer {
			cpy = reflect.Indirect(reflect.ValueOf(cpy)).Interface()
		}
		pbSeq := pbSequences[idx]
		sequencesWithKey[idx] = sequenceWithKey{
			Key: pbSeq.Key,
			Sequence: types.Sequence{
				Cursor: pbSeq.SequenceCursor,
				Head: types.Head{
					Height:    pbSeq.Head.Height,
					Hash:      pbSeq.Head.Hash,
					Timestamp: pbSeq.Head.Timestamp,
				},
				Data: cpy,
			},
		}
	}

	return func(yield func(string, types.Sequence) bool) {
		for _, s := range sequencesWithKey {
			if !yield(s.Key, s.Sequence) {
				return
			}
		}
	}, nil
}

func RegisterContractReaderService(s *grpc.Server, contractReader types.ContractReader) {
	service := goplugin.ServiceServer{Srv: contractReader}
	pb.RegisterServiceServer(s, &service)
	pb.RegisterContractReaderServer(s, NewServer(contractReader))
}
