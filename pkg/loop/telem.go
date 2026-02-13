package loop

import (
	"context"
	"net"
	"os"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/config/build"
	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
)

type GRPCOpts = loopnet.GRPCOpts

type OtelAttributes = beholder.OtelAttributes

type TracingConfig struct {
	// NodeAttributes are the attributes to attach to traces.
	NodeAttributes OtelAttributes

	// Enables tracing; requires a collector to be provided
	Enabled bool

	// Collector is the address of the OTEL collector to send traces to.
	CollectorTarget string

	// SamplingRatio is the ratio of traces to sample. 1.0 means sample all traces.
	SamplingRatio float64

	// TLSCertPath is the path to the TLS certificate to use when connecting to the collector.
	TLSCertPath string

	// OnDialError is called when the dialer fails, providing an opportunity to log.
	OnDialError func(error)

	// Auth
	AuthHeaders map[string]string
}

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

// SetupTracing initializes open telemetry with the provided config.
// It sets the global trace provider and opens a connection to the configured collector.
func SetupTracing(config TracingConfig) error {
	if !config.Enabled {
		return nil
	}

	traceExporter, err := config.NewSpanExporter()
	if err != nil {
		return err
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		config.Attributes()...,
	)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(
			sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(config.SamplingRatio),
			),
		),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return nil
}

func (config TracingConfig) Attributes() []attribute.KeyValue {
	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String(build.Program),
		semconv.ProcessPIDKey.Int(os.Getpid()),
		semconv.ServiceVersionKey.String(build.Version),
	}

	for k, v := range config.NodeAttributes {
		attributes = append(attributes, attribute.String(k, v))
	}
	return attributes
}

func (config TracingConfig) NewSpanExporter() (sdktrace.SpanExporter, error) {
	ctx := context.Background()

	var creds credentials.TransportCredentials
	if config.TLSCertPath != "" {
		var err error
		creds, err = credentials.NewClientTLSFromFile(config.TLSCertPath, "")
		if err != nil {
			return nil, err
		}
	} else {
		creds = insecure.NewCredentials()
	}

	//nolint:staticcheck
	conn, err := grpc.DialContext(ctx, config.CollectorTarget,
		// Note the potential use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(creds),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			conn, err2 := net.Dial("tcp", s)
			if err2 != nil {
				config.OnDialError(err2)
			}
			return conn, err2
		}))
	if err != nil {
		return nil, err
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithGRPCConn(conn),
		otlptracegrpc.WithHeaders(config.AuthHeaders),
	)
	if err != nil {
		return nil, err
	}
	return traceExporter, nil
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
