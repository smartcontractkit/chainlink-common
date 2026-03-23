package beholder_test

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// mockChipServer implements ChipIngressServer with controllable behaviour.
type mockChipServer struct {
	pb.UnimplementedChipIngressServer

	mu              sync.Mutex
	publishErr      error
	batchErr        error
	received        []*cepb.CloudEvent
	batchReceived   [][]*cepb.CloudEvent
	publishCount    atomic.Int64
	batchCount      atomic.Int64
	publishDelay    time.Duration
}

func (s *mockChipServer) Publish(_ context.Context, in *cepb.CloudEvent) (*pb.PublishResponse, error) {
	if s.publishDelay > 0 {
		time.Sleep(s.publishDelay)
	}
	s.publishCount.Add(1)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.publishErr != nil {
		return nil, s.publishErr
	}
	s.received = append(s.received, in)
	return &pb.PublishResponse{}, nil
}

func (s *mockChipServer) PublishBatch(_ context.Context, in *pb.CloudEventBatch) (*pb.PublishResponse, error) {
	s.batchCount.Add(1)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.batchErr != nil {
		return nil, s.batchErr
	}
	s.batchReceived = append(s.batchReceived, in.Events)
	s.received = append(s.received, in.Events...)
	return &pb.PublishResponse{}, nil
}

func (s *mockChipServer) Ping(context.Context, *pb.EmptyRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Message: "pong"}, nil
}

func (s *mockChipServer) setPublishErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.publishErr = err
}

func (s *mockChipServer) setBatchErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batchErr = err
}

func (s *mockChipServer) receivedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.received)
}

func (s *mockChipServer) batchCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.batchReceived)
}

// startMockServer starts a gRPC server on a random port and returns the
// server, address, and a cleanup function.
func startMockServer(t *testing.T, srv *mockChipServer) (*grpc.Server, string) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	gs := grpc.NewServer()
	pb.RegisterChipIngressServer(gs, srv)

	go func() {
		if err := gs.Serve(lis); err != nil {
			// Ignore errors from server being stopped during cleanup.
		}
	}()

	t.Cleanup(func() { gs.GracefulStop() })
	return gs, lis.Addr().String()
}

func newChipClient(t *testing.T, addr string) chipingress.Client {
	t.Helper()
	c, err := chipingress.NewClient(addr, chipingress.WithInsecureConnection())
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func emitAttrs() []any {
	return []any{"source", "test-domain", "type", "test-entity"}
}

func fastCfg() beholder.DurableEmitterConfig {
	return beholder.DurableEmitterConfig{
		RetransmitInterval:  100 * time.Millisecond,
		RetransmitAfter:     50 * time.Millisecond,
		RetransmitBatchSize: 50,
		ExpiryInterval:      200 * time.Millisecond,
		EventTTL:            500 * time.Millisecond,
		PublishTimeout:      2 * time.Second,
	}
}

// ---------- Test cases ----------

func TestIntegration_HappyPath(t *testing.T) {
	srv := &mockChipServer{}
	_, addr := startMockServer(t, srv)
	client := newChipClient(t, addr)
	store := beholder.NewMemDurableEventStore()

	em, err := beholder.NewDurableEmitter(store, client, fastCfg(), logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("billing-record-1"), emitAttrs()...))
	require.NoError(t, em.Emit(ctx, []byte("billing-record-2"), emitAttrs()...))

	require.Eventually(t, func() bool {
		return srv.receivedCount() == 2
	}, 3*time.Second, 10*time.Millisecond, "server should receive both events")

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 3*time.Second, 10*time.Millisecond, "store should be empty after delivery")
}

func TestIntegration_ServerUnavailable_RetransmitRecovers(t *testing.T) {
	// Start with server returning UNAVAILABLE.
	srv := &mockChipServer{}
	srv.setPublishErr(status.Error(codes.Unavailable, "chip down"))
	srv.setBatchErr(status.Error(codes.Unavailable, "chip down"))
	_, addr := startMockServer(t, srv)
	client := newChipClient(t, addr)
	store := beholder.NewMemDurableEventStore()

	em, err := beholder.NewDurableEmitter(store, client, fastCfg(), logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("will-retry"), emitAttrs()...))

	// Event should be in the store, not delivered.
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, 1, store.Len(), "event persists while server is unavailable")

	// "Recover" the server.
	srv.setPublishErr(nil)
	srv.setBatchErr(nil)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "retransmit loop should deliver after recovery")

	assert.GreaterOrEqual(t, srv.batchCount.Load(), int64(1),
		"retransmit should use PublishBatch")
}

