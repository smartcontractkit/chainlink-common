package internal

import (
	"context"
	"encoding/json"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

var _ types.ChainReader = (*chainReaderClient)(nil)

type chainReaderClient struct {
	*brokerExt
	grpc pb.ChainReaderClient
}

func (c *chainReaderClient) GetLatestValue(ctx context.Context, bc types.BoundContract, method string, params, retValues any) ([]byte, error) {
	boundContract := pb.BoundContract{Name: bc.Name, Address: bc.Address, Pending: bc.Pending}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	jsonRetValues, err := json.Marshal(retValues)
	if err != nil {
		return nil, err
	}

	reply, err := c.grpc.GetLatestValue(ctx, &pb.GetLatestValueRequest{Bc: &boundContract, Method: method, Params: jsonParams, ReturnValues: jsonRetValues})
	if err != nil {
		return nil, err
	}
	return reply.RetValues, nil
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
	retValues, err := c.impl.GetLatestValue(ctx, bc, request.Method, request.Params, request.ReturnValues)
	if err != nil {
		return nil, err
	}
	return &pb.GetLatestValueReply{RetValues: retValues}, nil
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
