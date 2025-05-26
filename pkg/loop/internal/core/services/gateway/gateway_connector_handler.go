package gateway

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/gatewayconnector"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/types/gateway"
)

var _ core.GatewayConnectorHandler = (*GatewayConnectorHandlerClient)(nil)

type GatewayConnectorHandlerClient struct {
	grpc pb.GatewayConnectorHandlerClient
}

func (c GatewayConnectorHandlerClient) Start(ctx context.Context) error {
	_, err := c.grpc.Start(ctx, &emptypb.Empty{})
	return err
}

func (c GatewayConnectorHandlerClient) Close() error {
	_, err := c.grpc.Close(context.Background(), &emptypb.Empty{})
	return err
}

func (c GatewayConnectorHandlerClient) Info() (core.GatewayConnectorHandlerInfo, error) {
	resp, err := c.grpc.Info(context.Background(), &emptypb.Empty{})
	if err != nil {
		return core.GatewayConnectorHandlerInfo{}, fmt.Errorf("failed to get handler info: %w", err)
	}
	return core.GatewayConnectorHandlerInfo{
		ID: resp.Id,
	}, nil
}

func (c GatewayConnectorHandlerClient) HandleGatewayMessage(ctx context.Context, gatewayID string, msg *gateway.Message) error {
	_, err := c.grpc.HandleGatewayMessage(ctx, &pb.SendMessageRequest{
		GatewayId: gatewayID,
		Message:   toPbMessage(msg),
	})
	return err
}

func NewGatewayConnectorHandlerClient(cc grpc.ClientConnInterface) *GatewayConnectorHandlerClient {
	return &GatewayConnectorHandlerClient{pb.NewGatewayConnectorHandlerClient(cc)}
}

var _ pb.GatewayConnectorServer = (*GatewayConnectorServer)(nil)

type GatewayConnectorHandlerServer struct {
	pb.UnimplementedGatewayConnectorHandlerServer
	*net.BrokerExt
	impl core.GatewayConnectorHandler
}

func NewGatewayConnectorHandlerServer(impl core.GatewayConnectorHandler) *GatewayConnectorHandlerServer {
	return &GatewayConnectorHandlerServer{impl: impl}
}

func (s GatewayConnectorHandlerServer) Start(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.impl.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start handler: %w", err)
	}
	return &emptypb.Empty{}, nil
}

func (s GatewayConnectorHandlerServer) Close(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.impl.Close(); err != nil {
		return nil, fmt.Errorf("failed to close handler: %w", err)
	}
	return &emptypb.Empty{}, nil
}

func (s GatewayConnectorHandlerServer) HandleGatewayMessage(ctx context.Context, req *pb.SendMessageRequest) (*emptypb.Empty, error) {
	if err := s.impl.HandleGatewayMessage(ctx, req.GatewayId, fromPbMessage(req.Message)); err != nil {
		return nil, fmt.Errorf("failed to handle gateway message: %s: %w", req.GatewayId, err)
	}
	return &emptypb.Empty{}, nil
}
