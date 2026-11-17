//go:build loadtest

package loadtest

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

// Window is a half-open time range [Start, End) measured from FakeServer.MarkStart.
type Window struct {
	Start, End time.Duration
}

// ServerProfile configures the fake chip-ingress behaviour. All fields may be changed while the
// server is running with SetProfile.
type ServerProfile struct {
	// Latency is the fixed response delay (the simulated chip-ingress RTT, excluding loopback).
	Latency time.Duration
	// Jitter adds a uniform random delay in [0, Jitter) on top of Latency.
	Jitter time.Duration
	// ErrorRate is the fraction of batches (0..1) answered with a gRPC error (Unavailable or
	// Internal, chosen at random) instead of results.
	ErrorRate float64
	// Outages are windows (relative to MarkStart) during which every batch is rejected
	// immediately with codes.Unavailable, like a refused connection or a 503 from the LB.
	Outages []Window
}

// ServerDomainStats are the server-side counters for one domain (from the chainlink-domain header).
type ServerDomainStats struct {
	Batches        int64 `json:"batches"`
	EventsReceived int64 `json:"events_received"` // arrived at the server, whatever the outcome
	EventsAccepted int64 `json:"events_accepted"` // in batches answered with success results
	EventsRejected int64 `json:"events_rejected"` // in batches answered with an error (outage, error rate, cancel)
	Bytes          int64 `json:"bytes"`           // sum of proto.Size of received batch requests
	MaxBatchEvents int64 `json:"max_batch_events"`
	MaxBatchBytes  int64 `json:"max_batch_bytes"`
	NoAckBatches   int64 `json:"noack_batches"`
	AckBatches     int64 `json:"ack_batches"`
	// SourceMismatch counts events whose CloudEvent source is not the header domain or a
	// subdomain of it (a routing bug if non-zero).
	SourceMismatch int64 `json:"source_mismatch"`
	// BatchSizeHist buckets batches by event count.
	BatchSizeHist map[string]int64 `json:"batch_size_hist"`
}

var batchSizeBuckets = []struct {
	label string
	max   int
}{{"<=10", 10}, {"<=100", 100}, {"<=500", 500}, {"<=1000", 1000}, {">1000", 1 << 30}}

type domainCounters struct {
	batches, received, accepted, rejected, bytes atomic.Int64
	maxBatchEvents, maxBatchBytes                atomic.Int64
	noAck, ack, sourceMismatch                   atomic.Int64
	hist                                         [5]atomic.Int64
}

func storeMax(a *atomic.Int64, v int64) {
	for {
		cur := a.Load()
		if v <= cur || a.CompareAndSwap(cur, v) {
			return
		}
	}
}

// FakeServer is a real gRPC chip-ingress server on loopback that implements PublishBatch and
// PublishBatchNoAck (and Ping) with configurable latency, errors and outages. It is safe for
// concurrent use.
type FakeServer struct {
	pb.UnimplementedChipIngressServer

	srv  *grpc.Server
	addr string

	profile atomic.Pointer[ServerProfile]
	startNs atomic.Int64 // unix nanos of MarkStart; 0 until marked

	domains sync.Map // string -> *domainCounters
}

// NewFakeServer listens on 127.0.0.1:0 (insecure) and serves until Close.
func NewFakeServer(p ServerProfile) (*FakeServer, error) {
	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("fake chip-ingress listen: %w", err)
	}
	s := &FakeServer{addr: lis.Addr().String()}
	s.SetProfile(p)
	s.MarkStart()
	// Receive limit above the client's 10 MiB request cap so a batch the client would send is
	// never refused by the fake for size; the client must enforce the cap itself.
	s.srv = grpc.NewServer(
		grpc.MaxRecvMsgSize(64<<20),
		// The beholder client also owns OTLP metric and trace exporters. Point them at this server
		// and swallow their exports, so Close does not stall flushing to a collector that is not there.
		grpc.UnknownServiceHandler(discardUnknown),
	)
	pb.RegisterChipIngressServer(s.srv, s)
	go func() { _ = s.srv.Serve(lis) }()
	return s, nil
}

// discardUnknown answers any RPC the fake does not implement (the OTLP metric and trace exports)
// with an empty message, which is a valid empty OTLP response.
func discardUnknown(_ any, stream grpc.ServerStream) error {
	var in emptypb.Empty
	if err := stream.RecvMsg(&in); err != nil {
		return err
	}
	return stream.SendMsg(&emptypb.Empty{})
}

// Addr is the host:port the server listens on.
func (s *FakeServer) Addr() string { return s.addr }

// SetProfile replaces the behaviour profile; Outages are relative to the last MarkStart.
func (s *FakeServer) SetProfile(p ServerProfile) {
	cp := p
	cp.Outages = append([]Window(nil), p.Outages...)
	s.profile.Store(&cp)
}

