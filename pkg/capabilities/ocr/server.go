package ocr

import (
	"context"
	"errors"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	ragetypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/ocr/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// MaxMessageSize is the largest gRPC message between a host and a process: an OCR message as large as libocr allows,
// with room for its framing. gRPC's 4 MiB default would break an endpoint's stream on a large message, such as a
// leader's proposal carrying many observations; the endpoint on the host still enforces the oracle's own limits.
const MaxMessageSize = ragetypes.MaxMessageLength + 1024

// ServerOptions are the options a server [RegisterServer] serves on needs, so that it receives what libocr sends.
func ServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{grpc.MaxRecvMsgSize(MaxMessageSize)}
}

// RegisterServer serves the Node node returns on s, which must be made with [ServerOptions]. node is asked on each call,
// so it can lend a peer that only exists once the node has started.
func RegisterServer(s grpc.ServiceRegistrar, lggr logger.Logger, node func(context.Context) (Node, error)) {
	pb.RegisterOCRServer(s, &server{lggr: logger.Named(lggr, "OCRServer"), node: node})
}

type server struct {
	pb.UnimplementedOCRServer
	lggr logger.Logger
	node func(context.Context) (Node, error)
}

func (s *server) getNode(ctx context.Context) (Node, error) {
	node, err := s.node(ctx)
	if err != nil {
		return Node{}, status.Errorf(codes.Unavailable, "failed to get the node's OCR: %v", err)
	}

	return node, nil
}

func (s *server) PublicKeys(ctx context.Context, _ *emptypb.Empty) (*pb.PublicKeysReply, error) {
	node, err := s.getNode(ctx)
	if err != nil {
		return nil, err
	}

	offchainPublicKey := node.OffchainKeyring.OffchainPublicKey()
	configPublicKey := node.OffchainKeyring.ConfigEncryptionPublicKey()
	return &pb.PublicKeysReply{
		OffchainPublicKey:         offchainPublicKey[:],
		ConfigEncryptionPublicKey: configPublicKey[:],
		OnchainPublicKey:          node.OnchainKeyring.PublicKey(),
		MaxSignatureLength:        uint32(node.OnchainKeyring.MaxSignatureLength()), //nolint:gosec // signature lengths are small
	}, nil
}

func (s *server) Connect(stream pb.OCR_ConnectServer) error {
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

	node, err := s.getNode(ctx)
	if err != nil {
		return err
	}

	endpoint, err := node.BinaryNetworkEndpointFactory.NewEndpoint(digest, open.PeerIds,
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

func forwardSent(endpoint commontypes.BinaryNetworkEndpoint, stream pb.OCR_ConnectServer) error {
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
func forwardReceived(ctx context.Context, endpoint commontypes.BinaryNetworkEndpoint, stream pb.OCR_ConnectServer) error {
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

func (s *server) OffchainSign(ctx context.Context, req *pb.OffchainSignRequest) (*pb.SignatureReply, error) {
	node, err := s.getNode(ctx)
	if err != nil {
		return nil, err
	}

	sig, err := node.OffchainKeyring.OffchainSign(req.Message)
	if err != nil {
		return nil, err
	}

	return &pb.SignatureReply{Signature: sig}, nil
}

func (s *server) ConfigDiffieHellman(ctx context.Context, req *pb.ConfigDiffieHellmanRequest) (*pb.ConfigDiffieHellmanReply, error) {
	var point [32]byte
	if len(req.Point) != len(point) {
		return nil, status.Errorf(codes.InvalidArgument, "point is %d bytes, want %d", len(req.Point), len(point))
	}

	copy(point[:], req.Point)
	node, err := s.getNode(ctx)
	if err != nil {
		return nil, err
	}

	secret, err := node.OffchainKeyring.ConfigDiffieHellman(point)
	if err != nil {
		return nil, err
	}

	return &pb.ConfigDiffieHellmanReply{SharedSecret: secret[:]}, nil
}

func (s *server) Sign(ctx context.Context, req *pb.SignRequest) (*pb.SignatureReply, error) {
	digest, err := ocrtypes.BytesToConfigDigest(req.ConfigDigest)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid config digest: %v", err)
	}

	node, err := s.getNode(ctx)
	if err != nil {
		return nil, err
	}

	sig, err := node.OnchainKeyring.Sign(digest, req.SeqNr, reportFromPB(req.Report))
	if err != nil {
		return nil, err
	}

	return &pb.SignatureReply{Signature: sig}, nil
}

func reportFromPB(r *pb.Report) ocr3types.ReportWithInfo[string] {
	return ocr3types.ReportWithInfo[string]{Report: r.GetReport(), Info: r.GetKeyBundle()}
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
