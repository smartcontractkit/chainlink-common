package internal

import (
	"context"
	"encoding/json"
	"fmt"

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
)

// Version to be used for encoding ( version used for decoding is determined by data received )
// These are separate constants in case we want to upgrade their data formats independently
const ParamsCurrentEncodingVersion = SimpleJsonEncodingVersion
const RetvalCurrentEncodingVersion = SimpleJsonEncodingVersion

func encodeVersionedBytes(data any, version int32) (*pb.VersionedBytes, error) {
	var jsonData []byte
	var err error

	switch version {
	case SimpleJsonEncodingVersion:
		jsonData, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unsupported encoding version %d for data %v", version, data)
	}

	return &pb.VersionedBytes{Version: SimpleJsonEncodingVersion, Data: jsonData}, nil
}

func decodeVersionedBytes(res any, vData *pb.VersionedBytes) error {
	switch vData.Version {
	case SimpleJsonEncodingVersion:
		err := json.Unmarshal(vData.Data, res)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported encoding version %d for versionedData %v", vData.Version, vData.Data)
	}

	return nil
}

func (c *chainReaderClient) GetLatestValue(ctx context.Context, bc types.BoundContract, method string, params, retVal any) error {
	versionedParams, err := encodeVersionedBytes(params, ParamsCurrentEncodingVersion)
	if err != nil {
		return err
	}

	boundContract := pb.BoundContract{Name: bc.Name, Address: bc.Address, Pending: bc.Pending}

	reply, err := c.grpc.GetLatestValue(ctx, &pb.GetLatestValueRequest{Bc: &boundContract, Method: method, Params: versionedParams})
	if err != nil {
		return err
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

	var paramsMap map[string]any
	err := decodeVersionedBytes(&paramsMap, request.Params)
	if err != nil {
		return nil, err
	}

	var retVal map[string]any
	err = c.impl.GetLatestValue(ctx, bc, request.Method, paramsMap, &retVal)
	if err != nil {
		return nil, err
	}

	jsonRetVal, err := encodeVersionedBytes(&retVal, RetvalCurrentEncodingVersion)
	if err != nil {
		return nil, err
	}

	return &pb.GetLatestValueReply{RetVal: jsonRetVal}, nil
}

func (c *chainReaderServer) RegisterEventFilter(ctx context.Context, in *pb.RegisterEventFilterRequest) (*pb.RegisterEventFilterReply, error) {
	return nil, nil
}
func (c *chainReaderServer) UnregisterEventFilter(ctx context.Context, in *pb.UnregisterEventFilterRequest) (*pb.RegisterEventFilterReply, error) {
	return nil, nil
}
func (c *chainReaderServer) QueryEvents(ctx context.Context, in *pb.QueryEventsRequest) (*pb.QueryEventsReply, error) {
	return nil, nil
}
