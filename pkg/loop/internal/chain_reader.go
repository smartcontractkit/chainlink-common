package internal

import (
	"context"
	jsonv1 "encoding/json"
	"fmt"
	"unicode"

	jsonv2 "github.com/go-json-experiment/json"

	"github.com/fxamacker/cbor/v2"
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
	JSONEncodingVersion1 = iota
	JSONEncodingVersion2
	CBOREncodingVersion
)

// Version to be used for encoding ( version used for decoding is determined by data received )
// These are separate constants in case we want to upgrade their data formats independently
const ParamsCurrentEncodingVersion = CBOREncodingVersion
const RetvalCurrentEncodingVersion = CBOREncodingVersion

func encodeVersionedBytes(data any, version int32) (*pb.VersionedBytes, error) {
	var bytes []byte
	var err error

	switch version {
	case JSONEncodingVersion1:
		bytes, err = jsonv1.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.InvalidTypeError{}, err)
		}
	case JSONEncodingVersion2:
		bytes, err = jsonv2.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.InvalidTypeError{}, err)
		}
	case CBOREncodingVersion:
		enco := cbor.CoreDetEncOptions()
		enco.Time = cbor.TimeRFC3339Nano
		var enc cbor.EncMode
		enc, err = enco.EncMode()
		if err != nil {
			return nil, err
		}
		bytes, err = enc.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.InvalidTypeError{}, err)
		}
	default:
		return nil, fmt.Errorf("unsupported encoding version %d for data %v", version, data)
	}

	return &pb.VersionedBytes{Version: uint32(version), Data: bytes}, nil
}

func decodeVersionedBytes(res any, vData *pb.VersionedBytes) error {
	var err error
	switch vData.Version {
	case JSONEncodingVersion1:
		err = jsonv1.Unmarshal(vData.Data, res)
	case JSONEncodingVersion2:
		err = jsonv2.Unmarshal(vData.Data, res)
	case CBOREncodingVersion:
		err = cbor.Unmarshal(vData.Data, res)
	default:
		return fmt.Errorf("unsupported encoding version %d for versionedData %v", vData.Version, vData.Data)
	}

	if err != nil {
		return fmt.Errorf("%w: %w", types.InvalidTypeError{}, err)
	}
	return nil
}

func isArray(vData *pb.VersionedBytes) (bool, error) {
	data := vData.Data
	if len(data) > 0 {
		switch vData.Version {
		case JSONEncodingVersion1:
			fallthrough
		case JSONEncodingVersion2:
			i := int(0)
			for ; i < len(data); i++ {
				if !unicode.IsSpace(rune(data[i])) {
					break
				}
			}
			return i < len(data) && data[i] == '[', nil
		case CBOREncodingVersion:

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

	params := &map[string]any{}

	if err := decodeVersionedBytes(params, request.Params); err != nil {
		return nil, err
	}

	retVal := &map[string]any{}
	err := c.impl.GetLatestValue(ctx, bc, request.Method, params, retVal)
	if err != nil {
		return nil, err
	}

	encodedRetVal, err := encodeVersionedBytes(retVal, RetvalCurrentEncodingVersion)
	if err != nil {
		return nil, err
	}

	return &pb.GetLatestValueReply{RetVal: encodedRetVal}, nil
}

func unwrapClientError(err error) error {
	if s, ok := status.FromError(err); ok {
		switch s.Message() {
		case types.FieldNotFoundError{}.Error():
			return types.FieldNotFoundError{}
		}
	}
	return err
}
