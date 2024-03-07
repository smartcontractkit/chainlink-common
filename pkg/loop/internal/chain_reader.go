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

func (c *chainReaderClient) QueryKeys(ctx context.Context, queryFilter types.QueryFilter) ([]types.Event, error) {
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! chainReaderClient Query Keys")
	err := logQueryFilterValues(queryFilter)
	if err != nil {
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
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!  Query Keys server")

	queryFilter, err := parseQueryFilterRequest(request)
	if err != nil {
		return nil, err
	}

	_, err = c.impl.QueryKeys(ctx, queryFilter)
	if err != nil {
		return nil, err
	}
	return &pb.QueryKeysReply{RetVal: nil}, nil
}

func parseQueryFilterRequest(request *pb.QueryKeysRequest) (types.QueryFilter, error) {
	switch filter := request.Filter.(type) {
	case *pb.QueryKeysRequest_AndFilter:
		var parsedQueryFilters []types.QueryFilter
		for _, subQueryFilterRequest := range filter.AndFilter.Filters {
			parsedQueryFilter, err := parseQueryFilterRequest(subQueryFilterRequest)
			if err != nil {
				return nil, err
			}
			parsedQueryFilters = append(parsedQueryFilters, parsedQueryFilter)
		}
		return &types.AndFilter{Filters: parsedQueryFilters}, nil
	case *pb.QueryKeysRequest_AddressFilter:
		return &types.AddressFilter{Address: filter.AddressFilter.Addresses}, nil
	case *pb.QueryKeysRequest_KeysFilter:
		return &types.KeysFilter{Keys: filter.KeysFilter.Keys}, nil
	default:
		return nil, status.Errorf(codes.InvalidArgument, "Unknown filter type")
	}
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
