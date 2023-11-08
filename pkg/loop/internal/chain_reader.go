package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

var _ types.ChainReader = (*chainReaderClient)(nil)

type chainReaderClient struct {
	*brokerExt
	grpc pb.ChainReaderClient
}

// enum of all known encoding formats for versioned data
const (
	SimpleJsonEncodingVersion = iota
	CborEncodingVersion
)

// Version to be used for encoding ( version used for decoding is determined by data received )
// These are separate constants in case we want to upgrade their data formats independently
const ParamsCurrentEncodingVersion = CborEncodingVersion
const RetvalCurrentEncodingVersion = CborEncodingVersion

func encodeVersionedBytes(data any, version int32) (*pb.VersionedBytes, error) {
	var bytes []byte
	var err error

	switch version {
	case SimpleJsonEncodingVersion:
		bytes, err = json.Marshal(data)
	case CborEncodingVersion:
		enco := cbor.CoreDetEncOptions()
		enco.Time = cbor.TimeRFC3339Nano
		var enc cbor.EncMode
		enc, err = enco.EncMode()
		if err != nil {
			return nil, err
		}
		bytes, err = enc.Marshal(data)
	default:
		return nil, fmt.Errorf("unsupported encoding version %d for data %v", version, data)
	}

	if err != nil {
		return nil, types.InvalidTypeError{}
	}
	return &pb.VersionedBytes{Version: uint32(version), Data: bytes}, nil
}

func decodeVersionedBytes(res any, vData *pb.VersionedBytes) error {
	var err error
	switch vData.Version {
	case SimpleJsonEncodingVersion:
		err = json.Unmarshal(vData.Data, res)
	case CborEncodingVersion:
		err = cbor.Unmarshal(vData.Data, res)
	default:
		return fmt.Errorf("unsupported encoding version %d for versionedData %v", vData.Version, vData.Data)
	}

	if err != nil {
		return types.InvalidTypeError{}
	}
	return nil
}

func isArray(vData *pb.VersionedBytes) (bool, error) {
	data := vData.Data
	if len(data) > 0 {
		switch vData.Version {
		case SimpleJsonEncodingVersion:
			return data[0] == '[', nil
		case CborEncodingVersion:

			// Major type for array in CBOR is 4 which is 100 in binary.
			// So, we are checking if the first 3 bits are 100.
			return data[0]>>5 == 4, nil
		default:
			return false, fmt.Errorf("Unsupported encoding version %d for versionedData %v", vData.Version, vData.Data)
		}
	}

	return false, nil
}

func (c *chainReaderClient) GetLatestValue(ctx context.Context, bc types.BoundContract, method string, params, retVal any) error {
	versionedParams, err := encodeVersionedBytes(params, ParamsCurrentEncodingVersion)
	if err != nil {
		return err
	}

	boundContract := pb.BoundContract{Name: bc.Name, Address: bc.Address, Pending: bc.Pending}

	reply, err := c.grpc.GetLatestValue(ctx, &pb.GetLatestValueRequest{Bc: &boundContract, Method: method, Params: versionedParams})
	if err != nil {
		return unwrapClientError(err)
	}

	return decodeVersionedBytes(retVal, reply.RetVal)
}

func (c *chainReaderClient) Encode(ctx context.Context, item any, itemType string) (libocr.Report, error) {
	versionedParams, err := encodeVersionedBytes(item, ParamsCurrentEncodingVersion)
	if err != nil {
		return nil, err
	}

	reply, err := c.grpc.GetEncoding(ctx, &pb.GetEncodingRequest{
		Params:   versionedParams,
		ItemType: itemType,
	})

	if err != nil {
		return nil, unwrapClientError(err)
	}

	return reply.RetVal, nil
}

func (c *chainReaderClient) Decode(ctx context.Context, raw []byte, into any, itemType string) error {
	k := reflect.ValueOf(into).Kind()
	request := &pb.GetDecodingRequest{
		Encoded:    raw,
		ItemType:   itemType,
		ForceSplit: k == reflect.Array || k == reflect.Slice,
	}
	resp, err := c.grpc.GetDecoding(ctx, request)
	if err != nil {
		return unwrapClientError(err)
	}

	return decodeVersionedBytes(into, resp.RetVal)
}

func (c *chainReaderClient) GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error) {
	res, err := c.grpc.GetMaxSize(ctx, &pb.GetMaxSizeRequest{N: int32(n), ItemType: itemType, ForEncoding: true})
	if err != nil {
		return 0, unwrapClientError(err)
	}

	return int(res.SizeInBytes), nil
}

