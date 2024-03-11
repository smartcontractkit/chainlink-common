package internal

import (
	"context"
	jsonv1 "encoding/json"
	"fmt"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/pkg/errors"
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

func (c *chainReaderClient) QueryKeys(ctx context.Context, queryFilter types.QueryFilter, limitAndSort types.LimitAndSort) ([]types.Event, error) {
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! chainReaderClient Query Keys")
	if err := logQueryFilterValues(queryFilter); err != nil {
		return nil, errors.Wrap(err, "test debug")
	}

	if err := logLimitAndSortValue(limitAndSort); err != nil {
		return nil, errors.Wrap(err, "test debug")
	}
	return nil, nil
}

func logQueryFilterValues(queryFilter types.QueryFilter) error {
	switch filter := queryFilter.(type) {
	case *types.AndFilter:
		for _, f := range filter.Filters {
			fmt.Println("And Filter subfilter is: ^^^^^^^^^^^^^^^^^^^^^^^ ", f)
			err := logQueryFilterValues(f)
			if err != nil {
				return err
			}
			fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
		}
		return nil
	case *types.AddressFilter:
		fmt.Println("****************************")
		fmt.Println("Address Filter values are: ", filter.Address)
		fmt.Println("****************************")
	case *types.KeysFilter:
		fmt.Println("****************************")
		fmt.Println("Keys Filter values are: ", filter.Keys)
		fmt.Println("****************************")
		return nil
	default:
		fmt.Println("Unknown Filter value is: ", fmt.Sprintf("Unknown filter type %T ", queryFilter))

		return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Unknown filter typeoo %T ", queryFilter))
	}

	return nil
}

func logLimitAndSortValue(limitAndSort types.LimitAndSort) error {
	for _, sortBy := range limitAndSort.SortBy {
		switch sort := sortBy.(type) {
		case *types.SortByTimestamp:
			fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			fmt.Println("SortByTimestamp ", sort.GetDirection())
			fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			return nil
		case *types.SortByBlock:
			fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			fmt.Println("SortByBlock ", sort.GetDirection())
			fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			return nil
		case *types.SortBySequence:
			fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			fmt.Println("SortBySequence ", sort.GetDirection())
			fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
			return nil
		default:
			fmt.Println("Unknown sort value is: ", fmt.Sprintf("Unknown filter type %T ", sortBy))
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Unknown filter typeoo %T ", sortBy))
		}
	}
	return nil
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

func (c *chainReaderServer) QueryKeys(ctx context.Context, request *pb.QueryKeysRequest) (*pb.QueryKeysReply, error) {
	queryFilter, err := parseQueryKeysFilter(request.GetFilter())
	if err != nil {
		return nil, err
	}

	limitAndSort, err := parseLimitAndSort(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	_, err = c.impl.QueryKeys(ctx, queryFilter, limitAndSort)
	if err != nil {
		return nil, err
	}
	return &pb.QueryKeysReply{RetVal: nil}, nil
}

func parseQueryKeysFilter(request *pb.QueryKeysFilter) (types.QueryFilter, error) {
	switch filter := request.Filter.(type) {
	case *pb.QueryKeysFilter_AndFilter:
		var parsedQueryFilters []types.QueryFilter
		for _, subQueryFilterRequest := range filter.AndFilter.Filters {
			parsedQueryFilter, err := parseQueryKeysFilter(subQueryFilterRequest)
			if err != nil {
				return nil, err
			}
			parsedQueryFilters = append(parsedQueryFilters, parsedQueryFilter)
		}
		return &types.AndFilter{Filters: parsedQueryFilters}, nil
	case *pb.QueryKeysFilter_AddressFilter:
		return &types.AddressFilter{Address: filter.AddressFilter.Addresses}, nil
	case *pb.QueryKeysFilter_KeysFilter:
		return &types.KeysFilter{Keys: filter.KeysFilter.Keys}, nil
	case *pb.QueryKeysFilter_KeysByValueFilter:
		var keysByValueFilter types.KeysByValueFilter
		for _, k := range filter.KeysByValueFilter.Keys {
			keysByValueFilter.Keys = append(keysByValueFilter.Keys, k.Keys...)
		}
		for _, v := range filter.KeysByValueFilter.Values {
			keysByValueFilter.Values = append(keysByValueFilter.Values, v.Values)
		}
		return &keysByValueFilter, nil
	case *pb.QueryKeysFilter_ConfirmationsFilter:
		return &types.ConfirmationFilter{Confirmations: types.Confirmations(filter.ConfirmationsFilter.Confirmations.Confirmation)}, nil
	case *pb.QueryKeysFilter_BlockFilter:
		return &types.BlockFilter{Block: filter.BlockFilter.BlockNumber, Operator: types.ComparisonOperator(filter.BlockFilter.Operator.ComparisonOperator)}, nil
	case *pb.QueryKeysFilter_TxHashFilter:
		return &types.TxHashFilter{TxHash: filter.TxHashFilter.TxHash}, nil
	case *pb.QueryKeysFilter_TimestampFilter:
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
