package proxy

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// defaultRecvBufferSize is used for a stream's receive channel when the caller
// requests a non-positive incoming buffer size.
const defaultRecvBufferSize = 100

// ProxyPeerGroupFactory is a PeerGroupFactory that delegates to a remote
// PeerGroupProxy server. Consumers that require a libocr
// networking.PeerGroupFactory (e.g. core's Don2DonSharedPeer) wrap this with a
// trivial adapter, since the method sets are identical.
type ProxyPeerGroupFactory struct {
	client PeerGroupProxyClient
	conn   *grpc.ClientConn
}

// ClosablePeerGroupFactory is a PeerGroupFactory whose underlying connection
// can be released.
type ClosablePeerGroupFactory interface {
	PeerGroupFactory
	io.Closer
}

var _ ClosablePeerGroupFactory = (*ProxyPeerGroupFactory)(nil)

// NewProxyPeerGroupFactory dials the proxy at proxyAddr and returns a factory
// that creates remote-backed peer groups.
func NewProxyPeerGroupFactory(proxyAddr string) (ClosablePeerGroupFactory, error) {
	conn, err := grpc.NewClient(proxyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}
	return &ProxyPeerGroupFactory{
		client: NewPeerGroupProxyClient(conn),
		conn:   conn,
	}, nil
}

func (f *ProxyPeerGroupFactory) Close() error {
	return f.conn.Close()
}

func (f *ProxyPeerGroupFactory) NewPeerGroup(
	configDigest [32]byte,
	peerIDs []string,
	bootstrappers []BootstrapperInfo,
) (PeerGroup, error) {
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := f.client.Connect(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to open peer group connection: %w", err)
	}

	pbBootstrappers := make([]*BootstrapperLocator, len(bootstrappers))
	for i, b := range bootstrappers {
		pbBootstrappers[i] = &BootstrapperLocator{
			PeerId: b.PeerID,
			Addrs:  b.Addrs,
		}
	}

	if err := stream.Send(&PeerGroupClientRequest{
		Message: &PeerGroupClientRequest_NewPeerGroup{
			NewPeerGroup: &NewPeerGroupRequest{
				ConfigDigest:    configDigest[:],
				PeerIds:         peerIDs,
				V2Bootstrappers: pbBootstrappers,
			},
		},
	}); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to send new peer group request: %w", err)
	}

	pg := &proxyPeerGroup{
		stream:  stream,
		cancel:  cancel,
		streams: map[string]*proxyStream{},
	}
	pg.wg.Add(1)
	go pg.receiveLoop()
	return pg, nil
}

// proxyPeerGroup implements PeerGroup over a single Connect stream.
type proxyPeerGroup struct {
	stream PeerGroupProxy_ConnectClient
	cancel context.CancelFunc

	// sendMu serializes writes on the gRPC stream (SendMessage/NewStream/Close
	// may be called concurrently).
	sendMu sync.Mutex

	mu      sync.Mutex
	streams map[string]*proxyStream
	nextID  uint64
	closed  bool

	wg sync.WaitGroup
}

var _ PeerGroup = (*proxyPeerGroup)(nil)

func (g *proxyPeerGroup) send(req *PeerGroupClientRequest) error {
	g.sendMu.Lock()
	defer g.sendMu.Unlock()
	return g.stream.Send(req)
}

func (g *proxyPeerGroup) receiveLoop() {
	defer g.wg.Done()
	for {
		msg, err := g.stream.Recv()
		if err != nil {
			return
		}
		recv, ok := msg.Message.(*PeerGroupServerMessage_StreamRecv)
		if !ok {
			continue
		}
		g.mu.Lock()
		st := g.streams[recv.StreamRecv.StreamId]
		g.mu.Unlock()
		if st != nil {
			st.deliver(recv.StreamRecv.Payload)
		}
	}
}

func (g *proxyPeerGroup) NewStream(remotePeerID string, args StreamArgs) (PeerGroupStream, error) {
	bufSize := args.IncomingBufferSize
	if bufSize <= 0 {
		bufSize = defaultRecvBufferSize
	}

	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return nil, fmt.Errorf("peer group is closed")
	}
	g.nextID++
	streamID := strconv.FormatUint(g.nextID, 10)
	st := &proxyStream{
		group:    g,
		streamID: streamID,
		recvChan: make(chan []byte, bufSize),
	}
	g.streams[streamID] = st
	g.mu.Unlock()

	if err := g.send(&PeerGroupClientRequest{
		Message: &PeerGroupClientRequest_NewStream{
			NewStream: &NewStreamRequest{
				StreamId:           streamID,
				RemotePeerId:       remotePeerID,
				StreamName:         args.StreamName,
				OutgoingBufferSize: int32(args.OutgoingBufferSize),
				IncomingBufferSize: int32(args.IncomingBufferSize),
				MaxMessageLength:   int32(args.MaxMessageLength),
				MessagesLimit:      &TokenBucketParams{Rate: args.MessagesLimit.Rate, Capacity: args.MessagesLimit.Capacity},
				BytesLimit:         &TokenBucketParams{Rate: args.BytesLimit.Rate, Capacity: args.BytesLimit.Capacity},
			},
		},
	}); err != nil {
		g.mu.Lock()
		delete(g.streams, streamID)
		g.mu.Unlock()
		return nil, fmt.Errorf("failed to send new stream request: %w", err)
	}
	return st, nil
}

func (g *proxyPeerGroup) Close() error {
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return nil
	}
	g.closed = true
	for id, st := range g.streams {
		st.closeRecv()
		delete(g.streams, id)
	}
	g.mu.Unlock()

	_ = g.stream.CloseSend()
	g.cancel()
	g.wg.Wait()
	return nil
}

// proxyStream implements PeerGroupStream over the parent group's connection.
type proxyStream struct {
	group    *proxyPeerGroup
	streamID string

	mu       sync.Mutex
	recvChan chan []byte
	closed   bool
}

var _ PeerGroupStream = (*proxyStream)(nil)

func (s *proxyStream) SendMessage(data []byte) {
	// Best-effort, matching libocr stream semantics: drop on transport error.
	_ = s.group.send(&PeerGroupClientRequest{
		Message: &PeerGroupClientRequest_StreamSend{
			StreamSend: &StreamSend{StreamId: s.streamID, Payload: data},
		},
	})
}

func (s *proxyStream) ReceiveMessages() <-chan []byte {
	return s.recvChan
}

func (s *proxyStream) Close() error {
	s.group.mu.Lock()
	delete(s.group.streams, s.streamID)
	s.group.mu.Unlock()

	_ = s.group.send(&PeerGroupClientRequest{
		Message: &PeerGroupClientRequest_CloseStream{
			CloseStream: &CloseStreamRequest{StreamId: s.streamID},
		},
	})
	s.closeRecv()
	return nil
}

// deliver hands a received payload to the stream's channel. It is a no-op once
// the stream is closed, so it can never send on a closed channel.
func (s *proxyStream) deliver(payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	select {
	case s.recvChan <- payload:
	default:
		// Buffer full: drop, matching the lenient best-effort delivery of the
		// underlying networking stack.
	}
}

func (s *proxyStream) closeRecv() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.recvChan)
}
