package network

import (
	"context"
	"errors"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/libocr/commontypes"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/ocr/network/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// RegisterServer serves the BinaryNetworkEndpointFactory factory returns on s. factory is asked on each call, so it can
// lend a peer that only exists once the node has started.
func RegisterServer(s grpc.ServiceRegistrar, lggr logger.Logger,
	factory func(context.Context) (ocrtypes.BinaryNetworkEndpointFactory, error)) {
	pb.RegisterNetworkServer(s, &server{lggr: logger.Named(lggr, "OCRNetworkServer"), factory: factory})
}

type server struct {
	pb.UnimplementedNetworkServer
	lggr    logger.Logger
	factory func(context.Context) (ocrtypes.BinaryNetworkEndpointFactory, error)
}

func (s *server) Connect(stream pb.Network_ConnectServer) error {
	ctx := stream.Context()
	first, err := stream.Recv()
	if err != nil {
		return err
	}

	open := first.GetOpen()
	if open == nil {
		return status.Error(codes.InvalidArgument, "the first request must open the endpoint")
	}

	digest, err := ocrtypes.BytesToConfigDigest(open.ConfigDigest)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid config digest: %v", err)
	}

	factory, err := s.factory(ctx)
	if err != nil {
		return status.Errorf(codes.Unavailable, "failed to get the node's OCR network: %v", err)
	}

	endpoint, err := factory.NewEndpoint(digest, open.PeerIds,
		bootstrappersFromPB(open.Bootstrappers), int(open.F), limitsFromPB(open.Limits))
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "failed to create endpoint: %v", err)
	}

	defer func() {
		if cerr := endpoint.Close(); cerr != nil {
			s.lggr.Warnw("Failed to close endpoint", "configDigest", digest, "err", cerr)
		}
	}()

	if err = endpoint.Start(); err != nil {
		return status.Errorf(codes.FailedPrecondition, "failed to start endpoint: %v", err)
	}

	if err = stream.Send(&pb.EndpointReply{Reply: &pb.EndpointReply_Opened{Opened: &pb.Opened{}}}); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	received := make(chan error, 1)
	go func() { received <- forwardReceived(ctx, endpoint, stream) }()
	err = forwardSent(endpoint, stream)
	cancel()
	return errors.Join(err, <-received)
}

func forwardSent(endpoint commontypes.BinaryNetworkEndpoint, stream pb.Network_ConnectServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) || status.Code(err) == codes.Canceled {
				return nil
			}

			return err
		}

		switch r := req.Request.(type) {
		case *pb.EndpointRequest_SendTo:
			endpoint.SendTo(r.SendTo.Payload, commontypes.OracleID(r.SendTo.To)) //nolint:gosec // oracle IDs fit in a uint8
		case *pb.EndpointRequest_Broadcast:
			endpoint.Broadcast(r.Broadcast)
		default:
			return status.Errorf(codes.InvalidArgument, "unexpected request %T after the endpoint is open", req.Request)
		}
	}
}

// forwardReceived does not drop: a slow reader holds the stream back, and the endpoint drops what it cannot buffer,
// as it does for a slow oracle.
func forwardReceived(ctx context.Context, endpoint commontypes.BinaryNetworkEndpoint, stream pb.Network_ConnectServer) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-endpoint.Receive():
			if !ok {
				return nil
			}

			err := stream.Send(&pb.EndpointReply{Reply: &pb.EndpointReply_Message{Message: &pb.Message{
				Payload: msg.Msg,
				Sender:  uint32(msg.Sender),
			}}})
			if err != nil {
				return err
			}
		}
	}
}

func bootstrappersToPB(locators []commontypes.BootstrapperLocator) []*pb.BootstrapperLocator {
	out := make([]*pb.BootstrapperLocator, len(locators))
	for i, l := range locators {
		out[i] = &pb.BootstrapperLocator{PeerId: l.PeerID, Addrs: l.Addrs}
	}

	return out
}

func bootstrappersFromPB(locators []*pb.BootstrapperLocator) []commontypes.BootstrapperLocator {
	out := make([]commontypes.BootstrapperLocator, len(locators))
	for i, l := range locators {
		out[i] = commontypes.BootstrapperLocator{PeerID: l.PeerId, Addrs: l.Addrs}
	}

	return out
}

func limitsToPB(l ocrtypes.BinaryNetworkEndpointLimits) *pb.EndpointLimits {
	return &pb.EndpointLimits{
		MaxMessageLength:          int64(l.MaxMessageLength),
		MessagesRatePerOracle:     l.MessagesRatePerOracle,
		MessagesCapacityPerOracle: int64(l.MessagesCapacityPerOracle),
		BytesRatePerOracle:        l.BytesRatePerOracle,
		BytesCapacityPerOracle:    int64(l.BytesCapacityPerOracle),
	}
}

func limitsFromPB(l *pb.EndpointLimits) ocrtypes.BinaryNetworkEndpointLimits {
	return ocrtypes.BinaryNetworkEndpointLimits{
		MaxMessageLength:          int(l.GetMaxMessageLength()),
		MessagesRatePerOracle:     l.GetMessagesRatePerOracle(),
		MessagesCapacityPerOracle: int(l.GetMessagesCapacityPerOracle()),
		BytesRatePerOracle:        l.GetBytesRatePerOracle(),
		BytesCapacityPerOracle:    int(l.GetBytesCapacityPerOracle()),
	}
}