// MarkStart sets the origin for outage windows to now.
func (s *FakeServer) MarkStart() { s.startNs.Store(time.Now().UnixNano()) }

// Close stops the server immediately, failing in-flight RPCs.
func (s *FakeServer) Close() { s.srv.Stop() }

// Snapshot returns a copy of the per-domain counters.
func (s *FakeServer) Snapshot() map[string]ServerDomainStats {
	out := map[string]ServerDomainStats{}
	s.domains.Range(func(k, v any) bool {
		c := v.(*domainCounters)
		st := ServerDomainStats{
			Batches:        c.batches.Load(),
			EventsReceived: c.received.Load(),
			EventsAccepted: c.accepted.Load(),
			EventsRejected: c.rejected.Load(),
			Bytes:          c.bytes.Load(),
			MaxBatchEvents: c.maxBatchEvents.Load(),
			MaxBatchBytes:  c.maxBatchBytes.Load(),
			NoAckBatches:   c.noAck.Load(),
			AckBatches:     c.ack.Load(),
			SourceMismatch: c.sourceMismatch.Load(),
			BatchSizeHist:  map[string]int64{},
		}
		for i, b := range batchSizeBuckets {
			st.BatchSizeHist[b.label] = c.hist[i].Load()
		}
		out[k.(string)] = st
		return true
	})
	return out
}

func (s *FakeServer) counters(domain string) *domainCounters {
	if v, ok := s.domains.Load(domain); ok {
		return v.(*domainCounters)
	}
	v, _ := s.domains.LoadOrStore(domain, &domainCounters{})
	return v.(*domainCounters)
}

// Ping answers like chip-ingress.
func (s *FakeServer) Ping(context.Context, *pb.EmptyRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Message: "pong"}, nil
}

// PublishBatch implements the acked RPC.
func (s *FakeServer) PublishBatch(ctx context.Context, in *pb.CloudEventBatch) (*pb.PublishResponse, error) {
	return s.handle(ctx, in, false)
}

// PublishBatchNoAck implements the no-ack RPC used by domain emitters. The fake answers it
// exactly like PublishBatch: validated and "handed to the producer" with no persistence.
func (s *FakeServer) PublishBatchNoAck(ctx context.Context, in *pb.CloudEventBatch) (*pb.PublishResponse, error) {
	return s.handle(ctx, in, true)
}

func (s *FakeServer) handle(ctx context.Context, in *pb.CloudEventBatch, noAck bool) (*pb.PublishResponse, error) {
	domain := "unknown"
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if v := md.Get(beholder.DomainHeaderKey); len(v) > 0 {
			domain = v[0]
		}
	}
	c := s.counters(domain)
	n := int64(len(in.Events))
	size := int64(proto.Size(in))

	c.batches.Add(1)
	c.received.Add(n)
	c.bytes.Add(size)
	storeMax(&c.maxBatchEvents, n)
	storeMax(&c.maxBatchBytes, size)
	if noAck {
		c.noAck.Add(1)
	} else {
		c.ack.Add(1)
	}
	for i, b := range batchSizeBuckets {
		if int(n) <= b.max {
			c.hist[i].Add(1)
			break
		}
	}
	var mismatch int64
	for _, ev := range in.Events {
		if ev.Source != domain && !strings.HasPrefix(ev.Source, domain+".") {
			mismatch++
		}
	}
	if mismatch > 0 {
		c.sourceMismatch.Add(mismatch)
	}

	reject := func(code codes.Code, msg string) (*pb.PublishResponse, error) {
		c.rejected.Add(n)
		return nil, status.Error(code, msg)
	}

	p := s.profile.Load()
	if s.inOutage(p) {
		return reject(codes.Unavailable, "fake chip-ingress: outage")
	}
	if p.ErrorRate > 0 && rand.Float64() < p.ErrorRate {
		if rand.IntN(2) == 0 {
			return reject(codes.Unavailable, "fake chip-ingress: injected unavailable")
		}
		return reject(codes.Internal, "fake chip-ingress: injected internal error")
	}

	if d := p.Latency; d > 0 || p.Jitter > 0 {
		if p.Jitter > 0 {
			d += time.Duration(rand.Int64N(int64(p.Jitter)))
		}
		t := time.NewTimer(d)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			c.rejected.Add(n)
			return nil, status.FromContextError(ctx.Err()).Err()
		}
	}

	results := make([]*pb.PublishResult, len(in.Events))
	for i, ev := range in.Events {
		results[i] = &pb.PublishResult{EventId: ev.Id}
	}
	c.accepted.Add(n)
	return &pb.PublishResponse{Results: results}, nil
}

func (s *FakeServer) inOutage(p *ServerProfile) bool {
	if len(p.Outages) == 0 {
		return false
	}
	el := time.Duration(time.Now().UnixNano() - s.startNs.Load())
	for _, w := range p.Outages {
		if el >= w.Start && el < w.End {
			return true
		}
	}
	return false
}
