package beholder_test

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// publishCall is one batch RPC received by batchCapturingChipServer.
type publishCall struct {
	method string
	md     metadata.MD
	events []*cepb.CloudEvent
}

// batchCapturingChipServer records every PublishBatch / PublishBatchNoAck call and answers
// with one positional success result per event, as the real server does.
type batchCapturingChipServer struct {
	pb.UnimplementedChipIngressServer

	mu    sync.Mutex
	calls []publishCall
}

func (s *batchCapturingChipServer) record(ctx context.Context, method string, in *pb.CloudEventBatch) (*pb.PublishResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, publishCall{method: method, md: md, events: in.Events})
	return &pb.PublishResponse{Results: make([]*pb.PublishResult, len(in.Events))}, nil
}

func (s *batchCapturingChipServer) PublishBatch(ctx context.Context, in *pb.CloudEventBatch) (*pb.PublishResponse, error) {
	return s.record(ctx, "PublishBatch", in)
}

func (s *batchCapturingChipServer) PublishBatchNoAck(ctx context.Context, in *pb.CloudEventBatch) (*pb.PublishResponse, error) {
	return s.record(ctx, "PublishBatchNoAck", in)
}

// callsWithEventSource returns the calls that carried an event with the given CloudEvent source.
func (s *batchCapturingChipServer) callsWithEventSource(source string) []publishCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []publishCall
	for _, c := range s.calls {
		for _, ev := range c.events {
			if ev.Source == source {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

func startBatchCapturingChipServer(t *testing.T) (*batchCapturingChipServer, string) {
	t.Helper()
	lis, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	capture := &batchCapturingChipServer{}
	pb.RegisterChipIngressServer(srv, capture)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return capture, lis.Addr().String()
}

const (
	testAuthHeaderKey = "X-Beholder-Node-Auth-Token"
	testAuthToken     = "1:abc:2:def"
)

func domainTestConfig(t *testing.T, chipAddr string) beholder.Config {
	return beholder.Config{
		OtelExporterGRPCEndpoint:       "localhost:4317",
		ChipIngressEmitterEnabled:      true,
		ChipIngressEmitterGRPCEndpoint: chipAddr,
		ChipIngressInsecureConnection:  true,
		ChipIngressBatchEmitterEnabled: true,
		ChipIngressLogger:              logger.Test(t),
		ChipIngressSendInterval:        10 * time.Millisecond,
		AuthHeaders:                    map[string]string{testAuthHeaderKey: testAuthToken},
		ResourceAttributes: []attribute.KeyValue{
			attribute.String("csa_public_key", "abc123"),
		},
		ChipDomainEmittersEnabled: true,
	}
}

func newDomainTestClient(t *testing.T, cfg beholder.Config) *beholder.Client {
	t.Helper()
	client, err := beholder.NewGRPCClient(cfg, func(...otlploggrpc.Option) (sdklog.Exporter, error) {
		return &mockLogExporter{}, nil
	})
	require.NoError(t, err)
	require.NoError(t, client.Start(t.Context()))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// legacyStyleEvent builds an event the way the node's telemetry manager will: a normal
// CloudEvent carrying its own extensions, which beholder's attribute-based Emit cannot express.
func legacyStyleEvent(t *testing.T, source string) chipingress.CloudEvent {
	t.Helper()
	ev, err := chipingress.NewEvent(source, "pb.TelemetryWrapper", []byte("payload"), nil)
	require.NoError(t, err)
	ev.SetExtension("telemetrytype", "ocr2-median")
	ev.SetExtension("partitionkey", "0xcontract-abc123")
	return ev
}

func TestForDomain_EmitEventSendsNoAckWithDomainHeader(t *testing.T) {
	capture, addr := startBatchCapturingChipServer(t)
	client := newDomainTestClient(t, domainTestConfig(t, addr))

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)

	const source = "data-feeds.telemetry.ocr2-median"
	done := make(chan error, 1)
	require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, source), func(err error) { done <- err }))

	select {
	case err := <-done:
		require.NoError(t, err, "callback reports the per-event result")
	case <-time.After(5 * time.Second):
		t.Fatal("callback not invoked")
	}

	calls := capture.callsWithEventSource(source)
	require.Len(t, calls, 1)
	call := calls[0]

	assert.Equal(t, "PublishBatchNoAck", call.method, "domain emitters are always no-ack")
	assert.Equal(t, []string{"data-feeds"}, call.md.Get(beholder.DomainHeaderKey), "routing header carries the domain")
	assert.Equal(t, []string{testAuthToken}, call.md.Get(testAuthHeaderKey), "inherits the client's auth")
	for key, val := range chipingress.SanitizeMetadataHeaders(map[string]string{"csa_public_key": "abc123"}) {
		assert.Equal(t, []string{val}, call.md.Get(key), "inherits resource attribute header %q", key)
	}
	assert.Equal(t, []string{"true"}, call.md.Get("x-include-nop-info"), "domain emitters request NOP info")
	require.Len(t, call.md.Get("user-agent"), 1)
	assert.Contains(t, call.md.Get("user-agent")[0], "beholder_data-feeds", "per-domain client name")

	require.Len(t, call.events, 1)
	attrs := call.events[0].Attributes
	require.Contains(t, attrs, "telemetrytype", "caller-set extensions survive")
	assert.Equal(t, "ocr2-median", attrs["telemetrytype"].GetCeString())
	require.Contains(t, attrs, "partitionkey")
	assert.Equal(t, "0xcontract-abc123", attrs["partitionkey"].GetCeString())
}

