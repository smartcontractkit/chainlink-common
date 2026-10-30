package beholder

import (
	"context"
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

// SetClient sets the global Beholder Client. Call it only from the process
// entrypoint, so libraries that still read the global see the configured client.
func SetClient(client *Client) {
	globalClient.Store(client)
}

// GetClient returns the global Beholder Client. It is safe for concurrent use.
//
// Deprecated: accept a *Client, or the field you need, from the caller instead.
func GetClient() *Client {
	return globalClient.Load()
}

// GetLogger returns the global client's Logger.
//
// Deprecated: accept a *Client, or an otellog.Logger, from the caller instead.
func GetLogger() otellog.Logger {
	return GetClient().Logger
}

// GetTracer returns the global client's Tracer.
//
// Deprecated: accept a *Client, or an oteltrace.Tracer, from the caller instead.
func GetTracer() oteltrace.Tracer {
	return GetClient().Tracer
}

// GetMeter returns the global client's Meter.
//
// Deprecated: accept a *Client, or an otelmetric.Meter, from the caller instead.
func GetMeter() otelmetric.Meter {
	return GetClient().Meter
}

// GetEmitter returns the global client's Emitter.
//
// Deprecated: accept a *Client, or an Emitter, from the caller instead.
func GetEmitter() Emitter {
	return GetClient().Emitter
}

// GlobalEmitter returns an Emitter that delegates to the current global client's
// Emitter at call time. Close is a no-op; the global client owns its emitter.
//
// Deprecated: accept a *Client, or an Emitter, from the caller instead.
func GlobalEmitter() Emitter { return globalEmitter{} }

type globalEmitter struct{}

func (globalEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	return GetEmitter().Emit(ctx, body, attrKVs...)
}
func (globalEmitter) Close() error { return nil }

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
