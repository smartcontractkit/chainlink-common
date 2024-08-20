package global

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/log"
	otellogglobal "go.opentelemetry.io/otel/log/global"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

// Pointer to the global Beholder Client
var globalClient = defaultClient()

// SetClient sets the global Beholder Client
func SetClient(client *beholder.OtelClient) {
	globalClient.Store(client)
}

// Returns the global Beholder Client
// Its thread-safe and can be used concurrently
func GetClient() *beholder.OtelClient {
	return globalClient.Load()
}

func Logger() otellog.Logger {
	return GetClient().Logger
}

func Tracer() oteltrace.Tracer {
	return GetClient().Tracer
}

func Meter() otelmetric.Meter {
	return GetClient().Meter
}

func Emitter() beholder.Emitter {
	return GetClient().Emitter
}

func Close() error {
	return GetClient().Close()
}

func SpanFromContext(ctx context.Context) oteltrace.Span {
	return oteltrace.SpanFromContext(ctx)
}

func defaultClient() *atomic.Pointer[beholder.OtelClient] {
	ptr := &atomic.Pointer[beholder.OtelClient]{}
	client := beholder.NewNoopClient()
	ptr.Store(&client)
	return ptr
}

func Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	return Emitter().Emit(ctx, body, attrKVs...)
}

func Bootstrap(cfg beholder.Config, errorHandler func(error)) error {
	// Initialize beholder client
	c, err := beholder.NewOtelClient(cfg, errorHandler)
	if err != nil {
		return err
	}
	// Set global client so it will be accessible from anywhere through beholder/global functions
	SetClient(&c)
	return nil
}

func NewConfig() beholder.Config {
	return beholder.DefaultConfig()
}

func NewMessage(body []byte, attrKVs ...any) beholder.Message {
	return beholder.NewMessage(body, attrKVs...)
}

// Sets the global OTel logger, tracer, meter providers from OtelClient
// Makes them accessible from anywhere in the code via global otel getters:
// - otellog.GetLoggerProvider()
// - otel.GetTracerProvider()
// - otel.GetTextMapPropagator()
// - otel.GetMeterProvider()
// Any package that relies on go.opentelemetry.io will be able to pick up configured global providers
// e.g [otelgrpc](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc#example-NewServerHandler)
func SetGlobalOtelProviders() {
	c := GetClient()
	// Logger
	otellogglobal.SetLoggerProvider(c.LoggerProvider)
	// Tracer
	otel.SetTracerProvider(c.TracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	// Meter
	otel.SetMeterProvider(c.MeterProvider)
}
