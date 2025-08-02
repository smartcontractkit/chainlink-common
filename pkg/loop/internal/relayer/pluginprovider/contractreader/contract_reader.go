package contractreader

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	loopjson "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/json"
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
}

func NewClient(serviceClient serviceClient, grpc pb.ContractReaderClient, opts ...ClientOpt) *Client {
	client := &Client{
		serviceClient: serviceClient,
		grpc:          grpc,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func handleValuesPointer(retVal any, jsonRetVal []byte, retValTypeHint string) (handled bool, err error) {
	valPtr, ok := retVal.(*values.Value)
	if !ok || retValTypeHint == "" {
		return false, nil
	}

	// Unmarshal with the type hint from the server
	result, err := loopjson.UnmarshalWithHint(jsonRetVal, retValTypeHint)
	if err != nil {
		return true, fmt.Errorf("failed to unmarshal with type hint: %w", err)
	}

	if retValTypeHint == "values.Value" {
		// Result is already a values.Value, assign directly
		if val, ok := result.(values.Value); ok {
			*valPtr = val
		} else {
			return true, fmt.Errorf("type hint indicates values.Value but result is %T", result)
		}
	} else {
		// Wrap the result into a values.Value
		wrapped, err := values.Wrap(result)
		if err != nil {
			return true, fmt.Errorf("failed to wrap value: %w", err)
		}
		*valPtr = wrapped
	}
	return true, nil
}

func (c *Client) GetLatestValue(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, retVal any) error {
	jsonParams, paramsTypeHint, err := loopjson.MarshalWithHint(params)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal params: %s", types.ErrInvalidType, err.Error())
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
			Params:         jsonParams,
			ParamsTypeHint: paramsTypeHint,
		},
	)
	if err != nil {
		return net.WrapRPCErr(err)
	}

	if handled, err := handleValuesPointer(retVal, reply.RetVal, reply.RetValTypeHint); err != nil {
		return err
	} else if handled {
		return nil
	}

	return loopjson.UnmarshalJson(reply.RetVal, retVal)
}

func (c *Client) GetLatestValueWithHeadData(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, retVal any) (*types.Head, error) {
	jsonParams, paramsTypeHint, err := loopjson.MarshalWithHint(params)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal params: %s", types.ErrInvalidType, err.Error())
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
			Params:         jsonParams,
			ParamsTypeHint: paramsTypeHint,
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

	if handled, err := handleValuesPointer(retVal, reply.RetVal, reply.RetValTypeHint); err != nil {
		return nil, err
	} else if handled {
		return headData, nil
	}

	return headData, loopjson.UnmarshalJson(reply.RetVal, retVal)
}