func TestIntegration_ServerDown_EventsSurvive(t *testing.T) {
	// Start server, then stop it to simulate total outage.
	srv := &mockChipServer{}
	gs, addr := startMockServer(t, srv)
	client := newChipClient(t, addr)
	store := beholder.NewMemDurableEventStore()

	cfg := fastCfg()
	cfg.PublishTimeout = 500 * time.Millisecond
	em, err := beholder.NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)

	// Stop the gRPC server entirely.
	gs.Stop()
	time.Sleep(100 * time.Millisecond)

	// Emit while server is down — Emit() itself must succeed (DB insert works).
	require.NoError(t, em.Emit(ctx, []byte("server-is-down"), emitAttrs()...))
	assert.Equal(t, 1, store.Len(), "event should be persisted even with server down")

	// Stop the emitter to simulate a "node shutdown".
	em.Close()

	// Bring up a new server on the same address.
	srv2 := &mockChipServer{}
	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	gs2 := grpc.NewServer()
	pb.RegisterChipIngressServer(gs2, srv2)
	go func() { _ = gs2.Serve(lis) }()
	t.Cleanup(func() { gs2.GracefulStop() })

	// Create a new client and DurableEmitter re-using the same store
	// (simulating node restart with Postgres).
	client2, err := chipingress.NewClient(addr, chipingress.WithInsecureConnection())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client2.Close() })

	em2, err := beholder.NewDurableEmitter(store, client2, cfg, logger.Test(t))
	require.NoError(t, err)
	em2.Start(ctx)
	defer em2.Close()

	require.Eventually(t, func() bool {
		return srv2.receivedCount() == 1
	}, 5*time.Second, 50*time.Millisecond, "new emitter should retransmit the surviving event")

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "store should be empty after retransmit")
}

func TestIntegration_HighThroughput(t *testing.T) {
	srv := &mockChipServer{}
	_, addr := startMockServer(t, srv)
	client := newChipClient(t, addr)
	store := beholder.NewMemDurableEventStore()

	cfg := fastCfg()
	cfg.RetransmitBatchSize = 200
	em, err := beholder.NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	const n = 500
	for i := 0; i < n; i++ {
		require.NoError(t, em.Emit(ctx, []byte("event"), emitAttrs()...))
	}

	require.Eventually(t, func() bool {
		return srv.receivedCount() >= n
	}, 10*time.Second, 50*time.Millisecond, "all %d events should be received", n)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 10*time.Second, 50*time.Millisecond, "store should drain completely")
}

func TestIntegration_EventExpiry(t *testing.T) {
	// Server always rejects — events can never be delivered.
	srv := &mockChipServer{}
	srv.setPublishErr(status.Error(codes.Internal, "permanent failure"))
	srv.setBatchErr(status.Error(codes.Internal, "permanent failure"))
	_, addr := startMockServer(t, srv)
	client := newChipClient(t, addr)
	store := beholder.NewMemDurableEventStore()

	cfg := fastCfg()
	cfg.EventTTL = 100 * time.Millisecond
	cfg.ExpiryInterval = 100 * time.Millisecond
	em, err := beholder.NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("will-expire"), emitAttrs()...))
	assert.Equal(t, 1, store.Len())

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond,
		"expiry loop should purge undeliverable events after TTL")
}

func TestIntegration_RetransmitUsesBatch(t *testing.T) {
	// Immediate publishes fail, only batch succeeds.
	srv := &mockChipServer{}
	srv.setPublishErr(status.Error(codes.Unavailable, "reject single"))
	_, addr := startMockServer(t, srv)
	client := newChipClient(t, addr)
	store := beholder.NewMemDurableEventStore()

	em, err := beholder.NewDurableEmitter(store, client, fastCfg(), logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	for i := 0; i < 5; i++ {
		require.NoError(t, em.Emit(ctx, []byte("batch-me"), emitAttrs()...))
	}

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond,
		"retransmit via PublishBatch should deliver all events")

	assert.GreaterOrEqual(t, srv.batchCallCount(), 1,
		"at least one PublishBatch call should have been made")
}

// TestIntegration_GRPCConnection verifies the emitter works over a real gRPC
// connection with proper proto serialization round-trip.
func TestIntegration_GRPCConnection(t *testing.T) {
	srv := &mockChipServer{}
	_, addr := startMockServer(t, srv)

	// Use a raw gRPC dial to prove we're going over the wire.
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	// Ping to verify connectivity.
	grpcClient := pb.NewChipIngressClient(conn)
	pong, err := grpcClient.Ping(context.Background(), &pb.EmptyRequest{})
	require.NoError(t, err)
	assert.Equal(t, "pong", pong.Message)

	// Now use the chipingress.Client wrapper with DurableEmitter.
	client := newChipClient(t, addr)
	store := beholder.NewMemDurableEventStore()

	em, err := beholder.NewDurableEmitter(store, client, fastCfg(), logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	payload := []byte("proto-round-trip-test")
	require.NoError(t, em.Emit(ctx, payload, emitAttrs()...))

	require.Eventually(t, func() bool {
		return srv.receivedCount() == 1
	}, 3*time.Second, 10*time.Millisecond)

	// Verify the CloudEvent arrived with correct source/type.
	srv.mu.Lock()
	received := srv.received[0]
	srv.mu.Unlock()

	assert.Equal(t, "test-domain", received.Source)
	assert.Equal(t, "test-entity", received.Type)
}
