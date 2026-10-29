package beholder_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// hungChipServer accepts batch RPCs and never answers until the RPC is cancelled.
type hungChipServer struct {
	pb.UnimplementedChipIngressServer
	received atomic.Int64
}

func (s *hungChipServer) PublishBatchNoAck(ctx context.Context, _ *pb.CloudEventBatch) (*pb.PublishResponse, error) {
	s.received.Add(1)
	<-ctx.Done()
	return nil, ctx.Err()
}

// ackOnlyChipServer predates PublishBatchNoAck: it implements only PublishBatch.
type ackOnlyChipServer struct {
	pb.UnimplementedChipIngressServer
}

func (s *ackOnlyChipServer) PublishBatch(_ context.Context, in *pb.CloudEventBatch) (*pb.PublishResponse, error) {
	return &pb.PublishResponse{Results: make([]*pb.PublishResult, len(in.Events))}, nil
}

func startChipServer(t *testing.T, impl pb.ChipIngressServer) string {
	t.Helper()
	lis, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	pb.RegisterChipIngressServer(srv, impl)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return lis.Addr().String()
}

// closeBound bounds Client.Close in these tests. Close always spends ~10s flushing the OTel
// providers to the absent test collector; the domain emitters add at most their drain timeout
// (under 1s here), so anything near this bound is a real hang.
const closeBound = 20 * time.Second

// unusedAddr returns an address nothing listens on.
func unusedAddr(t *testing.T) string {
	t.Helper()
	lis, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	require.NoError(t, lis.Close())
	return addr
}

func newStartedDomainClient(t *testing.T, cfg beholder.Config) *beholder.Client {
	t.Helper()
	client, err := beholder.NewGRPCClient(cfg, func(...otlploggrpc.Option) (sdklog.Exporter, error) {
		return &mockLogExporter{}, nil
	})
	require.NoError(t, err)
	require.NoError(t, client.Start(t.Context()))
	return client
}

// callbackCounter records how many times each event's callback fired and with what error.
type callbackCounter struct {
	mu     sync.Mutex
	counts map[int]int
	errs   map[int]error
}

func newCallbackCounter() *callbackCounter {
	return &callbackCounter{counts: map[int]int{}, errs: map[int]error{}}
}

func (c *callbackCounter) callback(i int) func(error) {
	return func(err error) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.counts[i]++
		c.errs[i] = err
	}
}

func (c *callbackCounter) snapshot() (map[int]int, map[int]error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	counts := make(map[int]int, len(c.counts))
	errs := make(map[int]error, len(c.errs))
	for k, v := range c.counts {
		counts[k] = v
	}
	for k, v := range c.errs {
		errs[k] = v
	}
	return counts, errs
}

func TestDomainEmitter_ServerDown(t *testing.T) {
	cfg := domainTestConfig(t, unusedAddr(t))
	cfg.ChipIngressSendTimeout = 500 * time.Millisecond
	cfg.ChipIngressDrainTimeout = time.Second
	client := newStartedDomainClient(t, cfg)

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)

	done := make(chan error, 1)
	require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, "data-feeds.telemetry.down"), func(err error) { done <- err }))
	select {
	case err := <-done:
		require.Error(t, err, "an unreachable server fails the event")
	case <-time.After(5 * time.Second):
		t.Fatal("callback not invoked while the server is down")
	}

	closed := make(chan error, 1)
	go func() { closed <- client.Close() }()
	select {
	case err := <-closed:
		requireNoDomainEmitterCloseErrors(t, err)
	case <-time.After(closeBound):
		t.Fatal("Close hung while the server is down")
	}
}

func TestDomainEmitter_HungServerCloseFiresEachCallbackOnce(t *testing.T) {
	hung := &hungChipServer{}
	cfg := domainTestConfig(t, startChipServer(t, hung))
	cfg.ChipIngressSendTimeout = 30 * time.Second // the send outlives the drain
	cfg.ChipIngressDrainTimeout = 500 * time.Millisecond
	client := newStartedDomainClient(t, cfg)

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)

	const n = 5
	counter := newCallbackCounter()
	for i := range n {
		require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, fmt.Sprintf("data-feeds.telemetry.hung-%d", i)), counter.callback(i)))
	}
	require.Eventually(t, func() bool { return hung.received.Load() > 0 }, 5*time.Second, 10*time.Millisecond,
		"a batch is in flight before Close")

	closed := make(chan struct{})
	go func() { _ = client.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(closeBound):
		t.Fatal("Close did not return against a hung server")
	}

	// In-flight callbacks can fire just after Close returns (the connection is force-closed).
	require.Eventually(t, func() bool {
		counts, _ := counter.snapshot()
		return len(counts) == n
	}, 5*time.Second, 10*time.Millisecond, "every callback fires")
	time.Sleep(200 * time.Millisecond) // give any duplicate a chance to show up
	counts, errs := counter.snapshot()
	for i := range n {
		assert.Equal(t, 1, counts[i], "callback %d fires exactly once", i)
		assert.Error(t, errs[i], "event %d was never acknowledged", i)
	}
}

