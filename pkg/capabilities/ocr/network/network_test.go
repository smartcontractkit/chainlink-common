package network

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/libocr/commontypes"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/ocr/network/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestConnect(t *testing.T) {
	t.Parallel()
	conn := serve(t, newFakeHost())

	factory, err := Connect(logger.Test(t), conn, "12D3KooWPeer", Config{Version: VersionLatest})
	require.NoError(t, err)
	assert.Equal(t, "12D3KooWPeer", factory.PeerID())

	_, err = Connect(logger.Test(t), conn, "12D3KooWPeer", Config{Version: latestVersion + 1})
	require.ErrorContains(t, err, "unknown OCR network version")
}

func TestEndpoint(t *testing.T) {
	t.Parallel()
	host := newFakeHost()
	factory := connect(t, serve(t, host))

	digest := ocrtypes.ConfigDigest{2}
	limits := ocrtypes.BinaryNetworkEndpointLimits{MaxMessageLength: 10, MessagesRatePerOracle: 1.5}
	endpoint, err := factory.NewEndpoint(digest, []string{"a", "b"}, nil, 1, limits)
	require.NoError(t, err)
	require.NoError(t, endpoint.Start())

	hosted := host.endpoint(t, digest)
	assert.Equal(t, []string{"a", "b"}, hosted.peerIDs)
	assert.Equal(t, limits, hosted.limits)

	endpoint.SendTo([]byte("to-1"), 1)
	endpoint.Broadcast([]byte("all"))
	assert.Equal(t, sent{payload: "to-1", to: 1}, receiveWithin(t, hosted.sent))
	assert.Equal(t, sent{payload: "all", broadcast: true}, receiveWithin(t, hosted.sent))

	hosted.received <- commontypes.BinaryMessageWithSender{Msg: []byte("from-2"), Sender: 2}
	assert.Equal(t, commontypes.BinaryMessageWithSender{Msg: []byte("from-2"), Sender: 2}, receiveWithin(t, endpoint.Receive()))

	require.NoError(t, endpoint.Close())
	require.NoError(t, endpoint.Close())
	receiveWithin(t, hosted.closed)
}

// TestEndpoint_LargeMessages checks messages past gRPC's 4 MiB default cross both ways, as a leader's proposal can.
func TestEndpoint_LargeMessages(t *testing.T) {
	t.Parallel()
	host := newFakeHost()
	factory := connect(t, serve(t, host))

	digest := ocrtypes.ConfigDigest{6}
	endpoint, err := factory.NewEndpoint(digest, nil, nil, 1, ocrtypes.BinaryNetworkEndpointLimits{})
	require.NoError(t, err)
	require.NoError(t, endpoint.Start())
	t.Cleanup(func() { _ = endpoint.Close() })
	hosted := host.endpoint(t, digest)

	large := strings.Repeat("x", 5*1024*1024)
	endpoint.SendTo([]byte(large), 1)
	assert.Equal(t, sent{payload: large, to: 1}, receiveWithin(t, hosted.sent))

	hosted.received <- commontypes.BinaryMessageWithSender{Msg: []byte(large), Sender: 2}
	assert.Equal(t, commontypes.BinaryMessageWithSender{Msg: []byte(large), Sender: 2}, receiveWithin(t, endpoint.Receive()))
}

func TestEndpoint_HostErrorsAreReturnedByNewEndpoint(t *testing.T) {
	t.Parallel()
	host := newFakeHost()
	host.newEndpointErr = errors.New("already open")
	factory := connect(t, serve(t, host))

	_, err := factory.NewEndpoint(ocrtypes.ConfigDigest{3}, nil, nil, 1, ocrtypes.BinaryNetworkEndpointLimits{})
	require.ErrorContains(t, err, "already open")
}

func TestEndpoint_HostClosesEndpointWhenStreamBreaks(t *testing.T) {
	t.Parallel()
	host := newFakeHost()
	conn := serve(t, host)
	factory := connect(t, conn)

	digest := ocrtypes.ConfigDigest{4}
	_, err := factory.NewEndpoint(digest, nil, nil, 1, ocrtypes.BinaryNetworkEndpointLimits{})
	require.NoError(t, err)
	hosted := host.endpoint(t, digest)

	require.NoError(t, conn.Close())
	receiveWithin(t, hosted.closed)
}