func TestForDomain_EachDomainUsesItsOwnHeader(t *testing.T) {
	capture, addr := startBatchCapturingChipServer(t)
	client := newDomainTestClient(t, domainTestConfig(t, addr))

	ccip, err := client.ForDomain(beholder.DomainCCIP)
	require.NoError(t, err)
	require.NoError(t, ccip.EmitEvent(t.Context(), legacyStyleEvent(t, "ccip.telemetry.ocr3"), nil))

	require.Eventually(t, func() bool { return len(capture.callsWithEventSource("ccip.telemetry.ocr3")) == 1 },
		5*time.Second, 10*time.Millisecond)
	assert.Equal(t, []string{"ccip"}, capture.callsWithEventSource("ccip.telemetry.ocr3")[0].md.Get(beholder.DomainHeaderKey))
}

func TestForDomain_DefaultEmitterIsUnaffected(t *testing.T) {
	capture, addr := startBatchCapturingChipServer(t)
	client := newDomainTestClient(t, domainTestConfig(t, addr))

	require.NoError(t, client.Emitter.Emit(t.Context(), []byte("payload"),
		beholder.AttrKeyDomain, "platform-custom",
		beholder.AttrKeyEntity, "SomeEntity",
		beholder.AttrKeyDataSchema, "/schemas/ids/1001",
	))

	require.Eventually(t, func() bool { return len(capture.callsWithEventSource("platform-custom")) == 1 },
		5*time.Second, 10*time.Millisecond)
	call := capture.callsWithEventSource("platform-custom")[0]
	assert.Equal(t, "PublishBatch", call.method, "the default emitter keeps acked publishing")
	assert.Empty(t, call.md.Get(beholder.DomainHeaderKey), "the default emitter sends no routing header")
}

func TestForDomain_BuildsEveryDomainAndRejectsUnknownOnes(t *testing.T) {
	_, addr := startBatchCapturingChipServer(t)
	client := newDomainTestClient(t, domainTestConfig(t, addr))

	for _, d := range []beholder.Domain{beholder.DomainDataFeeds, beholder.DomainDataStreams, beholder.DomainCCIP} {
		_, err := client.ForDomain(d)
		require.NoError(t, err, "domain %q is built", d)
	}

	first, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)
	second, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)
	assert.Same(t, first, second, "ForDomain is a lookup, not a constructor")

	_, err = client.ForDomain(beholder.Domain("not-a-domain"))
	require.Error(t, err, "an unknown domain is an error, never a silent fallback")
}

func TestForDomain_CloseOnReturnedEmitterDoesNotStopIt(t *testing.T) {
	capture, addr := startBatchCapturingChipServer(t)
	client := newDomainTestClient(t, domainTestConfig(t, addr))

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)
	require.NoError(t, em.Close(), "the client owns the emitter's lifecycle")

	require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, "data-feeds.telemetry.ocr"), nil))
	require.Eventually(t, func() bool { return len(capture.callsWithEventSource("data-feeds.telemetry.ocr")) == 1 },
		5*time.Second, 10*time.Millisecond)
}

