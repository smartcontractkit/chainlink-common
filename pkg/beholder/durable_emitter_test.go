package beholder

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// withTestBeholderMeter swaps the global beholder client meter for t's lifetime (for metrics assertions).
func withTestBeholderMeter(t *testing.T) *sdkmetric.ManualReader {
	t.Helper()
	prev := GetClient()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	c := NewNoopClient()
	c.MeterProvider = mp
	c.Meter = mp.Meter(defaultPackageName)
	SetClient(c)
	t.Cleanup(func() {
		SetClient(prev)
		_ = mp.Shutdown(context.Background())
	})
	return reader
}

// testChipClient is a minimal chipingress.Client for tests.
type testChipClient struct {
	chipingress.NoopClient

	mu           sync.Mutex
	publishErr   error
	publishCount atomic.Int64
	batchCount   atomic.Int64
	publishedIDs []string
}

func (c *testChipClient) Publish(_ context.Context, ev *chipingress.CloudEventPb, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
	c.publishCount.Add(1)
	c.mu.Lock()
	if ev != nil {
		c.publishedIDs = append(c.publishedIDs, ev.Id)
	}
	err := c.publishErr
	c.mu.Unlock()
	return &chipingress.PublishResponse{}, err
}

// PublishBatch mirrors production semantics: respect publishErr and count as a
// separate RPC (batch path / tests that assert Publish only would miss it).
func (c *testChipClient) PublishBatch(_ context.Context, b *chipingress.CloudEventBatch, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
	c.batchCount.Add(1)
	c.mu.Lock()
	if b != nil {
		for _, ev := range b.Events {
			if ev != nil {
				c.publishedIDs = append(c.publishedIDs, ev.Id)
			}
		}
	}
	err := c.publishErr
	c.mu.Unlock()
	return &chipingress.PublishResponse{}, err
}

// totalChipRPCs is unary Publish + PublishBatch for assertions that do not care which path ran.
func (c *testChipClient) totalChipRPCs() int64 {
	return c.publishCount.Load() + c.batchCount.Load()
}

func (c *testChipClient) setPublishErr(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.publishErr = err
}

func (c *testChipClient) getPublishedIDs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.publishedIDs))
	copy(out, c.publishedIDs)
	return out
}

func testEmitAttrs() []any {
	return []any{"source", "test-source", "type", "test-type"}
}

func newTestDurableEmitter(t *testing.T, store DurableEventStore, client chipingress.Client, cfgOverride *DurableEmitterConfig) *DurableEmitter {
	t.Helper()
	cfg := DefaultDurableEmitterConfig()
	if cfgOverride != nil {
		cfg = *cfgOverride
	}
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)
	return em
}

func TestDurableEmitter_HooksImmediatePath(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	var pubCalls, delCalls atomic.Int32
	cfg := DefaultDurableEmitterConfig()
	cfg.Hooks = &DurableEmitterHooks{
		OnImmediatePublish: func(time.Duration, error) { pubCalls.Add(1) },
		OnImmediateDelete:  func(time.Duration, error) { delCalls.Add(1) },
	}
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, int32(1), pubCalls.Load())
	assert.Equal(t, int32(1), delCalls.Load())
}

func TestDurableEmitter_HooksPublishFailureSkipsDeleteHook(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("down"))
	var pubCalls, delCalls atomic.Int32
	cfg := DefaultDurableEmitterConfig()
	cfg.Hooks = &DurableEmitterHooks{
		OnImmediatePublish: func(time.Duration, error) { pubCalls.Add(1) },
		OnImmediateDelete:  func(time.Duration, error) { delCalls.Add(1) },
	}
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return pubCalls.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, int32(0), delCalls.Load())
}

func TestDurableEmitter_EmitPersistsAndPublishes(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	em := newTestDurableEmitter(t, store, client, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("hello"), testEmitAttrs()...)
	require.NoError(t, err)

	// Immediate async publish should fire and delete the record.
	require.Eventually(t, func() bool {
		return client.publishCount.Load() == 1
	}, 2*time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 2*time.Second, 10*time.Millisecond)
}

func TestDurableEmitter_EmitReturnSuccessEvenWhenPublishFails(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("connection refused"))

	em := newTestDurableEmitter(t, store, client, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("hello"), testEmitAttrs()...)
	require.NoError(t, err, "Emit must succeed once the DB insert succeeds")

	// Wait for the async publish attempt to complete.
	require.Eventually(t, func() bool {
		return client.publishCount.Load() == 1
	}, 2*time.Second, 10*time.Millisecond)

	// Event must remain in the store for retransmit.
	assert.Equal(t, 1, store.Len())
}