func TestEndpoint_SendDropsOldestWhenFull(t *testing.T) {
	t.Parallel()
	e := &endpoint{lggr: logger.Test(t), outbound: make(chan *pb.EndpointRequest, 2)}

	e.Broadcast([]byte("1"))
	e.Broadcast([]byte("2"))
	e.Broadcast([]byte("3"))

	assert.Equal(t, []byte("2"), (<-e.outbound).GetBroadcast())
	assert.Equal(t, []byte("3"), (<-e.outbound).GetBroadcast())
}

func connect(t *testing.T, conn *grpc.ClientConn) ocrtypes.BinaryNetworkEndpointFactory {
	t.Helper()
	factory, err := Connect(logger.Test(t), conn, "12D3KooWPeer", Config{})
	require.NoError(t, err)
	return factory
}

func serve(t *testing.T, host *fakeHost) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer(grpc.MaxRecvMsgSize(MaxMessageSize))
	RegisterServer(s, logger.Test(t), func(context.Context) (ocrtypes.BinaryNetworkEndpointFactory, error) {
		return host, nil
	})
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func receiveWithin[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		require.FailNow(t, "timed out waiting to receive")
		var zero T
		return zero
	}
}

// fakeHost is a host whose endpoints record what they are given.
type fakeHost struct {
	peerID         string
	newEndpointErr error

	mu        sync.Mutex
	endpoints map[ocrtypes.ConfigDigest]*fakeEndpoint
	opened    chan ocrtypes.ConfigDigest
}

func newFakeHost() *fakeHost {
	return &fakeHost{
		peerID:    "12D3KooWPeer",
		endpoints: map[ocrtypes.ConfigDigest]*fakeEndpoint{},
		opened:    make(chan ocrtypes.ConfigDigest, 10),
	}
}

func (h *fakeHost) endpoint(t *testing.T, digest ocrtypes.ConfigDigest) *fakeEndpoint {
	t.Helper()
	require.Equal(t, digest, receiveWithin(t, h.opened))
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.endpoints[digest]
}

func (h *fakeHost) PeerID() string { return h.peerID }

func (h *fakeHost) NewEndpoint(digest ocrtypes.ConfigDigest, peerIDs []string, _ []commontypes.BootstrapperLocator,
	_ int, limits ocrtypes.BinaryNetworkEndpointLimits) (commontypes.BinaryNetworkEndpoint, error) {
	if h.newEndpointErr != nil {
		return nil, h.newEndpointErr
	}

	e := &fakeEndpoint{
		peerIDs:  peerIDs,
		limits:   limits,
		sent:     make(chan sent, 10),
		received: make(chan commontypes.BinaryMessageWithSender, 10),
		closed:   make(chan struct{}, 1),
	}
	h.mu.Lock()
	h.endpoints[digest] = e
	h.mu.Unlock()
	h.opened <- digest
	return e, nil
}

type sent struct {
	payload   string
	to        commontypes.OracleID
	broadcast bool
}

type fakeEndpoint struct {
	peerIDs  []string
	limits   ocrtypes.BinaryNetworkEndpointLimits
	sent     chan sent
	received chan commontypes.BinaryMessageWithSender
	closed   chan struct{}
}

func (e *fakeEndpoint) SendTo(payload []byte, to commontypes.OracleID) {
	e.sent <- sent{payload: string(payload), to: to}
}

func (e *fakeEndpoint) Broadcast(payload []byte) {
	e.sent <- sent{payload: string(payload), broadcast: true}
}

func (e *fakeEndpoint) Receive() <-chan commontypes.BinaryMessageWithSender { return e.received }

func (e *fakeEndpoint) Start() error { return nil }

func (e *fakeEndpoint) Close() error {
	e.closed <- struct{}{}
	return nil
}
