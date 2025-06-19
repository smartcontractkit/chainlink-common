package gateway

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/gatewayconnector"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ core.GatewayConnectorHandler = (*GatewayConnectorHandlerClient)(nil)

type GatewayConnectorHandlerClient struct {
	grpc pb.GatewayConnectorHandlerClient
}

func (c GatewayConnectorHandlerClient) ID(ctx context.Context) (string, error) {
	resp, err := c.grpc.Id(ctx, &emptypb.Empty{})
	if err != nil {
		return "", fmt.Errorf("failed to get handler info: %w", err)
	}
	return resp.Id, nil
}

func (c GatewayConnectorHandlerClient) HandleGatewayMessage(ctx context.Context, gatewayID string, msg []byte) error {
	_, err := c.grpc.HandleGatewayMessage(ctx, &pb.SendMessageRequest{
		GatewayId: gatewayID,
		Message:   msg,
	})
	return err
}

func NewGatewayConnectorHandlerClient(cc grpc.ClientConnInterface) *GatewayConnectorHandlerClient {
	return &GatewayConnectorHandlerClient{pb.NewGatewayConnectorHandlerClient(cc)}
}

var _ pb.GatewayConnectorHandlerServer = (*GatewayConnectorHandlerServer)(nil)

type GatewayConnectorHandlerServer struct {
	pb.UnimplementedGatewayConnectorHandlerServer
	*net.BrokerExt
	impl core.GatewayConnectorHandler
}

func NewGatewayConnectorHandlerServer(impl core.GatewayConnectorHandler) *GatewayConnectorHandlerServer {
	return &GatewayConnectorHandlerServer{impl: impl}
}

func (s GatewayConnectorHandlerServer) HandleGatewayMessage(ctx context.Context, req *pb.SendMessageRequest) (*emptypb.Empty, error) {
	if err := s.impl.HandleGatewayMessage(ctx, req.GatewayId, req.Message); err != nil {
		return nil, fmt.Errorf("failed to handle gateway message: %s: %w", req.GatewayId, err)
	}
	return &emptypb.Empty{}, nil
}
