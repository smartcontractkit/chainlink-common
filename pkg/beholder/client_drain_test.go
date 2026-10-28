package beholder_test

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	colmetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// otlpCollector is an in-process OTLP/gRPC collector that records the metric data points and
// spans it receives, so a test can see exactly what the client's providers exported.
type otlpCollector struct {
	colmetrics.UnimplementedMetricsServiceServer

	mu sync.Mutex
	// metricAttrs holds the attributes of every received metric data point.
	metricAttrs [][]*commonpb.KeyValue
	// spanAttrs holds the attributes of every received span.
	spanAttrs [][]*commonpb.KeyValue
}

func (c *otlpCollector) Export(_ context.Context, req *colmetrics.ExportMetricsServiceRequest) (*colmetrics.ExportMetricsServiceResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, rm := range req.GetResourceMetrics() {
		for _, sm := range rm.GetScopeMetrics() {
			for _, m := range sm.GetMetrics() {
				c.metricAttrs = append(c.metricAttrs, dataPointAttrs(m)...)
			}
		}
	}
	return &colmetrics.ExportMetricsServiceResponse{}, nil
}

// traceExporter adapts the collector to coltrace.TraceServiceServer: both services name their
// method Export, so they cannot share one receiver.
type traceExporter struct {
	coltrace.UnimplementedTraceServiceServer
	c *otlpCollector
}

func (t traceExporter) Export(_ context.Context, req *coltrace.ExportTraceServiceRequest) (*coltrace.ExportTraceServiceResponse, error) {
	t.c.mu.Lock()
	defer t.c.mu.Unlock()
	for _, rs := range req.GetResourceSpans() {
		for _, ss := range rs.GetScopeSpans() {
			for _, sp := range ss.GetSpans() {
				t.c.spanAttrs = append(t.c.spanAttrs, sp.GetAttributes())
			}
		}
	}
	return &coltrace.ExportTraceServiceResponse{}, nil
}

func dataPointAttrs(m *metricpb.Metric) (out [][]*commonpb.KeyValue) {
	switch d := m.GetData().(type) {
	case *metricpb.Metric_Sum:
		for _, dp := range d.Sum.GetDataPoints() {
			out = append(out, dp.GetAttributes())
		}
	case *metricpb.Metric_Gauge:
		for _, dp := range d.Gauge.GetDataPoints() {
			out = append(out, dp.GetAttributes())
		}
	case *metricpb.Metric_Histogram:
		for _, dp := range d.Histogram.GetDataPoints() {
			out = append(out, dp.GetAttributes())
		}
	case *metricpb.Metric_ExponentialHistogram:
		for _, dp := range d.ExponentialHistogram.GetDataPoints() {
			out = append(out, dp.GetAttributes())
		}
	}
	return out
}

func hasAttr(attrs []*commonpb.KeyValue, key, value string) bool {
	for _, kv := range attrs {
		if kv.GetKey() == key && kv.GetValue().GetStringValue() == value {
			return true
		}
	}
	return false
}

func countWithAttr(all [][]*commonpb.KeyValue, key, value string) (n int) {
	for _, attrs := range all {
		if hasAttr(attrs, key, value) {
			n++
		}
	}
	return n
}

func startOTLPCollector(t *testing.T) (*otlpCollector, string) {
	t.Helper()
	lis, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	c := &otlpCollector{}
	colmetrics.RegisterMetricsServiceServer(srv, c)
	coltrace.RegisterTraceServiceServer(srv, traceExporter{c: c})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return c, lis.Addr().String()
}

// TestClient_Close_DrainTelemetryReachesCollector proves that the chip ingress emitters drain
// before the client's OTel providers shut down: the otelgrpc metrics and spans of the RPCs sent
// during the drain are recorded on those providers, so they only reach the collector if the
// providers are still live when the drain runs.
func TestClient_Close_DrainTelemetryReachesCollector(t *testing.T) {
	capture, chipAddr := startBatchCapturingChipServer(t)
	collector, collectorAddr := startOTLPCollector(t)

	cfg := domainTestConfig(t, chipAddr)
	cfg.OtelExporterGRPCEndpoint = collectorAddr
	cfg.MetricCompressor = "none"
	cfg.TraceCompressor = "none"
	// The only metric and span exports are the providers' final flushes at shutdown.
	cfg.MetricReaderInterval = time.Hour
	cfg.TraceSampleRatio = 1
	cfg.TraceBatchTimeout = time.Hour
	// Nothing is sent before Close, so every chip ingress RPC happens during the drain.
	cfg.ChipIngressSendInterval = time.Hour

	client, err := beholder.NewGRPCClient(cfg, func(...otlploggrpc.Option) (sdklog.Exporter, error) {
		return &mockLogExporter{}, nil
	})
	require.NoError(t, err)
	require.NoError(t, client.Start(t.Context()))

	const (
		source     = "data-feeds.telemetry.drain-telemetry"
		clientName = "beholder_data-feeds"
	)
	em, err := client.ForDomain(beholder.DomainDataFeeds)
	require.NoError(t, err)
	require.NoError(t, em.EmitEvent(t.Context(), legacyStyleEvent(t, source), nil))
	require.Empty(t, capture.callsWithEventSource(source), "the event is only queued before Close")

	// Close may still report the double close of a domain's gRPC connection in
	// ChipIngressBatchEmitterService.stop, so only the lifecycle error this change could
	// introduce is asserted: the engine closing an emitter the client already drained.
	assert.NotErrorIs(t, client.Close(), services.ErrAlreadyStopped)

	require.Len(t, capture.callsWithEventSource(source), 1, "the queued event is published during Close")

	collector.mu.Lock()
	defer collector.mu.Unlock()
	assert.Positive(t, countWithAttr(collector.metricAttrs, "client_name", clientName),
		"otelgrpc metrics recorded while draining must be exported by the meter provider's final flush")
	assert.Positive(t, countWithAttr(collector.spanAttrs, "client_name", clientName),
		"otelgrpc spans recorded while draining must be exported by the tracer provider's final flush")
}
