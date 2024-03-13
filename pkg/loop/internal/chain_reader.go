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
	convertedQueryFilter, err := convertQueryFilter(queryFilter)
	if err != nil {
		return nil, err
	}

	convertedLimitAndSort, err := convertLimitAndSort(limitAndSort)
	if err != nil {
		return nil, err
	}

	_, err = c.grpc.QueryKey(ctx, &pb.QueryKeyRequest{Key: key, Filter: convertedQueryFilter, LimitAndSort: convertedLimitAndSort})
	if err != nil {
		return nil, wrapRPCErr(err)
	}

	return nil, nil
}

func (c *chainReaderClient) QueryKeys(ctx context.Context, keys []string, queryFilter types.QueryFilter, limitAndSort types.LimitAndSort) ([][]types.Sequence, error) {
	convertedQueryFilter, err := convertQueryFilter(queryFilter)
	if err != nil {
		return nil, err
	}

	convertedLimitAndSort, err := convertLimitAndSort(limitAndSort)
	if err != nil {
		return nil, err
	}

	_, err = c.grpc.QueryKeys(ctx, &pb.QueryKeysRequest{Keys: keys, Filter: convertedQueryFilter, LimitAndSort: convertedLimitAndSort})
	if err != nil {
		return nil, wrapRPCErr(err)
	}
	return nil, nil
}

func (c *chainReaderClient) QueryKeyByValues(ctx context.Context, key string, values []string, queryFilter types.QueryFilter, limitAndSort types.LimitAndSort) ([]types.Sequence, error) {
	convertedQueryFilter, err := convertQueryFilter(queryFilter)
	if err != nil {
		return nil, err
	}

	convertedLimitAndSort, err := convertLimitAndSort(limitAndSort)
	if err != nil {
		return nil, err
	}

	_, err = c.grpc.QueryKeyByValues(ctx, &pb.QueryKeyByValuesRequest{Key: key, Filter: convertedQueryFilter, LimitAndSort: convertedLimitAndSort})
	if err != nil {
		return nil, wrapRPCErr(err)
	}

	return nil, nil
}