func TestDurableEmitter_RetransmitLoopDeliversFailedEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("connection refused"))

	cfg := DefaultDurableEmitterConfig()
	cfg.PublishBatchSize = 0 // this test keys off unary Publish; batch mode uses PublishBatch
	cfg.RetransmitInterval = 100 * time.Millisecond
	cfg.RetransmitAfter = 50 * time.Millisecond

	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("retry-me"), testEmitAttrs()...)
	require.NoError(t, err)

	// Wait until the async immediate path has run with the error and the row
	// is still pending (not a success race after we clear the error).
	require.Eventually(t, func() bool {
		return client.publishCount.Load() >= 1 && store.Len() == 1
	}, 2*time.Second, 5*time.Millisecond, "failed immediate publish should leave the row")

	client.setPublishErr(nil)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "retransmit loop should eventually deliver and delete the event")

	// At least: one failed immediate publish + one successful delivery (retransmit
	// may be unary Publish or PublishBatch when batching is enabled).
	assert.GreaterOrEqual(t, client.totalChipRPCs(), int64(2))
}

func TestDurableEmitter_RetransmitSerialDistinctCloudEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("immediate fail"))

	cfg := DefaultDurableEmitterConfig()
	cfg.PublishBatchSize = 0 // unary immediate Publish; serial retransmit
	cfg.RetransmitInterval = 100 * time.Millisecond
	cfg.RetransmitAfter = 50 * time.Millisecond

	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("first"), testEmitAttrs()...))
	require.NoError(t, em.Emit(ctx, []byte("second"), testEmitAttrs()...))

	require.Eventually(t, func() bool {
		return client.publishCount.Load() >= 2 && store.Len() == 2
	}, 2*time.Second, 5*time.Millisecond, "both failed immediate publishes should leave two rows")

	client.setPublishErr(nil)

	require.Eventually(t, func() bool { return store.Len() == 0 }, 5*time.Second, 50*time.Millisecond)

	ids := client.getPublishedIDs()
	require.GreaterOrEqual(t, len(ids), 4, "two failed attempts then two successful deliveries (IDs recorded)")
	require.GreaterOrEqual(t, client.totalChipRPCs(), int64(4))
	a, b := ids[len(ids)-2], ids[len(ids)-1]
	assert.NotEmpty(t, a)
	assert.NotEmpty(t, b)
	assert.NotEqualf(t, a, b, "retransmit must publish two distinct CloudEvents, not one pointer reused for every row")
}

func TestDurableEmitter_ExpiryLoopDeletesOldEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("always fail"))

	cfg := DefaultDurableEmitterConfig()
	cfg.ExpiryInterval = 100 * time.Millisecond
	cfg.EventTTL = 50 * time.Millisecond
	cfg.RetransmitInterval = 10 * time.Minute // effectively disable retransmit

	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("will-expire"), testEmitAttrs()...)
	require.NoError(t, err)
	assert.Equal(t, 1, store.Len())

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "expiry loop should purge the event")
}

func TestDurableEmitter_PersistSourceFilter_skipsStoreBestEffortPublish(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	cfg := DefaultDurableEmitterConfig()
	cfg.PersistCloudEventSources = []string{"only-this"}
	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return client.publishCount.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, 0, store.Len())
}

func TestDurableEmitter_PersistSourceFilter_persistsAllowedSource(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	cfg := DefaultDurableEmitterConfig()
	cfg.PersistCloudEventSources = []string{"test-source"}
	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return client.publishCount.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
}

func TestDurableEmitter_PersistSourceWildcardStarAllowsAll(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	cfg := DefaultDurableEmitterConfig()
	cfg.PersistCloudEventSources = []string{"*"}
	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
}

func TestDurableEmitter_RetransmitDropsDisallowedSource(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}

	ev, err := chipingress.NewEvent("unknown-domain", "t", []byte("b"), nil)
	require.NoError(t, err)
	evPb, err := chipingress.EventToProto(ev)
	require.NoError(t, err)
	payload, err := proto.Marshal(evPb)
	require.NoError(t, err)

	_, err = store.Insert(context.Background(), payload)
	require.NoError(t, err)

	cfg := DefaultDurableEmitterConfig()
	cfg.PersistCloudEventSources = []string{"test-source"}
	cfg.RetransmitInterval = 50 * time.Millisecond
	cfg.RetransmitAfter = 30 * time.Millisecond

	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.Eventually(t, func() bool {
		return store.Len() == 0 && client.publishCount.Load() == 0
	}, 3*time.Second, 20*time.Millisecond, "disallowed row should be deleted without Publish")
}

func TestDurableEmitter_EmitRejectsInvalidAttributes(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	em := newTestDurableEmitter(t, store, client, nil)

	err := em.Emit(context.Background(), []byte("no-attrs"))
	require.Error(t, err)
	assert.Equal(t, 0, store.Len(), "nothing should be persisted when attributes are invalid")
}

func TestDurableEmitter_MultipleEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	em := newTestDurableEmitter(t, store, client, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	const n = 50
	for i := 0; i < n; i++ {
		err := em.Emit(ctx, []byte("event"), testEmitAttrs()...)
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		return client.publishCount.Load() == int64(n)
	}, 5*time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 10*time.Millisecond, "all events should be delivered and deleted")
}

