package internal

import (
	"context"
	jsonv1 "encoding/json"
	"errors"
	"fmt"
	"strings"

	jsonv2 "github.com/go-json-experiment/json"

	"github.com/fxamacker/cbor/v2"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
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

// Version to be used for encoding (version used for decoding is determined by data received)
const CurrentEncodingVersion = CBOREncodingVersion

func encodeVersionedBytes(data any, version int32) (*pb.VersionedBytes, error) {
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
			return nil, err
		}
		bytes, err = enc.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
		}
	default:
		return nil, fmt.Errorf("%w: unsupported encoding version %d for data %v", types.ErrInvalidEncoding, version, data)
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
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}
	return nil
}

func (c *chainReaderClient) GetLatestValue(ctx context.Context, bc types.BoundContract, method string, params, retVal any) error {
	versionedParams, err := encodeVersionedBytes(params, CurrentEncodingVersion)
	if err != nil {
		return err
	}

	boundContract := pb.BoundContract{Name: bc.Name, Address: bc.Address, Pending: bc.Pending}

	reply, err := c.grpc.GetLatestValue(ctx, &pb.GetLatestValueRequest{Bc: &boundContract, Method: method, Params: versionedParams})
	if err != nil {
		return types.UnwrapClientError(err)
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

	params, err := getEncodedType(request.Method, c.impl, true)
	if err != nil {
		return nil, err
	}

	if err = decodeVersionedBytes(params, request.Params); err != nil {
		return nil, err
	}

	retVal, err := getEncodedType(request.Method, c.impl, false)
	if err != nil {
		return nil, err
	}
	err = c.impl.GetLatestValue(ctx, bc, request.Method, params, retVal)
	if err != nil {
		return nil, err
	}

	encodedRetVal, err := encodeVersionedBytes(retVal, CurrentEncodingVersion)
	if err != nil {
		return nil, err
	}

	return &pb.GetLatestValueReply{RetVal: encodedRetVal}, nil
}

func getEncodedType(itemType string, possibleTypeProvider any, forEncoding bool) (any, error) {
	if rc, ok := possibleTypeProvider.(types.TypeProvider); ok {
		return rc.CreateType(itemType, forEncoding)
	}

	return &map[string]any{}, nil
}

func unwrapClientError(err error) error {
	if err == nil {
		return nil
	}
	errTypes := []error{
		types.ErrInvalidType,
		types.ErrFieldNotFound,
		types.ErrInvalidEncoding,
		types.ErrWrongNumberOfElements,
		types.ErrNotASlice,
		types.ErrUnknown,
	}

	s, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("%w: %w", types.ErrUnknown, err)
	}

	msg := s.Message()
	for _, etype := range errTypes {
		if msg == etype.Error() {
			return etype
		} else if strings.HasPrefix(msg, etype.Error()+":") {
			rest := strings.SplitN(msg, ":", 2)[1]
			return fmt.Errorf("%w: %w", etype, errors.New(rest))
		}
	}

	return fmt.Errorf("%w: %w", types.ErrUnknown, err)
}