func TestForDomain_ClientCloseDrainsQueuedEvents(t *testing.T) {
	capture, addr := startBatchCapturingChipServer(t)
	cfg := domainTestConfig(t, addr)
	cfg.ChipIngressSendInterval = time.Hour // nothing flushes on its own

	client, err := beholder.NewGRPCClient(cfg, func(...otlploggrpc.Option) (sdklog.Exporter, error) {
		return &mockLogExporter{}, nil
	})
	require.NoError(t, err)
	require.NoError(t, client.Start(t.Context()))

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)
	require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, "data-feeds.telemetry.drain"), nil))

	requireNoDomainEmitterCloseErrors(t, client.Close())
	assert.Len(t, capture.callsWithEventSource("data-feeds.telemetry.drain"), 1, "queued events are published on close")
}

// requireNoDomainEmitterCloseErrors checks a Client.Close error for anything the domain emitters
// caused. Close also shuts down the OTel providers, which fail to flush to the test's absent
// collector, so those errors are expected and ignored.
func requireNoDomainEmitterCloseErrors(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	assert.NotContains(t, err.Error(), "client connection is closing",
		"domain emitter connections are closed exactly once")
}

func TestForDomain_ClientCloseClosesDomainConnectionsOnce(t *testing.T) {
	capture, addr := startBatchCapturingChipServer(t)
	client, err := beholder.NewGRPCClient(domainTestConfig(t, addr), func(...otlploggrpc.Option) (sdklog.Exporter, error) {
		return &mockLogExporter{}, nil
	})
	require.NoError(t, err)
	require.NoError(t, client.Start(t.Context()))

	// Use every domain so each connection is actually dialed before Close.
	for _, d := range []beholder.Domain{beholder.DomainDataFeeds, beholder.DomainDataStreams, beholder.DomainCCIP} {
		em, err := client.ForDomain(d)
		require.NoError(t, err)
		source := string(d) + ".telemetry.close"
		require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, source), nil))
		require.Eventually(t, func() bool { return len(capture.callsWithEventSource(source)) == 1 },
			5*time.Second, 10*time.Millisecond)
	}

	requireNoDomainEmitterCloseErrors(t, client.Close())
}

func TestForDomain_EmitEventRejectsInvalidEvent(t *testing.T) {
	_, addr := startBatchCapturingChipServer(t)
	client := newDomainTestClient(t, domainTestConfig(t, addr))

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)

	// In-domain source, so only CloudEvent validation (a missing id) can reject it.
	ev := legacyStyleEvent(t, "data-feeds.telemetry.invalid")
	ev.SetID("")

	called := false
	err = em.EmitEvent(t.Context(), ev, func(error) { called = true })
	require.ErrorContains(t, err, "invalid CloudEvent")
	assert.False(t, called, "no callback when EmitEvent returns an error")
}

func TestForDomain_EmitEventRejectsCancelledContext(t *testing.T) {
	_, addr := startBatchCapturingChipServer(t)
	client := newDomainTestClient(t, domainTestConfig(t, addr))

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	called := false
	err = em.EmitEvent(ctx, legacyStyleEvent(t, "data-feeds.telemetry.cancelled"), func(error) { called = true })
	require.ErrorIs(t, err, context.Canceled)
	assert.False(t, called, "no callback when EmitEvent returns an error")
}

// Domain emitters are always batch emitters, whatever ChipIngressBatchEmitterEnabled says.
func TestForDomain_WorksWithBatchEmitterDisabled(t *testing.T) {
	capture, addr := startBatchCapturingChipServer(t)
	cfg := domainTestConfig(t, addr)
	cfg.ChipIngressBatchEmitterEnabled = false
	client := newDomainTestClient(t, cfg)

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)
	require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, "data-feeds.telemetry.unbatched"), nil))

	require.Eventually(t, func() bool { return len(capture.callsWithEventSource("data-feeds.telemetry.unbatched")) == 1 },
		5*time.Second, 10*time.Millisecond)
	assert.Equal(t, "PublishBatchNoAck", capture.callsWithEventSource("data-feeds.telemetry.unbatched")[0].method)
}

