package network

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/libocr/commontypes"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/ocr/network/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// Connect returns the BinaryNetworkEndpointFactory served by [RegisterServer] on conn, for the node whose peer is peerID.
func Connect(lggr logger.Logger, conn grpc.ClientConnInterface, peerID string, cfg Config) (ocrtypes.BinaryNetworkEndpointFactory, error) {
	version, err := resolveVersion(cfg.Version)
	if err != nil {
		return nil, err
	}

	switch version {
	case VersionV1:
		return &endpointFactory{lggr: logger.Named(lggr, "OCRNetworkClient"), client: pb.NewNetworkClient(conn), peerID: peerID}, nil
	default:
		return nil, fmt.Errorf("unknown OCR network version %d", version)
	}
}

// outboundBufferSize is how many messages an endpoint holds for the host before dropping the oldest.
const outboundBufferSize = 1024

var _ ocrtypes.BinaryNetworkEndpointFactory = (*endpointFactory)(nil)

type endpointFactory struct {
	lggr   logger.Logger
	client pb.NetworkClient
	peerID string
}

func (f *endpointFactory) PeerID() string { return f.peerID }

// NewEndpoint opens the endpoint on the host before returning, so that the host's errors, such as an endpoint
// already open for the config digest, are returned here.
func (f *endpointFactory) NewEndpoint(digest ocrtypes.ConfigDigest, peerIDs []string,
	bootstrappers []commontypes.BootstrapperLocator, failureThreshold int,
	limits ocrtypes.BinaryNetworkEndpointLimits) (commontypes.BinaryNetworkEndpoint, error) {
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := f.client.Connect(ctx, grpc.MaxCallRecvMsgSize(MaxMessageSize))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect endpoint: %w", err)
	}

	err = stream.Send(&pb.EndpointRequest{Request: &pb.EndpointRequest_Open{Open: &pb.OpenRequest{
		ConfigDigest:  digest[:],
		PeerIds:       peerIDs,
		Bootstrappers: bootstrappersToPB(bootstrappers),
		F:             int64(failureThreshold),
		Limits:        limitsToPB(limits),
	}}})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to open endpoint: %w", err)
	}

	reply, err := stream.Recv()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to open endpoint: %w", err)
	}

	if reply.GetOpened() == nil {
		cancel()
		return nil, fmt.Errorf("expected the endpoint to be opened, got %T", reply.Reply)
	}

	return &endpoint{
		lggr:     logger.With(f.lggr, "configDigest", digest),
		ctx:      ctx,
		cancel:   cancel,
		stream:   stream,
		outbound: make(chan *pb.EndpointRequest, outboundBufferSize),
		received: make(chan commontypes.BinaryMessageWithSender),
	}, nil
}

var _ commontypes.BinaryNetworkEndpoint = (*endpoint)(nil)

// endpoint is a BinaryNetworkEndpoint on the host, for the life of one stream.
type endpoint struct {
	lggr     logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	stream   pb.Network_ConnectClient
	outbound chan *pb.EndpointRequest
	received chan commontypes.BinaryMessageWithSender

	startOnce sync.Once
	closeOnce sync.Once
	wg        sync.WaitGroup
}

func (e *endpoint) Start() error {
	e.startOnce.Do(func() {
		e.wg.Go(e.sendLoop)
		e.wg.Go(e.receiveLoop)
	})

	return nil
}

func (e *endpoint) Close() error {
	e.closeOnce.Do(func() {
		e.cancel()
		e.wg.Wait()
	})

	return nil
}

func (e *endpoint) SendTo(payload []byte, to commontypes.OracleID) {
	e.enqueue(&pb.EndpointRequest{Request: &pb.EndpointRequest_SendTo{SendTo: &pb.SendToRequest{
		Payload: payload,
		To:      uint32(to),
	}}})
}

func (e *endpoint) Broadcast(payload []byte) {
	e.enqueue(&pb.EndpointRequest{Request: &pb.EndpointRequest_Broadcast{Broadcast: payload}})
}

func (e *endpoint) Receive() <-chan commontypes.BinaryMessageWithSender { return e.received }

// enqueue never blocks, as libocr requires: when the buffer is full, the oldest message is dropped.
func (e *endpoint) enqueue(req *pb.EndpointRequest) {
	for {
		select {
		case e.outbound <- req:
			return
		default:
		}

		select {
		case <-e.outbound:
			e.lggr.Warn("Dropped the oldest outbound OCR message, the buffer is full")
		default:
		}
	}
}

func (e *endpoint) sendLoop() {
	for {
		select {
		case <-e.ctx.Done():
			return
		case req := <-e.outbound:
			if err := e.stream.Send(req); err != nil {
				e.logStreamErr("Failed to send OCR message", err)
				return
			}
		}
	}
}

// receiveLoop blocks on a slow reader rather than dropping, which holds the stream back until the host's
// endpoint drops what it cannot buffer.
func (e *endpoint) receiveLoop() {
	for {
		reply, err := e.stream.Recv()
		if err != nil {
			e.logStreamErr("Failed to receive OCR message", err)
			return
		}

		msg := reply.GetMessage()
		if msg == nil {
			e.lggr.Warnw("Ignoring unexpected reply", "reply", fmt.Sprintf("%T", reply.Reply))
			continue
		}

		select {
		case <-e.ctx.Done():
			return
		case e.received <- commontypes.BinaryMessageWithSender{Msg: msg.Payload, Sender: commontypes.OracleID(msg.Sender)}: //nolint:gosec // oracle IDs fit in a uint8
		}
	}
}

func (e *endpoint) logStreamErr(msg string, err error) {
	if e.ctx.Err() != nil || errors.Is(err, context.Canceled) || status.Code(err) == codes.Canceled {
		return
	}

	e.lggr.Errorw(msg+", the endpoint is closed", "err", err)
}