func (c *chainReaderClient) QueryKeysByValues(ctx context.Context, keys []string, values [][]string, queryFilter types.QueryFilter, limitAndSort types.LimitAndSort) ([][]types.Sequence, error) {
	convertedQueryFilter, err := convertQueryFilter(queryFilter)
	if err != nil {
		return nil, err
	}

	convertedLimitAndSort, err := convertLimitAndSort(limitAndSort)
	if err != nil {
		return nil, err
	}

	_, err = c.grpc.QueryKeysByValues(ctx, &pb.QueryKeysByValuesRequest{Keys: keys, Filter: convertedQueryFilter, LimitAndSort: convertedLimitAndSort})
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
	queryFilter, err := parseQueryFilter(request.GetFilter())
	if err != nil {
		return nil, err
	}

	limitAndSort, err := parseLimitAndSort(request.GetLimitAndSort())
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
	queryFilter, err := parseQueryFilter(request.GetFilter())
	if err != nil {
		return nil, err
	}

	limitAndSort, err := parseLimitAndSort(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	_, err = c.impl.QueryKeys(ctx, request.Keys, queryFilter, limitAndSort)
	if err != nil {
		return nil, err
	}
	return &pb.QueryKeysReply{RetVal: nil}, nil
}

func (c *chainReaderServer) QueryKeyByValues(ctx context.Context, request *pb.QueryKeyByValuesRequest) (*pb.QueryKeysReply, error) {
	queryFilter, err := parseQueryFilter(request.GetFilter())
	if err != nil {
		return nil, err
	}

	limitAndSort, err := parseLimitAndSort(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	_, err = c.impl.QueryKeyByValues(ctx, request.Key, request.Values.Value, queryFilter, limitAndSort)
	if err != nil {
		return nil, err
	}
	return &pb.QueryKeysReply{RetVal: nil}, nil
}

func (c *chainReaderServer) QueryKeysByValues(ctx context.Context, request *pb.QueryKeysByValuesRequest) (*pb.QueryKeysReply, error) {
	queryFilter, err := parseQueryFilter(request.GetFilter())
	if err != nil {
		return nil, err
	}

	limitAndSort, err := parseLimitAndSort(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	var values [][]string
	for _, keyValues := range request.Values {
		values = append(values, keyValues.Value)
	}
	_, err = c.impl.QueryKeysByValues(ctx, request.Keys, values, queryFilter, limitAndSort)
	if err != nil {
		return nil, err
	}
	return &pb.QueryKeysReply{RetVal: nil}, nil
}

func convertQueryFilter(filter types.QueryFilter) (*pb.QueryFilter, error) {
	switch filter := filter.(type) {
	case *types.AndFilter:
		var parsedQueryFilters []*pb.QueryFilter
		for _, subQueryFilterRequest := range filter.Filters {
			parsedQueryFilter, err := convertQueryFilter(subQueryFilterRequest)
			if err != nil {
				return nil, err
			}
			parsedQueryFilters = append(parsedQueryFilters, parsedQueryFilter)
		}
		return &pb.QueryFilter{Filter: &pb.QueryFilter_AndFilter{
			AndFilter: &pb.AndFilter{Filters: parsedQueryFilters},
		}}, nil
	case *types.AddressFilter:
		return &pb.QueryFilter{Filter: &pb.QueryFilter_AddressFilter{
			AddressFilter: &pb.AddressFilter{Addresses: filter.Addresses},
		}}, nil
	case *types.ConfirmationsFilter:
		return &pb.QueryFilter{Filter: &pb.QueryFilter_ConfirmationsFilter{
			ConfirmationsFilter: &pb.ConfirmationsFilter{
				Confirmations: &pb.Confirmations{Confirmation: int32(filter.Confirmations)},
			}}}, nil
	case *types.BlockFilter:
		return &pb.QueryFilter{Filter: &pb.QueryFilter_BlockFilter{
			BlockFilter: &pb.BlockFilter{
				BlockNumber: filter.Block,
				Operator:    &pb.ComparisonOperator{ComparisonOperator: int32(filter.Operator)},
			},
		}}, nil
	case *types.TxHashFilter:
		return &pb.QueryFilter{Filter: &pb.QueryFilter_TxHashFilter{
			TxHashFilter: &pb.TxHashFilter{
				TxHash: filter.TxHash},
		}}, nil
	case *types.TimestampFilter:
		return &pb.QueryFilter{Filter: &pb.QueryFilter_TimestampFilter{
			TimestampFilter: &pb.TimestampFilter{
				Timestamp: filter.Timestamp,
				Operator:  &pb.ComparisonOperator{ComparisonOperator: int32(filter.Operator)},
			},
		}}, nil
	default:
		return nil, status.Errorf(codes.InvalidArgument, "Unknown filter type")
	}
}

func convertLimitAndSort(limitAndSort types.LimitAndSort) (*pb.LimitAndSort, error) {
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

func parseQueryFilter(request *pb.QueryFilter) (types.QueryFilter, error) {
	switch filter := request.Filter.(type) {
	case *pb.QueryFilter_AndFilter:
		var parsedQueryFilters []types.QueryFilter
		for _, subQueryFilterRequest := range filter.AndFilter.Filters {
			parsedQueryFilter, err := parseQueryFilter(subQueryFilterRequest)
			if err != nil {
				return nil, err
			}
			parsedQueryFilters = append(parsedQueryFilters, parsedQueryFilter)
		}
		return &types.AndFilter{Filters: parsedQueryFilters}, nil
	case *pb.QueryFilter_AddressFilter:
		return &types.AddressFilter{Addresses: filter.AddressFilter.Addresses}, nil
	case *pb.QueryFilter_ConfirmationsFilter:
		return &types.ConfirmationsFilter{Confirmations: types.Confirmations(filter.ConfirmationsFilter.Confirmations.Confirmation)}, nil
	case *pb.QueryFilter_BlockFilter:
		return &types.BlockFilter{Block: filter.BlockFilter.BlockNumber, Operator: types.ComparisonOperator(filter.BlockFilter.Operator.ComparisonOperator)}, nil
	case *pb.QueryFilter_TxHashFilter:
		return &types.TxHashFilter{TxHash: filter.TxHashFilter.TxHash}, nil
	case *pb.QueryFilter_TimestampFilter:
		return &types.TimestampFilter{
			Timestamp: filter.TimestampFilter.Timestamp,
			Operator:  types.ComparisonOperator(filter.TimestampFilter.Operator.ComparisonOperator),
		}, nil
	default:
		return nil, status.Errorf(codes.InvalidArgument, "Unknown filter type")
	}
}

func parseLimitAndSort(limitAndSort *pb.LimitAndSort) (types.LimitAndSort, error) {
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
