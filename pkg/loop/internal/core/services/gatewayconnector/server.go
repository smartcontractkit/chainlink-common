package gatewayconnector

import (
	"context"
	"fmt"

	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/gatewayconnector"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/types/gateway"
)

var _ pb.GatewayConnectorServer = (*Server)(nil)

type Server struct {
	pb.UnimplementedGatewayConnectorServer
	impl core.GatewayConnector
}

func NewServer(impl core.GatewayConnector) *Server {
	return &Server{impl: impl}
}

func (s Server) SendToGateway(ctx context.Context, req *pb.SendMessageRequest) (*emptypb.Empty, error) {
	if err := s.impl.SendToGateway(ctx, req.GatewayId, fromPbMessage(req.Message)); err != nil {
		return nil, fmt.Errorf("failed to send message to gateway: %s: %w", req.GatewayId, err)
	}
	return &emptypb.Empty{}, nil
}
func (s Server) SignAndSendToGateway(ctx context.Context, req *pb.SignAndSendMessageRequest) (*emptypb.Empty, error) {
	if err := s.impl.SignAndSendToGateway(ctx, req.GatewayId, fromPbMessageBody(req.Body)); err != nil {
		return nil, fmt.Errorf("failed to send message to gateway: %s: %w", req.GatewayId, err)
	}
	return &emptypb.Empty{}, nil
}
func (s Server) GatewayIDs(ctx context.Context, _ *emptypb.Empty) (*pb.GatewayIDsReply, error) {
	gatewayIDs, err := s.impl.GatewayIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway IDs: %w", err)
	}
	return &pb.GatewayIDsReply{GatewayIds: gatewayIDs}, nil
}

func (s Server) DonID(ctx context.Context, _ *emptypb.Empty) (*pb.DonIDReply, error) {
	donID, err := s.impl.DonID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DON ID: %w", err)
	}
	return &pb.DonIDReply{DonId: donID}, nil
}
func (s Server) AwaitConnection(ctx context.Context, req *pb.GatewayIDRequest) (*emptypb.Empty, error) {
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
