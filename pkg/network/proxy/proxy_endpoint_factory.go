package proxy

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProxyEndpointFactory struct {
	peerId    string
	proxyAddr string
	client    BinaryNetworkEndpointProxyClient
	conn      *grpc.ClientConn
}

func (f *ProxyEndpointFactory) PeerID() string {
	return f.peerId
}

type ClosableBinaryNetworkEndpointFactory interface {
	types.BinaryNetworkEndpointFactory
	io.Closer
}

func NewProxyEndpointFactory(peerId, proxyAddr string) (ClosableBinaryNetworkEndpointFactory, error) {
	conn, err := grpc.NewClient(proxyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}

	client := NewBinaryNetworkEndpointProxyClient(conn)

	return &ProxyEndpointFactory{
		peerId:    peerId,
		proxyAddr: proxyAddr,
		client:    client,
		conn:      conn,
	}, nil
}

func (f *ProxyEndpointFactory) Close() error {
	return f.conn.Close()
}

func (f *ProxyEndpointFactory) NewEndpoint(
	cd types.ConfigDigest,
	peerIDs []string,
	v2bootstrappers []commontypes.BootstrapperLocator,
	failureThreshold int,
	limits types.BinaryNetworkEndpointLimits,
) (commontypes.BinaryNetworkEndpoint, error) {
	return &ProxyEndpoint{
		factory:          f,
		configDigest:     cd[:],
		peerIDs:          peerIDs,
		bootstrappers:    v2bootstrappers,
		failureThreshold: failureThreshold,
		limits:           limits,
		recvChan:         make(chan commontypes.BinaryMessageWithSender, 100000),
		sendReqChan:      make(chan *BinaryNetworkClientRequest, 10000),
	}, nil
}

type ProxyEndpoint struct {
	factory          *ProxyEndpointFactory
	configDigest     []byte
	peerIDs          []string
	bootstrappers    []commontypes.BootstrapperLocator
	failureThreshold int
	limits           types.BinaryNetworkEndpointLimits

	mu          sync.Mutex
	started     bool
	recvChan    chan commontypes.BinaryMessageWithSender
	sendReqChan chan *BinaryNetworkClientRequest
	stream      BinaryNetworkEndpointProxy_ConnectClient
	cancelRecv  context.CancelFunc
	wg          sync.WaitGroup
}

func (e *ProxyEndpoint) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.started {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.cancelRecv = cancel

	stream, err := e.factory.client.Connect(ctx)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect: %w", err)
	}
	e.stream = stream

	pbBootstrappers := make([]*BootstrapperLocator, len(e.bootstrappers))
	for i, b := range e.bootstrappers {
		pbBootstrappers[i] = &BootstrapperLocator{
			PeerId: b.PeerID,
			Addrs:  b.Addrs,
		}
	}

	err = stream.Send(&BinaryNetworkClientRequest{
		Message: &BinaryNetworkClientRequest_NewEndpoint{
			NewEndpoint: &NewEndpointRequest{
				ConfigDigest:     e.configDigest,
				PeerIds:          e.peerIDs,
				V2Bootstrappers:  pbBootstrappers,
				FailureThreshold: int32(e.failureThreshold),
				Limits: &BinaryNetworkEndpointLimits{
					MaxMessageLength:          int32(e.limits.MaxMessageLength),
					MessagesRatePerOracle:     e.limits.MessagesRatePerOracle,
					MessagesCapacityPerOracle: int32(e.limits.MessagesCapacityPerOracle),
					BytesRatePerOracle:        e.limits.BytesRatePerOracle,
					BytesCapacityPerOracle:    int32(e.limits.BytesCapacityPerOracle),
				},
			},
		},
	})
	if err != nil {
		cancel()
		e.stream = nil
		return fmt.Errorf("failed to send new endpoint: %w", err)
	}

	// Capture stream in local variable to avoid race conditions
	streamForGoroutines := e.stream

	e.wg.Add(2)
	go func() {
		defer e.wg.Done()
		e.receiveLoop(ctx, streamForGoroutines)
	}()

	go func() {
		defer e.wg.Done()
		e.sendLoop(ctx, streamForGoroutines)
	}()

	e.started = true
	return nil
}

func (e *ProxyEndpoint) receiveLoop(ctx context.Context, stream BinaryNetworkEndpointProxy_ConnectClient) {
	defer func() {
		// Use recover to avoid panic if channel is already closed
		defer func() { recover() }()
		close(e.recvChan)
	}()

	// Check if stream is nil (shouldn't happen, but be defensive)
	if stream == nil {
		return
	}

	for {
		// Recv() will return when context is cancelled (stream was created with ctx)
		msg, err := stream.Recv()
		if err != nil {
			return
		}

		if msg == nil {
			return
		}

		select {
		case e.recvChan <- commontypes.BinaryMessageWithSender{
			Msg:    msg.Msg,
			Sender: commontypes.OracleID(msg.Sender),
		}:
		case <-ctx.Done():
			return
		default:
			fmt.Println("*** dropping receive")
		}
	}
}

func (e *ProxyEndpoint) sendLoop(ctx context.Context, stream BinaryNetworkEndpointProxy_ConnectClient) {
	// Check if stream is nil (shouldn't happen, but be defensive)
	if stream == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-e.sendReqChan:
			if !ok {
				return
			}
			if req == nil {
				continue
			}
			if err := stream.Send(req); err != nil {
				return
			}
		}
	}
}

func (e *ProxyEndpoint) Close() error {
	e.mu.Lock()
	if !e.started {
		e.mu.Unlock()
		return nil
	}

	if e.cancelRecv != nil {
		e.cancelRecv()
	}

	if e.sendReqChan != nil {
		// Use recover to avoid panic if channel is already closed
		func() {
			defer func() { recover() }()
			close(e.sendReqChan)
		}()
		// Set to nil to prevent sending to closed channel
		e.sendReqChan = nil
	}

	e.started = false
	e.mu.Unlock()

	// Wait for goroutines to finish - they should exit quickly after context cancellation
	e.wg.Wait()

	return nil
}

func (e *ProxyEndpoint) SendTo(msg []byte, to commontypes.OracleID) {
	e.mu.Lock()
	sendReqChan := e.sendReqChan
	started := e.started
	e.mu.Unlock()

	if !started || sendReqChan == nil {
		return
	}

	req := &BinaryNetworkClientRequest{
		Message: &BinaryNetworkClientRequest_SendTo{
			SendTo: &SendToRequest{
				Payload:    msg,
				ToOracleId: uint32(to),
			},
		},
	}

	select {
	case sendReqChan <- req:
	default:
		fmt.Println("*** dropping send")
	}
}

func (e *ProxyEndpoint) Broadcast(msg []byte) {
	e.mu.Lock()
	sendReqChan := e.sendReqChan
	started := e.started
	e.mu.Unlock()

	if !started || sendReqChan == nil {
		return
	}

	req := &BinaryNetworkClientRequest{
		Message: &BinaryNetworkClientRequest_Broadcast{
			Broadcast: msg,
		},
	}

	select {
	case sendReqChan <- req:
	default:
		fmt.Println("*** dropping broadcast")
	}
}

func (e *ProxyEndpoint) Receive() <-chan commontypes.BinaryMessageWithSender {
	return e.recvChan
}
