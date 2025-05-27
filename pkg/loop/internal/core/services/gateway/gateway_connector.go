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

var _ core.GatewayConnector = (*GatewayConnectorClient)(nil)

type GatewayConnectorClient struct {
	*net.BrokerExt
	grpc pb.GatewayConnectorClient
}

func (c GatewayConnectorClient) Start(ctx context.Context) error {
	_, err := c.grpc.Start(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to start gateway connector: %w", err)
	}
	return nil
}

func (c GatewayConnectorClient) Close() error {
	_, err := c.grpc.Close(context.Background(), &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to close gateway connector: %w", err)
	}
	return nil
}

func (c GatewayConnectorClient) AddHandler(methods []string, handler core.GatewayConnectorHandler) error {
	handlerID, err := handler.ID()
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

	_, err = c.grpc.AddHandler(context.Background(), &pb.AddHandlerRequest{
		HandlerId: id,
		Methods:   methods,
	})
	if err == nil {
		cRes.Close()
		return err
	}
	return nil
}

func (c GatewayConnectorClient) GatewayIDs() ([]string, error) {
	resp, err := c.grpc.GatewayIDs(context.Background(), &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway IDs: %w", err)
	}
	gatewayIDs := make([]string, len(resp.GatewayIds))
	copy(gatewayIDs, resp.GatewayIds)
	return gatewayIDs, nil
}

func (c GatewayConnectorClient) DonID() (string, error) {
	resp, err := c.grpc.DonID(context.Background(), &emptypb.Empty{})
	if err != nil {
		return "", fmt.Errorf("failed to get DON ID: %w", err)
	}
	return resp.DonId, nil
}

func (c GatewayConnectorClient) AwaitConnection(ctx context.Context, gatewayID string) error {
	_, err := c.grpc.AwaitConnection(ctx, &pb.GatewayIDRequest{GatewayId: gatewayID})
	return err
}

func (c GatewayConnectorClient) SendToGateway(ctx context.Context, gatewayID string, msg *gateway.Message) error {
	_, err := c.grpc.SendToGateway(ctx, &pb.SendMessageRequest{
		GatewayId: gatewayID,
		Message:   toPbMessage(msg),
	})
	return err
}

func (c GatewayConnectorClient) SignAndSendToGateway(ctx context.Context, gatewayID string, body *gateway.MessageBody) error {
	_, err := c.grpc.SignAndSendToGateway(ctx, &pb.SignAndSendMessageRequest{
		GatewayId: gatewayID,
		Body:      toPbMessageBody(body),
	})
	return err
}

func NewGatewayConnectorClient(cc grpc.ClientConnInterface, b *net.BrokerExt) *GatewayConnectorClient {
	return &GatewayConnectorClient{
		grpc:      pb.NewGatewayConnectorClient(cc),
		BrokerExt: b.WithName("GatewayConnectorClient"),
	}
}

func toPbMessage(msg *gateway.Message) *pb.Message {
	return &pb.Message{
		Body:      toPbMessageBody(&msg.Body),
		Signature: msg.Signature,
	}
}

func toPbMessageBody(msg *gateway.MessageBody) *pb.MessageBody {
	return &pb.MessageBody{
		MessageId: msg.MessageId,
		Method:    msg.Method,
		DonId:     msg.DonId,
		Receiver:  msg.Receiver,
		Payload:   msg.Payload,
		Sender:    msg.Sender,
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
		return nil, fmt.Errorf("failed to dial handler: %s: %w", req.HandlerId, err)
	}
	client := NewGatewayConnectorHandlerClient(conn)
	err = s.impl.AddHandler(req.Methods, client)
	if err != nil {
		return nil, fmt.Errorf("failed to add handler: %s: %w", req.HandlerId, err)
	}
	return &emptypb.Empty{}, nil
}

func (s GatewayConnectorServer) SendToGateway(ctx context.Context, req *pb.SendMessageRequest) (*emptypb.Empty, error) {
	if err := s.impl.SendToGateway(ctx, req.GatewayId, fromPbMessage(req.Message)); err != nil {
		return nil, fmt.Errorf("failed to send message to gateway: %s: %w", req.GatewayId, err)
	}
	return &emptypb.Empty{}, nil
}
func (s GatewayConnectorServer) SignAndSendToGateway(ctx context.Context, req *pb.SignAndSendMessageRequest) (*emptypb.Empty, error) {
	if err := s.impl.SignAndSendToGateway(ctx, req.GatewayId, fromPbMessageBody(req.Body)); err != nil {
		return nil, fmt.Errorf("failed to send message to gateway: %s: %w", req.GatewayId, err)
	}
	return &emptypb.Empty{}, nil
}
func (s GatewayConnectorServer) GatewayIDs(ctx context.Context, _ *emptypb.Empty) (*pb.GatewayIDsReply, error) {
	gatewayIDs, err := s.impl.GatewayIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway IDs: %w", err)
	}
	return &pb.GatewayIDsReply{GatewayIds: gatewayIDs}, nil
}

func (s GatewayConnectorServer) DonID(ctx context.Context, _ *emptypb.Empty) (*pb.DonIDReply, error) {
	donID, err := s.impl.DonID()
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

func fromPbMessage(msg *pb.Message) *gateway.Message {
	body := fromPbMessageBody(msg.Body)
	return &gateway.Message{
		Signature: msg.Signature,
		Body:      *body,
	}
}

func fromPbMessageBody(msgBody *pb.MessageBody) *gateway.MessageBody {
	return &gateway.MessageBody{
		MessageId: msgBody.MessageId,
		Method:    msgBody.Method,
		DonId:     msgBody.DonId,
		Receiver:  msgBody.Receiver,
		Payload:   msgBody.Payload,
		Sender:    msgBody.Sender,
	}
}
