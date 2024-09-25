package beholder

import (
	"sync/atomic"

	"go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/log"
	otellogglobal "go.opentelemetry.io/otel/log/global"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Pointer to the global Beholder Client
var globalClient atomic.Pointer[Client]

func init() { globalClient.Store(NewNoopClient()) }

// SetClient sets the global Beholder Client
func SetClient(client *Client) {
	globalClient.Store(client)
}

// Returns the global Beholder Client
// Its thread-safe and can be used concurrently
func GetClient() *Client {
	return globalClient.Load()
}

func GetLogger() otellog.Logger {
	return GetClient().Logger
}

func GetTracer() oteltrace.Tracer {
	return GetClient().Tracer
}

func GetMeter() otelmetric.Meter {
	return GetClient().Meter
}

func GetEmitter() Emitter {
	return GetClient().Emitter
}

// Sets global OTel logger, tracer, meter providers from Client.
// Makes them accessible from anywhere in the code via global otel getters.
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