func TestNewDurableEmitter_ValidationErrors(t *testing.T) {
	log := logger.Test(t)
	cfg := DefaultDurableEmitterConfig()

	_, err := NewDurableEmitter(nil, &testChipClient{}, cfg, log)
	assert.ErrorContains(t, err, "store")

	_, err = NewDurableEmitter(NewMemDurableEventStore(), nil, cfg, log)
	assert.ErrorContains(t, err, "client")

	_, err = NewDurableEmitter(NewMemDurableEventStore(), &testChipClient{}, cfg, nil)
	assert.ErrorContains(t, err, "logger")
}

func TestDurableEmitter_MetricsRegistersEmitSuccess(t *testing.T) {
	reader := withTestBeholderMeter(t)

	store := NewMemDurableEventStore()
	client := &testChipClient{}
	cfg := DefaultDurableEmitterConfig()
	cfg.RetransmitInterval = time.Hour
	cfg.Metrics = &DurableEmitterMetricsConfig{PollInterval: 25 * time.Millisecond}

	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer func() { _ = em.Close() }()

	require.NoError(t, em.Emit(ctx, []byte("m"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))

	var found bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "beholder.durable_emitter.emit.success" {
				found = true
			}
		}
	}
	assert.True(t, found, "expected beholder.durable_emitter.emit.success in exported metrics")
}

// mockChipServer implements ChipIngressServer with controllable behaviour.
type mockChipServer struct {
	pb.UnimplementedChipIngressServer

	mu            sync.Mutex
	publishErr    error
	batchErr      error
	received      []*cepb.CloudEvent
	batchReceived [][]*cepb.CloudEvent
	publishCount  atomic.Int64
	batchCount    atomic.Int64
	publishDelay  time.Duration
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

func fastCfg() DurableEmitterConfig {
	return DurableEmitterConfig{
		// Retransmit must use unary Publish (not batch enqueue) in these tests.
		PublishBatchSize:    0,
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
	store := NewMemDurableEventStore()

	em, err := NewDurableEmitter(store, client, fastCfg(), logger.Test(t))
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
	_, addr := startMockServer(t, srv)
	client := newChipClient(t, addr)
	store := NewMemDurableEventStore()

	em, err := NewDurableEmitter(store, client, fastCfg(), logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("will-retry"), emitAttrs()...))

	require.Eventually(t, func() bool {
		return srv.publishCount.Load() >= 1 && store.Len() == 1
	}, 2*time.Second, 10*time.Millisecond, "failed immediate Publish should leave the row pending")

	// "Recover" the server.
	srv.setPublishErr(nil)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "retransmit loop should deliver after recovery")

	assert.GreaterOrEqual(t, srv.publishCount.Load(), int64(2),
		"one failed immediate Publish then one retransmit Publish")
	assert.Equal(t, int64(0), srv.batchCount.Load(), "retransmit should not use PublishBatch")
}

func TestIntegration_ServerDown_EventsSurvive(t *testing.T) {
	// Start server, then stop it to simulate total outage.
	srv := &mockChipServer{}
	gs, addr := startMockServer(t, srv)
	client := newChipClient(t, addr)
	store := NewMemDurableEventStore()

	cfg := fastCfg()
	cfg.PublishTimeout = 500 * time.Millisecond
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
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

	em2, err := NewDurableEmitter(store, client2, cfg, logger.Test(t))
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
	store := NewMemDurableEventStore()

	cfg := fastCfg()
	cfg.RetransmitBatchSize = 200
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
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
	_, addr := startMockServer(t, srv)
	client := newChipClient(t, addr)
	store := NewMemDurableEventStore()

	cfg := fastCfg()
	cfg.EventTTL = 100 * time.Millisecond
	cfg.ExpiryInterval = 100 * time.Millisecond
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
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

func TestIntegration_RetransmitUsesSerialPublish(t *testing.T) {
	// Immediate Publish fails; retransmit uses one Publish per queued row.
	srv := &mockChipServer{}
	srv.setPublishErr(status.Error(codes.Unavailable, "reject immediate"))
	_, addr := startMockServer(t, srv)
	client := newChipClient(t, addr)
	store := NewMemDurableEventStore()

	em, err := NewDurableEmitter(store, client, fastCfg(), logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	for i := 0; i < 5; i++ {
		require.NoError(t, em.Emit(ctx, []byte("retry-me"), emitAttrs()...))
	}

	// All five async immediate publishes must observe the error before we clear
	// it, or they succeed immediately and the retransmit loop has nothing to do.
	require.Eventually(t, func() bool {
		return srv.publishCount.Load() >= 5 && store.Len() == 5
	}, 3*time.Second, 10*time.Millisecond, "all five immediate Publish RPCs should have failed and left five rows")

	srv.setPublishErr(nil)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond,
		"retransmit should deliver each event with its own Publish RPC")

	assert.Equal(t, 0, srv.batchCallCount(), "retransmit should not call PublishBatch")
	assert.GreaterOrEqual(t, srv.publishCount.Load(), int64(10),
		"five failed immediate attempts plus five retransmit publishes")
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
	store := NewMemDurableEventStore()

	em, err := NewDurableEmitter(store, client, fastCfg(), logger.Test(t))
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