func TestForDomain_RejectsEventsFromOtherDomains(t *testing.T) {
	capture, addr := startBatchCapturingChipServer(t)
	client := newDomainTestClient(t, domainTestConfig(t, addr))

	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)

	for _, tc := range []struct {
		source string
		ok     bool
	}{
		{"data-feeds", true},                       // the top-level domain itself
		{"data-feeds.telemetry.ocr2-median", true}, // a subdomain of it
		{"ccip.telemetry.ocr3", false},             // another domain
		{"data-feeds-extra.telemetry", false},      // shares a prefix, not a subdomain
		{"telemetry.data-feeds", false},            // contains it, not under it
		{"Data-Feeds.telemetry", false},            // the match is case-sensitive
	} {
		t.Run(tc.source, func(t *testing.T) {
			called := false
			err := em.EmitEvent(t.Context(), legacyStyleEvent(t, tc.source), func(error) { called = true })
			if tc.ok {
				require.NoError(t, err)
				require.Eventually(t, func() bool { return len(capture.callsWithEventSource(tc.source)) == 1 },
					5*time.Second, 10*time.Millisecond)
				return
			}
			require.Error(t, err, "event from another domain must not ride this domain's route")
			assert.False(t, called, "no callback when EmitEvent returns an error")
		})
	}

	t.Run("Emit is checked too", func(t *testing.T) {
		err := em.Emit(t.Context(), []byte("payload"),
			beholder.AttrKeyDomain, "ccip.telemetry.ocr3",
			beholder.AttrKeyEntity, "SomeEntity",
		)
		require.Error(t, err)
	})

	// Nothing from the rejected sources reached the server.
	for _, source := range []string{"ccip.telemetry.ocr3", "data-feeds-extra.telemetry", "telemetry.data-feeds", "Data-Feeds.telemetry"} {
		assert.Empty(t, capture.callsWithEventSource(source), "source %q", source)
	}
}

func TestNewGRPCClient_DomainEmittersRequireChipIngress(t *testing.T) {
	cfg := beholder.Config{OtelExporterGRPCEndpoint: "localhost:4317", ChipDomainEmittersEnabled: true}
	_, err := beholder.NewGRPCClient(cfg, func(...otlploggrpc.Option) (sdklog.Exporter, error) {
		return &mockLogExporter{}, nil
	})
	require.Error(t, err)
}

func TestForDomain_DisabledByDefault(t *testing.T) {
	_, addr := startBatchCapturingChipServer(t)
	cfg := domainTestConfig(t, addr)
	cfg.ChipDomainEmittersEnabled = false
	client := newDomainTestClient(t, cfg)

	_, err := client.ForDomain(beholder.DomainDataFeeds)
	require.Error(t, err, "chip ingress alone does not build domain emitters")
}

func TestNoopClient_ForDomain(t *testing.T) {
	_, err := beholder.NewNoopClient().ForDomain(beholder.DomainDataFeeds)
	require.ErrorIs(t, err, beholder.ErrDomainEmittersDisabled)
}

func TestForDomain_ErrorsAreDistinguishable(t *testing.T) {
	_, addr := startBatchCapturingChipServer(t)

	enabled := newDomainTestClient(t, domainTestConfig(t, addr))
	_, err := enabled.ForDomain(beholder.Domain("not-a-domain"))
	require.ErrorIs(t, err, beholder.ErrUnknownDomain)
	require.NotErrorIs(t, err, beholder.ErrDomainEmittersDisabled)

	cfg := domainTestConfig(t, addr)
	cfg.ChipDomainEmittersEnabled = false
	disabled := newDomainTestClient(t, cfg)
	_, err = disabled.ForDomain(beholder.DomainDataFeeds)
	require.ErrorIs(t, err, beholder.ErrDomainEmittersDisabled)
	require.NotErrorIs(t, err, beholder.ErrUnknownDomain)
}

func TestNewGRPCClient_DomainEmittersRequireLogger(t *testing.T) {
	cfg := domainTestConfig(t, "localhost:9999")
	cfg.ChipIngressLogger = nil
	cfg.ChipIngressBatchEmitterEnabled = false // the batch emitter's own logger check must not be what fails
	_, err := beholder.NewGRPCClient(cfg, func(...otlploggrpc.Option) (sdklog.Exporter, error) {
		return &mockLogExporter{}, nil
	})
	require.ErrorContains(t, err, "ChipIngressLogger is required when ChipDomainEmittersEnabled is true")
}

func TestNewHTTPClient_RejectsDomainEmitters(t *testing.T) {
	_, err := beholder.NewClient(beholder.Config{
		OtelExporterHTTPEndpoint:  "localhost:4318",
		ChipDomainEmittersEnabled: true,
	})
	require.ErrorContains(t, err, "ChipDomainEmittersEnabled is not supported by the HTTP client")
}
