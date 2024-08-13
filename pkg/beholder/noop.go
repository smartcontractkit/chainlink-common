package beholder

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otellognoop "go.opentelemetry.io/otel/log/noop"
	otelmetricnoop "go.opentelemetry.io/otel/metric/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltracenoop "go.opentelemetry.io/otel/trace/noop"
)

// Default client to fallback when is is not initialized properly
func NewNoopClient() OtelClient {
	cfg := DefaultConfig()
	// Logger
	loggerProvider := otellognoop.NewLoggerProvider()
	logger := loggerProvider.Logger(cfg.PackageName)
	// Tracer
	tracerProvider := oteltracenoop.NewTracerProvider()
	tracer := tracerProvider.Tracer(cfg.PackageName)

	// Meter
	meterProvider := otelmetricnoop.NewMeterProvider()
	meter := meterProvider.Meter(cfg.PackageName)

	// MessageEmitter
	messageEmitter := noopMessageEmitter{}

	client := OtelClient{cfg, logger, tracer, meter, messageEmitter, loggerProvider, tracerProvider, meterProvider, noopOnClose}

	return client
}

// NewStdoutClient creates a new Client with stdout exporters
// Use for testing and debugging
// Also this client is used as a noop client when otel exporter is not initialized properly
func NewStdoutClient() OtelClient {
	cfg := DefaultConfig()
	// Logger
	loggerExporter, _ := stdoutlog.New(stdoutlog.WithoutTimestamps()) // stdoutlog.New() never returns an error
	loggerProvider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(loggerExporter)))
	logger := loggerProvider.Logger(cfg.PackageName)
	setOtelErrorHandler(func(err error) {
		fmt.Printf("OTel error %s", err)
	})

	// Tracer
	traceExporter, _ := stdouttrace.New() // stdouttrace.New() never returns an error
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(
		sdktrace.NewSimpleSpanProcessor(traceExporter),
	))
	tracer := tracerProvider.Tracer(cfg.PackageName)

	// Meter
	metricExporter, _ := stdoutmetric.New() // stdoutmetric.New() never returns an error
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				metricExporter,
				sdkmetric.WithInterval(time.Second), // Default is 10s
			)),
	)
	meter := meterProvider.Meter(cfg.PackageName)

	// MessageEmitter
	emitter := messageEmitter{loggerExporter, logger}

	client := OtelClient{cfg, logger, tracer, meter, emitter, loggerProvider, tracerProvider, meterProvider, noopOnClose}

	return client
}

type noopMessageEmitter struct{}

func (noopMessageEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	return nil
}
func (noopMessageEmitter) EmitMessage(ctx context.Context, message Message) error {
	return nil
}

func noopOnClose() error {
	return nil
}