func (c *chainReaderClient) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
	res, err := c.grpc.GetMaxSize(ctx, &pb.GetMaxSizeRequest{N: int32(n), ItemType: itemType, ForEncoding: false})
	if err != nil {
		return 0, unwrapClientError(err)
	}

	return int(res.SizeInBytes), nil
}

var _ pb.ChainReaderServer = (*chainReaderServer)(nil)

type chainReaderServer struct {
	pb.UnimplementedChainReaderServer
	impl types.ChainReader
}

func (c *chainReaderServer) GetLatestValue(ctx context.Context, request *pb.GetLatestValueRequest) (*pb.GetLatestValueReply, error) {
	var bc types.BoundContract
	bc.Name = request.Bc.Name[:]
	bc.Address = request.Bc.Address[:]
	bc.Pending = request.Bc.Pending

	params, err := c.getEncodedType(request.Method, false, true)
	if err != nil {
		return nil, err
	}

	if err = decodeVersionedBytes(params, request.Params); err != nil {
		return nil, err
	}

	retVal, err := c.getEncodedType(request.Method, false, false)
	if err != nil {
		return nil, err
	}
	err = c.impl.GetLatestValue(ctx, bc, request.Method, params, retVal)
	if err != nil {
		return nil, err
	}

	encodedRetVal, err := encodeVersionedBytes(retVal, RetvalCurrentEncodingVersion)
	if err != nil {
		return nil, err
	}

	return &pb.GetLatestValueReply{RetVal: encodedRetVal}, nil
}

func (c *chainReaderServer) GetEncoding(ctx context.Context, req *pb.GetEncodingRequest) (*pb.GetEncodingResponse, error) {
	forceArray, err := isArray(req.Params)
	if err != nil {
		return nil, err
	}

	encodedType, err := c.getEncodedType(req.ItemType, forceArray, true)
	if err != nil {
		return nil, err
	}

	if err = decodeVersionedBytes(encodedType, req.Params); err != nil {
		return nil, err
	}

	encoded, err := c.impl.Encode(ctx, encodedType, req.ItemType)
	return &pb.GetEncodingResponse{RetVal: encoded}, err
}

func (c *chainReaderServer) GetDecoding(ctx context.Context, req *pb.GetDecodingRequest) (*pb.GetDecodingResponse, error) {
	encodedType, err := c.getEncodedType(req.ItemType, req.ForceSplit, false)
	if err != nil {
		return nil, err
	}

	err = c.impl.Decode(ctx, req.Encoded, encodedType, req.ItemType)
	if err != nil {
		return nil, err
	}

	versionBytes, err := encodeVersionedBytes(encodedType, RetvalCurrentEncodingVersion)
	return &pb.GetDecodingResponse{RetVal: versionBytes}, err
}

func (c *chainReaderServer) GetMaxSize(ctx context.Context, req *pb.GetMaxSizeRequest) (*pb.GetMaxSizeResponse, error) {
	var sizeFn func(context.Context, int, string) (int, error)
	if req.ForEncoding {
		sizeFn = c.impl.GetMaxEncodingSize
	} else {
		sizeFn = c.impl.GetMaxDecodingSize
	}

	maxSize, err := sizeFn(ctx, int(req.N), req.ItemType)
	if err != nil {
		return nil, err
	}
	return &pb.GetMaxSizeResponse{SizeInBytes: int32(maxSize)}, nil
}

//func (c *chainReaderServer) RegisterEventFilter(ctx context.Context, in *pb.RegisterEventFilterRequest) (*pb.RegisterEventFilterReply, error) {
//	return nil, nil
//}
//func (c *chainReaderServer) UnregisterEventFilter(ctx context.Context, in *pb.UnregisterEventFilterRequest) (*pb.RegisterEventFilterReply, error) {
//	return nil, nil
//}
//func (c *chainReaderServer) QueryEvents(ctx context.Context, in *pb.QueryEventsRequest) (*pb.QueryEventsReply, error) {
//	return nil, nil
//}

func (c *chainReaderServer) getEncodedType(itemType string, forceArray bool, forEncoding bool) (any, error) {
	if rc, ok := c.impl.(types.RemoteCodec); ok {
		return rc.CreateType(itemType, forceArray, forEncoding)
	}

	return &map[string]any{}, nil
}

func unwrapClientError(err error) error {
	if s, ok := status.FromError(err); ok {
		switch s.Message() {
		case types.InvalidEncodingError{}.Error():
			return types.InvalidEncodingError{}
		case types.InvalidTypeError{}.Error():
			return types.InvalidTypeError{}
		case types.FieldNotFoundError{}.Error():
			return types.FieldNotFoundError{}
		case types.WrongNumberOfElements{}.Error():
			return types.WrongNumberOfElements{}
		case types.NotASliceError{}.Error():
			return types.NotASliceError{}
		}
	}
	return err
}
