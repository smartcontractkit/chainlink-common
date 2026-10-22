package telemetry

import (
	"context"
	"net"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/config/build"
)

type TracingConfig struct {
	// NodeAttributes are the attributes to attach to traces.
	NodeAttributes beholder.OtelAttributes

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
