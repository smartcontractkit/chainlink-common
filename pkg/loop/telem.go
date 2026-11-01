package loop

import (
	"context"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/telemetry"
)

type GRPCOpts = loopnet.GRPCOpts

type OtelAttributes = beholder.OtelAttributes

type TracingConfig = telemetry.TracingConfig

// SetupTracing initializes open telemetry with the provided config.
// It sets the global trace provider and opens a connection to the configured collector.
func SetupTracing(config TracingConfig) error { return telemetry.SetupTracing(config) }

type GRPCOptsConfig struct {
	Registerer           prometheus.Registerer // leave nil to use default [prometheus.Registerer]
	ServerMaxRecvMsgSize int                   // [grpc.MaxRecvMsgSize]
}

func (c GRPCOptsConfig) New(lggr logger.Logger) GRPCOpts {
	if c.Registerer == nil {
		c.Registerer = prometheus.DefaultRegisterer
	}
	opts := GRPCOpts{DialOpts: dialOptions(c.Registerer), NewServer: newServerFn(lggr, c.Registerer)}
	if c.ServerMaxRecvMsgSize > 0 {
		newFn := opts.NewServer
		opts.NewServer = func(opts []grpc.ServerOption) *grpc.Server {
			opts = append(opts, grpc.MaxRecvMsgSize(c.ServerMaxRecvMsgSize))
			return newFn(opts)
		}
	}
	return opts
}

// NewGRPCOpts initializes open telemetry and returns GRPCOpts with telemetry interceptors.
// It is called from the host and each plugin - intended as there is bidirectional communication
// Deprecated: Use GRPCOptsConfig.New
func NewGRPCOpts(registerer prometheus.Registerer) GRPCOpts {
	return GRPCOptsConfig{Registerer: registerer}.New(logger.Nop())
}

var grpcpromBuckets = []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}

// dialOptions returns [grpc.DialOption]s to intercept and reports telemetry.
func dialOptions(r prometheus.Registerer) []grpc.DialOption {
	cm := grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(grpcprom.WithHistogramBuckets(grpcpromBuckets)),
	)
	r.MustRegister(cm)
	ctxExemplar := grpcprom.WithExemplarFromContext(exemplarFromContext)
	return []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		// Order matters e.g. tracing interceptor have to create span first for the later exemplars to work.
		grpc.WithChainUnaryInterceptor(
			contexts.CREUnaryInterceptor,
			cm.UnaryClientInterceptor(ctxExemplar),
		),
		grpc.WithChainStreamInterceptor(
			contexts.CREStreamInterceptor,
			cm.StreamClientInterceptor(ctxExemplar),
		),
	}
}

// newServerFn return a func for constructing [*grpc.Server]s that intercepts and reports telemetry.
func newServerFn(lggr logger.Logger, r prometheus.Registerer) func(opts []grpc.ServerOption) *grpc.Server {
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(grpcprom.WithHistogramBuckets(grpcpromBuckets)),
	)
	r.MustRegister(srvMetrics)
	ctxExemplar := grpcprom.WithExemplarFromContext(exemplarFromContext)
	creInterceptor := contexts.NewCREServerInterceptor(lggr)
	interceptors := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// Order matters e.g. tracing interceptor have to create span first for the later exemplars to work.
		grpc.ChainUnaryInterceptor(
			creInterceptor.UnaryServerInterceptor,
			srvMetrics.UnaryServerInterceptor(ctxExemplar),
		),
		grpc.ChainStreamInterceptor(
			creInterceptor.StreamServerInterceptor,
			srvMetrics.StreamServerInterceptor(ctxExemplar),
		),
	}
	return func(opts []grpc.ServerOption) *grpc.Server {
		s := grpc.NewServer(append(opts, interceptors...)...)
		srvMetrics.InitializeMetrics(s)
		return s
	}
}

func exemplarFromContext(ctx context.Context) prometheus.Labels {
	if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
		return prometheus.Labels{"traceID": span.TraceID().String()}
	}
	return nil
}