func TestDomainEmitter_BufferFull(t *testing.T) {
	cfg := domainTestConfig(t, startChipServer(t, &hungChipServer{}))
	cfg.ChipIngressBufferSize = 1
	cfg.ChipIngressMaxBatchSize = 1
	cfg.ChipIngressMaxConcurrentSends = 1
	cfg.ChipIngressSendTimeout = 30 * time.Second
	cfg.ChipIngressDrainTimeout = 200 * time.Millisecond
	client := newStartedDomainClient(t, cfg)

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)

	// With the only send slot hung, the buffer fills; the next event is dropped via its callback.
	var (
		pending sync.WaitGroup
		dropped atomic.Bool
	)
	for i := 0; i < 100 && !dropped.Load(); i++ {
		pending.Add(1)
		require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, fmt.Sprintf("data-feeds.telemetry.full-%d", i)), func(err error) {
			defer pending.Done()
			if errors.Is(err, batch.ErrMessageBufferFull) {
				dropped.Store(true)
			}
		}))
		time.Sleep(5 * time.Millisecond)
	}
	require.True(t, dropped.Load(), "an event is dropped with ErrMessageBufferFull")

	// The hung sends outlive the drain timeout, so their callbacks fire after Close returns. Wait
	// for every one before the test ends, so nothing logs through the finished test's logger.
	requireNoDomainEmitterCloseErrors(t, client.Close())
	waited := make(chan struct{})
	go func() { pending.Wait(); close(waited) }()
	select {
	case <-waited:
	case <-time.After(closeBound):
		t.Fatal("not every accepted event's callback fired")
	}
}

func TestDomainEmitter_ServerWithoutNoAck(t *testing.T) {
	client := newDomainTestClient(t, domainTestConfig(t, startChipServer(t, &ackOnlyChipServer{})))

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)

	done := make(chan error, 1)
	require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, "data-feeds.telemetry.old-server"), func(err error) { done <- err }))
	select {
	case err := <-done:
		require.Error(t, err, "no fallback to PublishBatch")
		assert.Equal(t, codes.Unimplemented.String(), batch.ErrorCodeFor(err))
	case <-time.After(5 * time.Second):
		t.Fatal("callback not invoked")
	}
}

func TestDomainEmitter_OutageLogsAreRateLimited(t *testing.T) {
	lggr, logs := logger.TestObserved(t, zapcore.ErrorLevel)
	cfg := domainTestConfig(t, startChipServer(t, &ackOnlyChipServer{}))
	cfg.ChipIngressLogger = lggr
	cfg.ChipIngressMaxBatchSize = 1 // one RPC, and so one failure, per event
	client := newDomainTestClient(t, cfg)

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)

	const n = 20
	var pending sync.WaitGroup
	for i := range n {
		pending.Add(1)
		require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, fmt.Sprintf("data-feeds.telemetry.outage-%d", i)),
			func(err error) {
				defer pending.Done()
				assert.Error(t, err)
			}))
	}
	pending.Wait()

	assert.Equal(t, 1, logs.FilterMessage("failed to emit to chip ingress").Len(),
		"every event failed, but only the first failure per error code is logged")
}

func TestDomainEmitter_ConcurrentEmitAndClose(t *testing.T) {
	_, addr := startBatchCapturingChipServer(t)
	client := newStartedDomainClient(t, domainTestConfig(t, addr))

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for g := range 8 {
		wg.Go(func() {
			for i := range 200 {
				// Errors are expected once Close has stopped the emitter.
				_ = em.EmitEvent(context.Background(), legacyStyleEvent(t, fmt.Sprintf("data-feeds.telemetry.race-%d-%d", g, i)), nil)
			}
		})
	}
	time.Sleep(20 * time.Millisecond)
	requireNoDomainEmitterCloseErrors(t, client.Close())
	wg.Wait()
}

func TestDomainEmitter_EmitAttributePath(t *testing.T) {
	capture, addr := startBatchCapturingChipServer(t)
	client := newDomainTestClient(t, domainTestConfig(t, addr))

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)
	require.NoError(t, em.Emit(t.Context(), []byte("payload"),
		beholder.AttrKeyDomain, "data-feeds.telemetry.emit",
		beholder.AttrKeyEntity, "SomeEntity",
	))

	require.Eventually(t, func() bool { return len(capture.callsWithEventSource("data-feeds.telemetry.emit")) == 1 },
		5*time.Second, 10*time.Millisecond)
	call := capture.callsWithEventSource("data-feeds.telemetry.emit")[0]
	assert.Equal(t, "PublishBatchNoAck", call.method)
	assert.Equal(t, []string{"data-feeds"}, call.md.Get(beholder.DomainHeaderKey))
}

func TestNewGRPCClient_DomainEmitterConstructionFailures(t *testing.T) {
	newClient := func(cfg beholder.Config) error {
		_, err := beholder.NewGRPCClient(cfg, func(...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		})
		return err
	}

	t.Run("bad endpoint fails while building domain emitters", func(t *testing.T) {
		err := newClient(domainTestConfig(t, "missing-port"))
		require.ErrorContains(t, err, `chip ingress client for domain "data-feeds"`)
	})
}

// A NewGRPCClient that fails after creating its chip ingress client must close it. The
// ChipIngressLogger check fails after the default chip client exists (domain emitters off, since
// they require the logger up front). Each leaked client keeps gRPC goroutines alive.
func TestNewGRPCClient_ConstructionFailureClosesChipClient(t *testing.T) {
	cfg := beholder.Config{
		OtelExporterGRPCEndpoint:       "localhost:4317",
		ChipIngressEmitterEnabled:      true,
		ChipIngressEmitterGRPCEndpoint: "localhost:9999",
		ChipIngressInsecureConnection:  true,
		ChipIngressBatchEmitterEnabled: true, // with a nil ChipIngressLogger
	}
	before := runtime.NumGoroutine()
	for range 3 {
		_, err := beholder.NewGRPCClient(cfg, func(...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		})
		require.ErrorContains(t, err, "ChipIngressLogger is required")
	}
	// Leaking the chip client costs 3 goroutines per failed call; allow some unrelated noise.
	require.Eventually(t, func() bool { return runtime.NumGoroutine()-before < 5 }, 5*time.Second, 50*time.Millisecond,
		"failed constructions leaked goroutines: before=%d now=%d", before, runtime.NumGoroutine())
}
