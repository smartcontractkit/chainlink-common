package gatewayconnector

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/gatewayconnector"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/types/gateway"
)

var _ core.GatewayConnector = (*Client)(nil)

type Client struct {
	grpc pb.GatewayConnectorClient
}

func (c Client) GatewayIDs(ctx context.Context) ([]string, error) {
	resp, err := c.grpc.GatewayIDs(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway IDs: %w", err)
	}
	gatewayIDs := make([]string, len(resp.GatewayIds))
	for i, id := range resp.GatewayIds {
		gatewayIDs[i] = id
	}
	return gatewayIDs, nil
}

func (c Client) DonID(ctx context.Context) (string, error) {
	resp, err := c.grpc.DonID(ctx, &emptypb.Empty{})
	if err != nil {
		return "", fmt.Errorf("failed to get DON ID: %w", err)
	}
	return resp.DonId, nil
}

func (c Client) AwaitConnection(ctx context.Context, gatewayID string) error {
	_, err := c.grpc.AwaitConnection(ctx, &pb.GatewayIDRequest{GatewayId: gatewayID})
	return err
}

func (c Client) SendToGateway(ctx context.Context, gatewayID string, msg *gateway.Message) error {
	_, err := c.grpc.SendToGateway(ctx, &pb.SendMessageRequest{
		GatewayId: gatewayID,
		Message:   toPbMessage(msg),
	})
	return err
}

func (c Client) SignAndSendToGateway(ctx context.Context, gatewayID string, body *gateway.MessageBody) error {
	_, err := c.grpc.SignAndSendToGateway(ctx, &pb.SignAndSendMessageRequest{
		GatewayId: gatewayID,
		Body:      toPbMessageBody(body),
	})
	return err
}

func NewClient(cc grpc.ClientConnInterface) *Client {
	return &Client{pb.NewGatewayConnectorClient(cc)}
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