func (c *Client) BatchGetLatestValues(ctx context.Context, request types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	pbRequest, err := convertBatchGetLatestValuesRequestToProto(request)
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
		contract := convertBoundContractToProto(keyQuery.Contract)

		pbQueryFilter, err := convertQueryFilterToProto(keyQuery.KeyFilter)
		if err != nil {
			return nil, err
		}

		filters = append(filters, &pb.ContractKeyFilter{
			Contract: contract,
			Filter:   pbQueryFilter,
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
	impl types.ContractReader
}

func NewServer(impl types.ContractReader, opts ...ServerOpt) pb.ContractReaderServer {
	server := &Server{
		impl: impl,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func (c *Server) GetLatestValue(ctx context.Context, request *pb.GetLatestValueRequest) (*pb.GetLatestValueReply, error) {
	confidenceLevel, err := chaincommonpb.ConfidenceFromProto(request.Confidence)
	if err != nil {
		return nil, err
	}

	params, err := loopjson.UnmarshalWithHint(request.Params, request.ParamsTypeHint)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal params with type hint: %w", err)
	}

	var result any
	err = c.impl.GetLatestValue(ctx, request.ReadIdentifier, confidenceLevel, params, &result)
	if err != nil {
		return nil, err
	}

	// Marshal result with type hint
	jsonResult, typeHint, err := loopjson.MarshalWithHint(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &pb.GetLatestValueReply{
		RetVal:         jsonResult,
		RetValTypeHint: typeHint,
	}, nil
}

func (c *Server) GetLatestValueWithHeadData(ctx context.Context, request *pb.GetLatestValueRequest) (*pb.GetLatestValueWithHeadDataReply, error) {
	confidenceLevel, err := chaincommonpb.ConfidenceFromProto(request.Confidence)
	if err != nil {
		return nil, err
	}

	params, err := loopjson.UnmarshalWithHint(request.Params, request.ParamsTypeHint)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal params with type hint: %w", err)
	}

	var result any
	headData, err := c.impl.GetLatestValueWithHeadData(ctx, request.ReadIdentifier, confidenceLevel, params, &result)
	if err != nil {
		return nil, err
	}

	// Marshal result with type hint
	jsonResult, typeHint, err := loopjson.MarshalWithHint(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var pbHeadData *pb.Head
	if headData != nil {
		pbHeadData = &pb.Head{
			Height:    headData.Height,
			Hash:      headData.Hash,
			Timestamp: headData.Timestamp,
		}
	}

	return &pb.GetLatestValueWithHeadDataReply{
		RetVal:         jsonResult,
		RetValTypeHint: typeHint,
		HeadData:       pbHeadData,
	}, nil
}

func (c *Server) BatchGetLatestValues(ctx context.Context, pbRequest *pb.BatchGetLatestValuesRequest) (*pb.BatchGetLatestValuesReply, error) {
	request, err := convertBatchGetLatestValuesRequestFromProto(pbRequest)
	if err != nil {
		return nil, err
	}

	reply, err := c.impl.BatchGetLatestValues(ctx, request)
	if err != nil {
		return nil, err
	}

	return newPbBatchGetLatestValuesReply(reply)
}

func (c *Server) QueryKey(ctx context.Context, request *pb.QueryKeyRequest) (*pb.QueryKeyReply, error) {
	contract := convertBoundContractFromProto(request.Contract)

	queryFilter, err := convertQueryFiltersFromProto(request.Filter)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := chaincommonpb.ConvertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	var sequenceData any
	sequences, err := c.impl.QueryKey(ctx, contract, queryFilter, limitAndSort, &sequenceData)
	if err != nil {
		return nil, err
	}

	pbSequences, err := convertSequencesToProto(sequences)
	if err != nil {
		return nil, err
	}

	return &pb.QueryKeyReply{Sequences: pbSequences}, nil
}

func (c *Server) QueryKeys(ctx context.Context, request *pb.QueryKeysRequest) (*pb.QueryKeysReply, error) {
	var filters []types.ContractKeyFilter
	for _, keyQuery := range request.Filters {
		contract := convertBoundContractFromProto(keyQuery.Contract)

		queryFilter, err := convertQueryFiltersFromProto(keyQuery.Filter)
		if err != nil {
			return nil, err
		}

		var sequenceData any
		filters = append(filters, types.ContractKeyFilter{
			Contract:         contract,
			KeyFilter:        queryFilter,
			SequenceDataType: &sequenceData,
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

	pbSequences, err := convertSequencesWithKeyToProto(sequences)
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

func newPbBatchGetLatestValuesReply(result types.BatchGetLatestValuesResult) (*pb.BatchGetLatestValuesReply, error) {
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

			// Marshal return value with type hint
			jsonRetVal, typeHint, err := loopjson.MarshalWithHint(returnVal)
			if err != nil {
				return nil, err
			}

			batchResult.Results = append(batchResult.Results, &pb.BatchReadResult{
				ReadName:          batchCall.ReadName,
				ReturnVal:         jsonRetVal,
				ReturnValTypeHint: typeHint,
				Error:             replyErr,
			})
		}
	}

	return &pb.BatchGetLatestValuesReply{Results: results}, nil
}

func convertBatchGetLatestValuesRequestToProto(request types.BatchGetLatestValuesRequest) (*pb.BatchGetLatestValuesRequest, error) {
	requests := make([]*pb.ContractBatch, len(request))

	var requestIdx int

	for binding, nextBatch := range request {
		requests[requestIdx] = &pb.ContractBatch{
			Contract: convertBoundContractToProto(binding),
			Reads:    make([]*pb.BatchRead, len(nextBatch)),
		}

		for readIdx, batchCall := range nextBatch {
			jsonParams, paramsTypeHint, err := loopjson.MarshalWithHint(batchCall.Params)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to marshal params for read %s: %s", types.ErrInvalidType, batchCall.ReadName, err.Error())
			}

			requests[requestIdx].Reads[readIdx] = &pb.BatchRead{
				ReadName:       batchCall.ReadName,
				Params:         jsonParams,
				ParamsTypeHint: paramsTypeHint,
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

	for _, expr := range filter.Expressions {
		pbExpr, err := chaincommonpb.ConvertExpressionToProto(expr)
		if err != nil {
			return nil, err
		}
		pbQueryFilter.Expression = append(pbQueryFilter.Expression, pbExpr)
	}

	return pbQueryFilter, nil
}

func convertSequencesToProto(sequences []types.Sequence) ([]*pb.Sequence, error) {
	var pbSequences []*pb.Sequence
	for _, sequence := range sequences {
		jsonData, typeHint, err := loopjson.MarshalWithHint(sequence.Data)
		if err != nil {
			return nil, err
		}
		pbSequence := &pb.Sequence{
			SequenceCursor: sequence.Cursor,
			TxHash:         sequence.TxHash,
			DataTypeHint:   typeHint,
			Head: &pb.Head{
				Height:    sequence.Height,
				Hash:      sequence.Hash,
				Timestamp: sequence.Timestamp,
			},
			Data: jsonData,
		}
		pbSequences = append(pbSequences, pbSequence)
	}
	return pbSequences, nil
}

func convertSequencesWithKeyToProto(sequences iter.Seq2[string, types.Sequence]) ([]*pb.SequenceWithKey, error) {
	var pbSequences []*pb.SequenceWithKey
	for key, sequence := range sequences {
		jsonData, typeHint, err := loopjson.MarshalWithHint(sequence.Data)
		if err != nil {
			return nil, err
		}
		pbSequence := &pb.SequenceWithKey{
			Key:            key,
			SequenceCursor: sequence.Cursor,
			TxHash:         sequence.TxHash,
			DataTypeHint:   typeHint,
			Head: &pb.Head{
				Height:    sequence.Height,
				Hash:      sequence.Hash,
				Timestamp: sequence.Timestamp,
			},
			Data: jsonData,
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

			if handled, err := handleValuesPointer(req.ReturnVal, res.ReturnVal, res.ReturnValTypeHint); err != nil {
				return nil, err
			} else if !handled {
				if err := loopjson.UnmarshalJson(res.ReturnVal, req.ReturnVal); err != nil {
					return nil, err
				}
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

func convertBatchGetLatestValuesRequestFromProto(pbRequest *pb.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesRequest, error) {
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
			params, err := loopjson.UnmarshalWithHint(pbReadCall.Params, pbReadCall.ParamsTypeHint)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal params with type hint for batch read %s: %w", pbReadCall.ReadName, err)
			}

			var returnVal any
			call := types.BatchRead{
				ReadName:  pbReadCall.ReadName,
				Params:    params,
				ReturnVal: &returnVal,
			}
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
	for _, pbExpr := range pbQueryFilters.Expression {
		expression, err := chaincommonpb.ConvertExpressionFromProto(pbExpr)
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

		if err := unmarshalSequenceData(pbSequence.Data, pbSequence.DataTypeHint, sequenceDataType, cpy); err != nil {
			return nil, err
		}

		// match provided data type either as pointer or non-pointer
		if seqTypeOf.Kind() != reflect.Pointer {
			cpy = reflect.Indirect(reflect.ValueOf(cpy)).Interface()
		}

		sequences[idx] = types.Sequence{
			Cursor: pbSequences[idx].SequenceCursor,
			TxHash: pbSequences[idx].TxHash,
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

func isValueSequenceType(sequenceDataType any) bool {
	if _, ok := sequenceDataType.(values.Value); ok {
		return true
	}

	typ := reflect.TypeOf(sequenceDataType)
	if typ == nil || typ.Kind() != reflect.Ptr {
		return false
	}

	if typ.Elem() == reflect.TypeOf((*values.Value)(nil)).Elem() {
		return true
	}

	return false
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

func unmarshalSequenceData(data []byte, dataTypeHint string, sequenceDataType any, cpy any) error {
	if isValueSequenceType(sequenceDataType) && dataTypeHint != "" {
		// Unmarshal with the type hint from the server
		unmarshaledResult, err := loopjson.UnmarshalWithHint(data, dataTypeHint)
		if err != nil {
			return fmt.Errorf("failed to unmarshal with type hint: %w", err)
		}

		// Check if result is already a values.Value
		if dataTypeHint == "values.Value" {
			// Result is already a values.Value, use directly
			if val, ok := unmarshaledResult.(values.Value); ok {
				reflect.ValueOf(cpy).Elem().Set(reflect.ValueOf(val))
			} else {
				return fmt.Errorf("type hint indicates values.Value but result is %T", unmarshaledResult)
			}
		} else {
			// Wrap the result into a values.Value
			wrapped, err := values.Wrap(unmarshaledResult)
			if err != nil {
				return fmt.Errorf("failed to wrap value: %w", err)
			}

			// For values.Value, we need to set the wrapped value
			reflect.ValueOf(cpy).Elem().Set(reflect.ValueOf(wrapped))
		}
		return nil
	}

	return loopjson.UnmarshalJson(data, cpy)
}

func findKeyQueryByKey(keyQueries []types.ContractKeyFilter, key string) *types.ContractKeyFilter {
	for i := range keyQueries {
		if keyQueries[i].Key == key {
			return &keyQueries[i]
		}
	}
	return nil
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

		keyQuery := findKeyQueryByKey(keyQueries, pbSequence.Key)
		if keyQuery != nil {
			if err := unmarshalSequenceData(pbSequence.Data, pbSequence.DataTypeHint, keyQuery.SequenceDataType, cpy); err != nil {
				return nil, err
			}
		} else if err := loopjson.UnmarshalJson(pbSequence.Data, cpy); err != nil {
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
				TxHash: pbSeq.TxHash,
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
