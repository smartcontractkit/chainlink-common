package proxy

import (
	"fmt"
	"io"
	"sync"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// Server implements the BinaryNetworkEndpointProxy gRPC service. It is backed
// by a real libocr BinaryNetworkEndpointFactory (i.e. a running rage peer) and
// exposes it over the network so that an out-of-process client (see
// ProxyEndpointFactory) can drive OCR message passing without owning the peer.
//
// Each Connect stream corresponds to exactly one BinaryNetworkEndpoint: the
// first message on the stream must be a NewEndpointRequest, after which the
// stream carries SendTo/Broadcast requests up and received messages down.
type Server struct {
	UnimplementedBinaryNetworkEndpointProxyServer

	peerFactory types.BinaryNetworkEndpointFactory
}

// NewServer returns a Server that serves endpoints created by the given
// factory. peerFactory is typically obtained from a libocr peer, e.g.
// networking.NewPeer(...).OCR2BinaryNetworkEndpointFactory().
func NewServer(peerFactory types.BinaryNetworkEndpointFactory) *Server {
	return &Server{peerFactory: peerFactory}
}

func (s *Server) Connect(stream BinaryNetworkEndpointProxy_ConnectServer) error {
	var closers []io.Closer
	wg := sync.WaitGroup{}

	defer func() {
		for _, c := range closers {
			_ = c.Close()
		}
		wg.Wait()
	}()

	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive initial NewEndpointRequest: %w", err)
	}

	newEndpointReq, ok := req.Message.(*BinaryNetworkClientRequest_NewEndpoint)
	if !ok {
		return fmt.Errorf("first message must be NewEndpointRequest, got %T", req.Message)
	}

	endpoint, err := s.handleNewEndpoint(newEndpointReq.NewEndpoint)
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}
	closers = append(closers, endpoint)

	recvChan := endpoint.Receive()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range recvChan {
			pbMsg := &BinaryMessageWithSender{
				Msg:    msg.Msg,
				Sender: uint32(msg.Sender),
			}
			if err := stream.Send(pbMsg); err != nil {
				return
			}
		}
	}()

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch msg := req.Message.(type) {
		case *BinaryNetworkClientRequest_NewEndpoint:
			return fmt.Errorf("NewEndpointRequest not allowed after initial setup")
		case *BinaryNetworkClientRequest_SendTo:
			endpoint.SendTo(msg.SendTo.Payload, commontypes.OracleID(msg.SendTo.ToOracleId))
		case *BinaryNetworkClientRequest_Broadcast:
			endpoint.Broadcast(msg.Broadcast)
		}
	}
}

func (s *Server) handleNewEndpoint(req *NewEndpointRequest) (commontypes.BinaryNetworkEndpoint, error) {
	bootstrappers := make([]commontypes.BootstrapperLocator, len(req.V2Bootstrappers))
	for i, b := range req.V2Bootstrappers {
		bootstrappers[i] = commontypes.BootstrapperLocator{
			PeerID: b.PeerId,
			Addrs:  b.Addrs,
		}
	}

	if len(req.ConfigDigest) != len(types.ConfigDigest{}) {
		return nil, fmt.Errorf("invalid config digest length: got %d, expected %d", len(req.ConfigDigest), len(types.ConfigDigest{}))
	}
	var configDigest types.ConfigDigest
	copy(configDigest[:], req.ConfigDigest)

	endpoint, err := s.peerFactory.NewEndpoint(
		configDigest,
		req.PeerIds,
		bootstrappers,
		int(req.FailureThreshold),
		types.BinaryNetworkEndpointLimits{
			MaxMessageLength:          int(req.Limits.MaxMessageLength),
			MessagesRatePerOracle:     req.Limits.MessagesRatePerOracle,
			MessagesCapacityPerOracle: int(req.Limits.MessagesCapacityPerOracle),
			BytesRatePerOracle:        req.Limits.BytesRatePerOracle,
			BytesCapacityPerOracle:    int(req.Limits.BytesCapacityPerOracle),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	if err := endpoint.Start(); err != nil {
		return nil, fmt.Errorf("failed to start endpoint: %w", err)
	}

	return endpoint, nil
}
