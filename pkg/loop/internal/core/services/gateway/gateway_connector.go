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

var _ core.GatewayConnector = (*GatewayConnectorClient)(nil)

type GatewayConnectorClient struct {
	*net.BrokerExt
	grpc pb.GatewayConnectorClient
}

func (c GatewayConnectorClient) AddHandler(ctx context.Context, methods []string, handler core.GatewayConnectorHandler) error {
	handlerID, err := handler.ID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get handler info: %w", err)
	}
	gatewayConnectorServer := NewGatewayConnectorHandlerServer(handler)

	var cRes net.Resource
	id, cRes, err := c.ServeNew(handlerID, func(s *grpc.Server) {
		pb.RegisterGatewayConnectorHandlerServer(s, gatewayConnectorServer)
	})
	if err != nil {
		return fmt.Errorf("failed to serve handler: %s: %w", handlerID, err)
	}

	_, err = c.grpc.AddHandler(ctx, &pb.AddHandlerRequest{
		HandlerId: id,
		Methods:   methods,
	})
	if err != nil {
		cRes.Close()
		return fmt.Errorf("failed to add handler: %w", err)
	}
	return nil
}

func (c GatewayConnectorClient) GatewayIDs(ctx context.Context) ([]string, error) {
	resp, err := c.grpc.GatewayIDs(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway IDs: %w", err)
	}
	gatewayIDs := make([]string, len(resp.GatewayIds))
	copy(gatewayIDs, resp.GatewayIds)
	return gatewayIDs, nil
}

func (c GatewayConnectorClient) DonID(ctx context.Context) (string, error) {
	resp, err := c.grpc.DonID(ctx, &emptypb.Empty{})
	if err != nil {
		return "", fmt.Errorf("failed to get DON ID: %w", err)
	}
	return resp.DonId, nil
}

func (c GatewayConnectorClient) AwaitConnection(ctx context.Context, gatewayID string) error {
	_, err := c.grpc.AwaitConnection(ctx, &pb.GatewayIDRequest{GatewayId: gatewayID})
	return err
}

func (c GatewayConnectorClient) SendToGateway(ctx context.Context, gatewayID string, msg []byte) error {
	_, err := c.grpc.SendToGateway(ctx, &pb.SendMessageRequest{
		GatewayId: gatewayID,
		Message:   msg,
	})
	return err
}

func (c GatewayConnectorClient) Sign(ctx context.Context, msg []byte) ([]byte, error) {
	signReply, err := c.grpc.Sign(ctx, &pb.SignRequest{
		Message: msg,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}
	return signReply.Signature, nil
}

func NewGatewayConnectorClient(cc grpc.ClientConnInterface, b *net.BrokerExt) *GatewayConnectorClient {
	return &GatewayConnectorClient{
		grpc:      pb.NewGatewayConnectorClient(cc),
		BrokerExt: b.WithName("GatewayConnectorClient"),
	}
}

var _ pb.GatewayConnectorServer = (*GatewayConnectorServer)(nil)

type GatewayConnectorServer struct {
	pb.UnimplementedGatewayConnectorServer
	*net.BrokerExt
	impl core.GatewayConnector
}

func NewGatewayConnectorServer(b *net.BrokerExt, impl core.GatewayConnector) *GatewayConnectorServer {
	return &GatewayConnectorServer{
		BrokerExt: b.WithName("GatewayConnectorServer"),
		impl:      impl,
	}
}

func (s GatewayConnectorServer) AddHandler(ctx context.Context, req *pb.AddHandlerRequest) (*emptypb.Empty, error) {
	conn, err := s.Dial(req.HandlerId)
	if err != nil {
		return nil, fmt.Errorf("failed to dial handler: %d: %w", req.HandlerId, err)
	}
	client := NewGatewayConnectorHandlerClient(conn)
	err = s.impl.AddHandler(ctx, req.Methods, client)
	if err != nil {
		return nil, fmt.Errorf("failed to add handler: %d: %w", req.HandlerId, err)
	}
	return &emptypb.Empty{}, nil
}

func (s GatewayConnectorServer) SendToGateway(ctx context.Context, req *pb.SendMessageRequest) (*emptypb.Empty, error) {
	if err := s.impl.SendToGateway(ctx, req.GatewayId, req.Message); err != nil {
		return nil, fmt.Errorf("failed to send message to gateway: %s: %w", req.GatewayId, err)
	}
	return &emptypb.Empty{}, nil
}
func (s GatewayConnectorServer) Sign(ctx context.Context, req *pb.SignRequest) (*pb.SignReply, error) {
	signature, err := s.impl.Sign(ctx, req.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}
	return &pb.SignReply{
		Signature: signature,
	}, nil
}
func (s GatewayConnectorServer) GatewayIDs(ctx context.Context, _ *emptypb.Empty) (*pb.GatewayIDsReply, error) {
	gatewayIDs, err := s.impl.GatewayIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway IDs: %w", err)
	}
	return &pb.GatewayIDsReply{GatewayIds: gatewayIDs}, nil
}

func (s GatewayConnectorServer) DonID(ctx context.Context, _ *emptypb.Empty) (*pb.DonIDReply, error) {
	donID, err := s.impl.DonID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DON ID: %w", err)
	}
	return &pb.DonIDReply{DonId: donID}, nil
}
func (s GatewayConnectorServer) AwaitConnection(ctx context.Context, req *pb.GatewayIDRequest) (*emptypb.Empty, error) {
	if err := s.impl.AwaitConnection(ctx, req.GatewayId); err != nil {
		return nil, fmt.Errorf("failed to await connection to gateway: %s: %w", req.GatewayId, err)
	}
	return &emptypb.Empty{}, nil
}
